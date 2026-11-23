package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	maxConcurrentDownloads = 2
	maxStoredDownloadTasks = 200
)

// --- Download Task Manager ---

func (a *App) nextTaskID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), a.taskSeq.Add(1))
}

func (a *App) ensureTaskMapsLocked() {
	if a.taskCancels == nil {
		a.taskCancels = make(map[string]context.CancelFunc)
	}
	if a.taskDone == nil {
		a.taskDone = make(map[string]chan struct{})
	}
	if a.downloadSem == nil {
		a.downloadSem = make(chan struct{}, maxConcurrentDownloads)
	}
}

func (a *App) downloadSemaphore() chan struct{} {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()
	a.ensureTaskMapsLocked()
	return a.downloadSem
}

func (a *App) currentDownloadPlatform() string {
	a.settingsMu.RLock()
	platform := strings.TrimSpace(a.settings.DefaultPlatform)
	a.settingsMu.RUnlock()
	if platform == "" {
		return "iphone"
	}
	return platform
}

func (a *App) AddDownloadTask(appName, bundleId string, appId int64, version, versionId, fileSize, platform string) (DownloadTask, error) {
	accountID := a.activeAccountID()
	if accountID == "" {
		return DownloadTask{}, fmt.Errorf("尚未登录 Apple ID")
	}
	resolvedPlatform := strings.TrimSpace(platform)
	if resolvedPlatform == "" {
		resolvedPlatform = a.currentDownloadPlatform()
	}
	if _, err := parseStorePlatform(resolvedPlatform); err != nil {
		return DownloadTask{}, err
	}

	if strings.TrimSpace(appName) == "" {
		appName = bundleId
	}
	if strings.TrimSpace(version) == "" || version == "未查询" {
		version = "最新版"
	}

	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	if a.shuttingDown {
		return DownloadTask{}, fmt.Errorf("软件正在退出，无法添加下载任务")
	}

	id := a.nextTaskID()
	task := newDownloadTask(accountID, id, appName, bundleId, appId, version, versionId, fileSize, resolvedPlatform)
	if a.hasActiveDuplicateLocked(task) {
		return DownloadTask{}, fmt.Errorf("该应用版本已在下载队列中")
	}

	a.tasks = append([]*DownloadTask{task}, a.tasks...)
	a.trimDownloadHistoryLocked()
	a.saveTasksLocked()
	a.startTaskLocked(task)
	return *task, nil
}

