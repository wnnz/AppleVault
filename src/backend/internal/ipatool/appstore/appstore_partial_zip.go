package appstore

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	apphttp "AppleVault/src/backend/internal/ipatool/http"
	"howett.net/plist"
)

const (
	zipMaxTrailerBytes         int64 = zipDirectoryEndLen + math.MaxUint16
	maxPrefetchCentralDirBytes int64 = 64 << 20
	maxPrefetchEntryBytes      int64 = 16 << 20
	localHeaderExtraSlackBytes int64 = 1024
	fallbackRangeBlockBytes    int64 = 256 << 10
	maxCachedByteSpans               = 16
)

var infoPlistReleaseDateKeys = []string{
	"releaseDate",
	"ReleaseDate",
}

var infoPlistDisplayVersionKeys = []string{
	"CFBundleShortVersionString",
	"bundleShortVersionString",
}

type versionMetadata struct {
	DisplayVersion string
	ReleaseDate    time.Time
}

type cachedByteSpan struct {
	offset int64
	data   []byte
}

type httpRangeReaderAt struct {
	ctx    context.Context
	client apphttp.Client[interface{}]
	url    string
	size   int64
	mu     sync.Mutex
	spans  []cachedByteSpan
}

func newHTTPRangeReaderAt(ctx context.Context, client apphttp.Client[interface{}], url string) (*httpRangeReaderAt, int64, error) {
	if url == "" {
		return nil, 0, errors.New("url is empty")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	if reader, size, ok, err := trySuffixRangeReaderAt(ctx, client, url); err != nil {
		return nil, 0, err
	} else if ok {
		return reader, size, nil
	}

	size, err := remoteFileSize(ctx, client, url)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read remote file size: %w", err)
	}

	reader := &httpRangeReaderAt{
		ctx:    ctx,
		client: client,
		url:    url,
		size:   size,
	}

	if size > 0 {
		tailStart := max(int64(0), size-zipMaxTrailerBytes)
		_ = reader.ensureRange(tailStart, size)
	}

	return reader, size, nil
}

func newIdentityRangeRequest(ctx context.Context, client apphttp.Client[interface{}], url, rangeHeader string) (*http.Request, error) {
	req, err := client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Range", rangeHeader)

	return req, nil
}

// trySuffixRangeReaderAt fetches the archive size and ZIP trailer together in one request.
func trySuffixRangeReaderAt(ctx context.Context, client apphttp.Client[interface{}], url string) (*httpRangeReaderAt, int64, bool, error) {
	req, err := newIdentityRangeRequest(ctx, client, url, fmt.Sprintf("bytes=-%d", zipMaxTrailerBytes))
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to read remote file size: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to read remote file size: request failed: %w", err)
	}

	if res != nil && res.Body != nil {
		defer func() {
			_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64))
			_ = res.Body.Close()
		}()
	}

	if res.StatusCode == http.StatusRequestedRangeNotSatisfiable || res.StatusCode == http.StatusBadRequest {
		return nil, 0, false, nil
	}

	if res.StatusCode != http.StatusPartialContent {
		return nil, 0, false, fmt.Errorf("failed to read remote file size: expected partial content response, got status %d", res.StatusCode)
	}

	start, end, size, err := parseContentRangeBounds(res.Header.Get("Content-Range"))
	if err != nil || size <= 0 || end != size-1 || (end-start+1) > zipMaxTrailerBytes {
		return nil, 0, false, nil
	}

	tail := make([]byte, int(end-start+1))
	if _, err := io.ReadFull(res.Body, tail); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, 0, false, ctxErr
		}

		return nil, 0, false, nil
	}

	return &httpRangeReaderAt{
		ctx:    ctx,
		client: client,
		url:    url,
		size:   size,
		spans:  []cachedByteSpan{{offset: start, data: tail}},
	}, size, true, nil
}

func remoteFileSize(ctx context.Context, client apphttp.Client[interface{}], url string) (int64, error) {
	req, err := newIdentityRangeRequest(ctx, client, url, "bytes=0-0")
	if err != nil {
		return 0, err
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	if res != nil && res.Body != nil {
		defer func() {
			_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64))
			_ = res.Body.Close()
		}()
	}

	if res.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("expected partial content response, got status %d", res.StatusCode)
	}

	size, err := parseContentRangeSize(res.Header.Get("Content-Range"))
	if err != nil {
		return 0, err
	}

	return size, nil
}