func newDownloadTask(accountID, id, appName, bundleID string, appID int64, version, versionID, fileSize, platform string) *DownloadTask {
	return &DownloadTask{
		ID:        id,
		AccountID: accountID,
		AppName:   appName,
		BundleID:  bundleID,
		AppID:     appID,
		Version:   version,
		VersionID: versionID,
		FileSize:  fileSize,
		Platform:  platform,
		Status:    "pending",
		Speed:     "等待下载",
		Progress:  0,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
}

func retryDownloadTaskFrom(source *DownloadTask, id, fallbackAccountID, fallbackPlatform string) (*DownloadTask, error) {
	if source == nil {
		return nil, fmt.Errorf("下载任务不存在")
	}
	if source.Status != "error" && source.Status != "canceled" {
		return nil, fmt.Errorf("只有失败或已取消的任务可以重试")
	}
	accountID := source.AccountID
	if accountID == "" {
		accountID = fallbackAccountID
	}
	if accountID == "" {
		return nil, fmt.Errorf("下载任务缺少可用账号")
	}
	platform := strings.TrimSpace(source.Platform)
	if platform == "" {
		platform = strings.TrimSpace(fallbackPlatform)
	}
	if platform == "" {
		platform = "iphone"
	}
	return newDownloadTask(accountID, id, source.AppName, source.BundleID, source.AppID, source.Version, source.VersionID, source.FileSize, platform), nil
}

func (a *App) RetryDownloadTask(id string) (DownloadTask, error) {
	fallbackPlatform := a.currentDownloadPlatform()
	fallbackAccountID := a.activeAccountID()

	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	if a.shuttingDown {
		return DownloadTask{}, fmt.Errorf("软件正在退出，无法重试下载任务")
	}

	var source *DownloadTask
	for _, task := range a.tasks {
		if task.ID == id {
			source = task
			break
		}
	}
	retry, err := retryDownloadTaskFrom(source, a.nextTaskID(), fallbackAccountID, fallbackPlatform)
	if err != nil {
		return DownloadTask{}, err
	}
	if _, ok := a.accountByID(retry.AccountID); !ok {
		return DownloadTask{}, fmt.Errorf("原下载账号已被移除，无法重试")
	}
	if a.hasActiveDuplicateLocked(retry) {
		return DownloadTask{}, fmt.Errorf("该应用版本已在下载队列中")
	}

	a.tasks = append([]*DownloadTask{retry}, a.tasks...)
	a.trimDownloadHistoryLocked()
	a.saveTasksLocked()
	a.startTaskLocked(retry)
	return *retry, nil
}

func (a *App) trimDownloadHistoryLocked() {
	for len(a.tasks) > maxStoredDownloadTasks {
		removeIdx := -1
		for i := len(a.tasks) - 1; i >= 0; i-- {
			t := a.tasks[i]
			if t == nil || (t.Status != "pending" && t.Status != "downloading") {
				removeIdx = i
				break
			}
		}
		if removeIdx == -1 {
			break
		}
		a.tasks = append(a.tasks[:removeIdx], a.tasks[removeIdx+1:]...)
	}
}

func (a *App) taskExistsLocked(task *DownloadTask) bool {
	for _, item := range a.tasks {
		if item == task {
			return true
		}
	}
	return false
}

func (a *App) hasActiveDuplicateLocked(candidate *DownloadTask) bool {
	for _, item := range a.tasks {
		if item.Status != "pending" && item.Status != "downloading" {
			continue
		}
		if item.AccountID == candidate.AccountID &&
			strings.EqualFold(strings.TrimSpace(item.BundleID), strings.TrimSpace(candidate.BundleID)) &&
			item.AppID == candidate.AppID &&
			item.VersionID == candidate.VersionID &&
			item.Platform == candidate.Platform {
			return true
		}
	}
	return false
}

func (a *App) startTaskLocked(task *DownloadTask) {
	a.ensureTaskMapsLocked()
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	a.taskCancels[task.ID] = cancel
	a.taskDone[task.ID] = doneCh
	go a.runDownloadTaskWithSync(ctx, task, doneCh, a.downloadSem)
}

func (a *App) runDownloadTask(ctx context.Context, task *DownloadTask) {
	a.runDownloadTaskWithSync(ctx, task, nil, a.downloadSemaphore())
}

func (a *App) runDownloadTaskWithSync(ctx context.Context, task *DownloadTask, doneCh chan struct{}, sem chan struct{}) {
	defer func() {
		a.tasksMu.Lock()
		if cancel, ok := a.taskCancels[task.ID]; ok {
			cancel()
			delete(a.taskCancels, task.ID)
		}
		if ch, ok := a.taskDone[task.ID]; ok && ch == doneCh {
			delete(a.taskDone, task.ID)
		}
		a.tasksMu.Unlock()
		if doneCh != nil {
			close(doneCh)
		}
	}()

	if sem != nil {
		a.tasksMu.Lock()
		if ctx.Err() != nil || task.Status != "pending" || !a.taskExistsLocked(task) {
			a.tasksMu.Unlock()
			return
		}
		if len(sem) >= cap(sem) {
			task.Speed = "排队等待中"
			queuedSnap := *task
			a.tasksMu.Unlock()
			a.emitTaskUpdated(queuedSnap)
		} else {
			a.tasksMu.Unlock()
		}

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			a.finishDownloadTask(ctx, task, DownloadResult{}, ctx.Err())
			return
		}
	}

	a.tasksMu.Lock()
	if ctx.Err() != nil || task.Status != "pending" || !a.taskExistsLocked(task) {
		a.tasksMu.Unlock()
		return
	}
	task.Status = "downloading"
	task.Speed = "连接中..."
	platform := strings.TrimSpace(task.Platform)
	a.saveTasksLocked()
	snap := *task
	a.tasksMu.Unlock()
	a.emitTaskUpdated(snap)

	if platform == "" {
		platform = a.currentDownloadPlatform()
	}

	result, err := a.downloadFromStore(
		ctx,
		snap.AccountID,
		snap.BundleID,
		snap.AppID,
		snap.VersionID,
		a.getDownloadsDirForAccount(snap.AccountID),
		platform,
		true,
		func(progress downloadProgress) {
			a.tasksMu.Lock()
			if task.Status != "downloading" || !a.taskExistsLocked(task) {
				a.tasksMu.Unlock()
				return
			}
			task.CurrBytes = progress.Current
			task.Progress = progress.Percent
			if progress.Total > 0 {
				task.TotalBytes = progress.Total
			}
			if progress.Speed > 0 {
				task.Speed = formatBytes(progress.Speed) + "/s"
			} else {
				task.Speed = "处理中..."
			}
			progressSnap := *task
			a.tasksMu.Unlock()
			a.emitTaskUpdated(progressSnap)
		},
	)

	a.finishDownloadTask(ctx, task, result, err)
}

func (a *App) finishDownloadTask(ctx context.Context, task *DownloadTask, result DownloadResult, err error) {
	a.tasksMu.Lock()
	if !a.taskExistsLocked(task) {
		a.tasksMu.Unlock()
		return
	}
	if ctx.Err() != nil {
		if task.Status == "canceled" {
			a.tasksMu.Unlock()
			return
		}
		task.Status = "canceled"
		task.Speed = "已取消"
	} else if err != nil {
		task.Status = "error"
		task.ErrorMessage = err.Error()
		task.Speed = "下载失败"
	} else {
		task.Status = "completed"
		task.Progress = 100
		task.Speed = "已完成"
		task.OutputPath = result.Output
	}
	a.saveTasksLocked()
	snap := *task
	a.tasksMu.Unlock()
	a.emitTaskUpdated(snap)
}

func (a *App) emitTaskUpdated(task DownloadTask) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-task-updated", task)
	}
}

func (a *App) GetDownloadTasks() []DownloadTask {
	a.tasksMu.RLock()
	defer a.tasksMu.RUnlock()
	res := make([]DownloadTask, len(a.tasks))
	for i, t := range a.tasks {
		res[i] = *t
	}
	return res
}

func (a *App) CancelDownloadTask(id string) {
	var snap *DownloadTask
	a.tasksMu.Lock()
	if cancel, ok := a.taskCancels[id]; ok {
		cancel()
	}
	for _, t := range a.tasks {
		if t.ID == id && (t.Status == "downloading" || t.Status == "pending") {
			t.Status = "canceled"
			t.Speed = "已取消"
			a.saveTasksLocked()
			copyTask := *t
			snap = &copyTask
			break
		}
	}
	a.tasksMu.Unlock()
	if snap != nil {
		a.emitTaskUpdated(*snap)
	}
}

func (a *App) cancelDownloadTasksForAccount(accountID string) {
	if accountID == "" {
		return
	}
	var snaps []DownloadTask
	var waitChans []chan struct{}
	changed := false

	a.tasksMu.Lock()
	for _, t := range a.tasks {
		if t.AccountID != accountID || (t.Status != "downloading" && t.Status != "pending") {
			continue
		}
		if cancel, ok := a.taskCancels[t.ID]; ok {
			cancel()
		}
		if doneCh, ok := a.taskDone[t.ID]; ok && doneCh != nil {
			waitChans = append(waitChans, doneCh)
		}
		t.Status = "canceled"
		t.Speed = "账号已移除"
		snaps = append(snaps, *t)
		changed = true
	}
	if changed {
		a.saveTasksLocked()
	}
	a.tasksMu.Unlock()

	for _, snap := range snaps {
		a.emitTaskUpdated(snap)
	}
	waitForTaskChannels(waitChans, 2*time.Second)
}

func waitForTaskChannels(waitChans []chan struct{}, timeout time.Duration) {
	if len(waitChans) == 0 {
		return
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for _, ch := range waitChans {
		select {
		case <-ch:
		case <-timer.C:
			return
		}
	}
}

func (a *App) DeleteDownloadTask(id string) {
	a.tasksMu.Lock()
	if cancel, ok := a.taskCancels[id]; ok {
		cancel()
		delete(a.taskCancels, id)
	}
	for i, t := range a.tasks {
		if t.ID == id {
			a.tasks = append(a.tasks[:i], a.tasks[i+1:]...)
			break
		}
	}
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-tasks-reload")
	}
}

func (a *App) ClearCompletedDownloadTasks() {
	a.tasksMu.Lock()
	active := make([]*DownloadTask, 0)
	for _, t := range a.tasks {
		if t.Status == "downloading" || t.Status == "pending" {
			active = append(active, t)
		}
	}
	a.tasks = active
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-tasks-reload")
	}
}