func parseContentRangeBounds(header string) (start, end, size int64, err error) {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return 0, 0, 0, errors.New("content range is empty")
	}

	if len(trimmed) < 6 || !strings.EqualFold(trimmed[:6], "bytes ") {
		return 0, 0, 0, fmt.Errorf("invalid content range: %s", header)
	}

	rest := strings.TrimSpace(trimmed[6:])
	slash := strings.LastIndex(rest, "/")
	if slash <= 0 || slash == len(rest)-1 {
		return 0, 0, 0, fmt.Errorf("invalid content range: %s", header)
	}

	sizeText := strings.TrimSpace(rest[slash+1:])
	if sizeText == "*" {
		return 0, 0, 0, fmt.Errorf("invalid content range size: %s", sizeText)
	}

	size, err = strconv.ParseInt(sizeText, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse content range size: %w", err)
	}

	boundsText := strings.TrimSpace(rest[:slash])
	dash := strings.Index(boundsText, "-")
	if dash <= 0 || dash == len(boundsText)-1 {
		return 0, 0, 0, fmt.Errorf("invalid content range bounds: %s", header)
	}

	start, err = strconv.ParseInt(strings.TrimSpace(boundsText[:dash]), 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse content range start: %w", err)
	}

	end, err = strconv.ParseInt(strings.TrimSpace(boundsText[dash+1:]), 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse content range end: %w", err)
	}

	if start < 0 || end < start || end >= size {
		return 0, 0, 0, fmt.Errorf("out of bounds content range: %s", header)
	}

	return start, end, size, nil
}

func parseContentRangeSize(header string) (int64, error) {
	if header == "" {
		return 0, errors.New("content range is empty")
	}

	slash := strings.LastIndex(header, "/")
	if slash == -1 || slash == len(header)-1 {
		return 0, fmt.Errorf("invalid content range: %s", header)
	}

	sizeText := header[slash+1:]
	if sizeText == "*" {
		return 0, fmt.Errorf("invalid content range size: %s", sizeText)
	}

	size, err := strconv.ParseInt(sizeText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse content range size: %w", err)
	}

	return size, nil
}

func (r *httpRangeReaderAt) missingSubrangeLocked(start, end int64) (int64, int64) {
	if start < 0 {
		start = 0
	}

	if end > r.size {
		end = r.size
	}

	for _, s := range r.spans {
		sEnd := s.offset + int64(len(s.data))
		if s.offset <= start && sEnd > start {
			start = sEnd
		}

		if start >= end {
			return 0, 0
		}
	}

	for i := len(r.spans) - 1; i >= 0; i-- {
		s := r.spans[i]
		sEnd := s.offset + int64(len(s.data))

		if s.offset < end && sEnd >= end {
			end = s.offset
		}

		if start >= end {
			return 0, 0
		}
	}

	return start, end
}

func (r *httpRangeReaderAt) addSpanLocked(offset int64, data []byte) {
	if len(data) == 0 {
		return
	}

	end := offset + int64(len(data))
	merged := make([]cachedByteSpan, 0, len(r.spans)+1)
	inserted := false

	for _, s := range r.spans {
		sEnd := s.offset + int64(len(s.data))
		if sEnd < offset {
			merged = append(merged, s)
			continue
		}

		if end < s.offset {
			if !inserted {
				merged = append(merged, cachedByteSpan{offset: offset, data: data})
				inserted = true
			}

			merged = append(merged, s)
			continue
		}

		newStart := min(offset, s.offset)
		newEnd := max(end, sEnd)
		combined := make([]byte, newEnd-newStart)
		copy(combined[s.offset-newStart:], s.data)
		copy(combined[offset-newStart:], data)
		offset = newStart
		data = combined
		end = newEnd
	}

	if !inserted {
		merged = append(merged, cachedByteSpan{offset: offset, data: data})
	}

	for len(merged) > maxCachedByteSpans {
		evictIdx := -1
		for i := range merged {
			if merged[i].offset == offset {
				continue
			}

			if evictIdx == -1 || len(merged[i].data) < len(merged[evictIdx].data) {
				evictIdx = i
			}
		}

		if evictIdx == -1 {
			break
		}

		merged = append(merged[:evictIdx], merged[evictIdx+1:]...)
	}

	r.spans = merged
}

func (r *httpRangeReaderAt) readCachedLocked(dst []byte, off int64) (int, bool) {
	need := int64(len(dst))
	for _, s := range r.spans {
		sEnd := s.offset + int64(len(s.data))
		if s.offset <= off && off+need <= sEnd {
			start := int(off - s.offset)
			n := copy(dst, s.data[start:start+len(dst)])

			return n, n == len(dst)
		}
	}

	return 0, false
}

func (r *httpRangeReaderAt) fetchRange(start, end int64) ([]byte, error) {
	if end <= start {
		return nil, nil
	}

	if err := r.ctx.Err(); err != nil {
		return nil, err
	}

	req, err := newIdentityRangeRequest(r.ctx, r.client, r.url, fmt.Sprintf("bytes=%d-%d", start, end-1))
	if err != nil {
		return nil, err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if res != nil && res.Body != nil {
		defer func() {
			_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64))
			_ = res.Body.Close()
		}()
	}

	if res.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("expected partial content response, got status %d", res.StatusCode)
	}

	buf := make([]byte, end-start)
	if _, err := io.ReadFull(res.Body, buf); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return buf, nil
}

func (r *httpRangeReaderAt) ensureRange(start, end int64) error {
	r.mu.Lock()
	missStart, missEnd := r.missingSubrangeLocked(start, end)
	r.mu.Unlock()

	if missStart >= missEnd {
		return nil
	}

	data, err := r.fetchRange(missStart, missEnd)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.addSpanLocked(missStart, data)
	r.mu.Unlock()

	return nil
}

// prefetchMainInfoPlist fetches the full Central Directory and Payload/*.app/Info.plist
// in coalesced Range requests so zip.NewReader and Info.plist decompression read from RAM.
func (r *httpRangeReaderAt) prefetchMainInfoPlist() error {
	directory, err := readZIPDirectoryEnd(r, r.size)
	if err != nil {
		return err
	}

	if directory.size > uint64(directory.endOffset) || directory.offset > math.MaxInt64 {
		return zip.ErrFormat
	}

	base := directory.endOffset - int64(directory.size) - int64(directory.offset)
	if base > 0 {
		var signature [4]byte
		if _, err := r.ReadAt(signature[:], int64(directory.offset)); err == nil && binary.LittleEndian.Uint32(signature[:]) == zipCentralHeaderSignature {
			base = 0
		}
	}

	if base < 0 || base > r.size-int64(directory.offset) {
		return zip.ErrFormat
	}

	cdStart := base + int64(directory.offset)
	cdEnd := min(r.size, max(directory.endOffset, cdStart+int64(directory.size)))

	cdLen := cdEnd - cdStart
	if cdLen <= 0 || cdLen > maxPrefetchCentralDirBytes {
		return nil
	}

	if err := r.ensureRange(cdStart, cdEnd); err != nil {
		return err
	}

	cdBufLen := min(cdEnd-cdStart, int64(directory.size))
	if cdBufLen <= 0 {
		return nil
	}

	cdBuf := make([]byte, int(cdBufLen))
	if _, err := r.ReadAt(cdBuf, cdStart); err != nil {
		return err
	}

	compSize, entryOffset, nameLen, found := findMainInfoPlistCentralEntry(cdBuf)
	if !found || entryOffset > uint64(math.MaxInt64-base) {
		return nil
	}

	localOffset := base + int64(entryOffset)
	if localOffset < 0 || localOffset >= r.size || compSize > uint64(maxPrefetchEntryBytes) {
		return nil
	}

	windowLen := int64(zipLocalHeaderLen) + int64(nameLen) + localHeaderExtraSlackBytes + int64(compSize) + 24
	if windowLen <= 0 || windowLen > maxPrefetchEntryBytes {
		return nil
	}

	return r.ensureRange(localOffset, min(r.size, localOffset+windowLen))
}

func findMainInfoPlistCentralEntry(cd []byte) (compSize, localOffset uint64, nameLen int, found bool) {
	pos := 0
	for pos+zipCentralHeaderLen <= len(cd) {
		if binary.LittleEndian.Uint32(cd[pos:]) != zipCentralHeaderSignature {
			break
		}

		nameLen = int(binary.LittleEndian.Uint16(cd[pos+28:]))
		extraLen := int(binary.LittleEndian.Uint16(cd[pos+30:]))
		commentLen := int(binary.LittleEndian.Uint16(cd[pos+32:]))

		metaStart := pos + zipCentralHeaderLen
		nextPos := metaStart + nameLen + extraLen + commentLen
		if nextPos > len(cd) {
			break
		}

		name := string(cd[metaStart : metaStart+nameLen])
		if isMainAppInfoPlist(name) {
			extra := cd[metaStart+nameLen : metaStart+nameLen+extraLen]

			compSize, localOffset, err := zipCentralEntryMeta(cd[pos:metaStart], extra)
			if err != nil {
				return 0, 0, 0, false
			}

			return compSize, localOffset, nameLen, true
		}

		pos = nextPos
	}

	return 0, 0, 0, false
}

func zipCentralEntryMeta(central, extra []byte) (compSize, localOffset uint64, err error) {
	offset32 := binary.LittleEndian.Uint32(central[42:])
	comp32 := binary.LittleEndian.Uint32(central[20:])
	uncomp32 := binary.LittleEndian.Uint32(central[24:])

	if comp32 != math.MaxUint32 && offset32 != math.MaxUint32 {
		return uint64(comp32), uint64(offset32), nil
	}

	var zip64 []byte

	rem := extra
	for len(rem) >= 4 {
		tag := binary.LittleEndian.Uint16(rem)

		length := int(binary.LittleEndian.Uint16(rem[2:]))
		if length > len(rem)-4 {
			return 0, 0, zip.ErrFormat
		}

		if tag == zip64ExtraID {
			zip64 = rem[4 : 4+length]
			break
		}

		rem = rem[4+length:]
	}

	pos := 0
	if uncomp32 == math.MaxUint32 {
		if len(zip64) < pos+8 {
			return 0, 0, zip.ErrFormat
		}

		pos += 8
	}

	if comp32 == math.MaxUint32 {
		if len(zip64) < pos+8 {
			return 0, 0, zip.ErrFormat
		}

		compSize = binary.LittleEndian.Uint64(zip64[pos:])
		pos += 8
	} else {
		compSize = uint64(comp32)
	}

	if offset32 == math.MaxUint32 {
		if len(zip64) < pos+8 {
			return 0, 0, zip.ErrFormat
		}

		localOffset = binary.LittleEndian.Uint64(zip64[pos:])
	} else {
		localOffset = uint64(offset32)
	}

	return compSize, localOffset, nil
}

func (r *httpRangeReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if off < 0 {
		return 0, errors.New("offset is negative")
	}

	if off >= r.size {
		return 0, io.EOF
	}

	readLength := int(min(int64(len(p)), r.size-off))
	wantEnd := off + int64(readLength)

	r.mu.Lock()
	if n, ok := r.readCachedLocked(p[:readLength], off); ok {
		r.mu.Unlock()

		if off+int64(len(p)) > r.size {
			return n, io.EOF
		}

		return n, nil
	}
	r.mu.Unlock()

	fetchEnd := min(r.size, max(wantEnd, off+fallbackRangeBlockBytes))
	if err := r.ensureRange(off, fetchEnd); err != nil {
		return 0, err
	}

	r.mu.Lock()
	n, ok := r.readCachedLocked(p[:readLength], off)
	r.mu.Unlock()

	if !ok {
		return n, io.ErrUnexpectedEOF
	}

	if off+int64(len(p)) > r.size {
		return n, io.EOF
	}

	return n, nil
}

func (t *appstore) readVersionMetadataFromIPA(ctx context.Context, url string) (versionMetadata, int64, error) {
	if ctx == nil {
		ctx = t.requestContext()
	}

	reader, size, err := newHTTPRangeReaderAt(ctx, t.httpClient, url)
	if err != nil {
		return versionMetadata{}, 0, err
	}

	if ctx.Err() != nil {
		return versionMetadata{}, 0, ctx.Err()
	}

	_ = reader.prefetchMainInfoPlist()

	if ctx.Err() != nil {
		return versionMetadata{}, 0, ctx.Err()
	}

	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return versionMetadata{}, 0, fmt.Errorf("failed to open zip reader: %w", err)
	}

	for _, file := range zipReader.File {
		if !isMainAppInfoPlist(file.Name) {
			continue
		}

		metadata, err := readVersionMetadataFromInfoPlist(file)
		if err != nil {
			return versionMetadata{}, 0, err
		}

		return metadata, size, nil
	}

	return versionMetadata{}, 0, errors.New("could not find Info.plist")
}

func isMainAppInfoPlist(name string) bool {
	parts := strings.Split(name, "/")

	return len(parts) == 3 && parts[0] == "Payload" && strings.HasSuffix(parts[1], ".app") && parts[2] == "Info.plist"
}

func readVersionMetadataFromInfoPlist(file *zip.File) (versionMetadata, error) {
	src, err := file.Open()
	if err != nil {
		return versionMetadata{}, fmt.Errorf("failed to open Info.plist: %w", err)
	}
	defer src.Close()

	const maxInfoPlistReadBytes int64 = 4 << 20
	data := new(bytes.Buffer)

	_, err = io.Copy(data, io.LimitReader(src, maxInfoPlistReadBytes+1))
	if err != nil {
		return versionMetadata{}, fmt.Errorf("failed to read Info.plist: %w", err)
	}
	if int64(data.Len()) > maxInfoPlistReadBytes {
		return versionMetadata{}, errors.New("Info.plist exceeds size limit")
	}

	metadata := map[string]interface{}{}

	_, err = plist.Unmarshal(data.Bytes(), &metadata)
	if err != nil {
		return versionMetadata{}, fmt.Errorf("failed to unmarshal Info.plist: %w", err)
	}

	displayVersion, err := readDisplayVersionFromMetadata(metadata)
	if err != nil {
		return versionMetadata{}, err
	}

	releaseDate, err := readReleaseDateFromInfoPlist(metadata, file.Modified)
	if err != nil {
		return versionMetadata{}, err
	}

	return versionMetadata{
		DisplayVersion: displayVersion,
		ReleaseDate:    releaseDate,
	}, nil
}

func readDisplayVersionFromMetadata(metadata map[string]interface{}) (string, error) {
	for _, key := range infoPlistDisplayVersionKeys {
		value, ok := metadata[key]
		if !ok {
			continue
		}

		displayVersion := strings.TrimSpace(fmt.Sprintf("%v", value))
		if displayVersion == "" || displayVersion == "<nil>" {
			return "", fmt.Errorf("%s is empty", key)
		}

		return displayVersion, nil
	}

	return "", errors.New("info plist does not contain a display version")
}

func readReleaseDateFromInfoPlist(metadata map[string]interface{}, modified time.Time) (time.Time, error) {
	for _, key := range infoPlistReleaseDateKeys {
		value, ok := metadata[key]
		if !ok {
			continue
		}

		releaseDate, err := parseReleaseDateValue(value)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse release date: %w", err)
		}

		return releaseDate, nil
	}

	if modified.IsZero() {
		return time.Time{}, errors.New("info plist does not contain a release date")
	}

	return modified.UTC(), nil
}

func parseReleaseDateValue(value interface{}) (time.Time, error) {
	switch val := value.(type) {
	case time.Time:
		return val.UTC(), nil
	case string:
		return parseReleaseDateString(val)
	case int:
		return time.Unix(int64(val), 0).UTC(), nil
	case int64:
		return time.Unix(val, 0).UTC(), nil
	case uint64:
		if val > math.MaxInt64 {
			return time.Time{}, fmt.Errorf("timestamp is too large: %d", val)
		}

		return time.Unix(int64(val), 0).UTC(), nil
	case float64:
		return time.Unix(int64(val), 0).UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported release date type %T", value)
	}
}

func parseReleaseDateString(value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"Monday, January 2, 2006",
		"2006-01-02",
	} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid release date: %s", value)
}
