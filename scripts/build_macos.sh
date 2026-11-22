#!/usr/bin/env bash
set -euo pipefail

ARCH="universal"
SKIP_FRONTEND=false
GO_IOS_VERSION="v1.3.2"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --arch)
      ARCH="${2:-}"
      shift 2
      ;;
    --skip-frontend)
      SKIP_FRONTEND=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [--arch universal|arm64|amd64] [--skip-frontend]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

case "$ARCH" in
  universal|arm64|amd64) ;;
  *)
    echo "Unsupported --arch value: $ARCH (expected universal, arm64, or amd64)" >&2
    exit 1
    ;;
esac

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script must run on macOS (Darwin)." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

IFS=$'\x1f' read -r VERSION PRODUCT_NAME COPYRIGHT COMMENTS < <(python3 - "$REPO_ROOT" <<'PY'
import json, pathlib, re, sys

root = pathlib.Path(sys.argv[1])
update_go = (root / "src/backend/update.go").read_text(encoding="utf-8")
m = re.search(r'AppVersion\s*=\s*"([^"]+)"', update_go)
if not m:
    raise SystemExit("Failed to parse AppVersion from src/backend/update.go")
update_ver = m.group(1)

pkg_ver = json.loads((root / "src/frontend/package.json").read_text(encoding="utf-8"))["version"]
wails_cfg = json.loads((root / "wails.json").read_text(encoding="utf-8"))
info = wails_cfg.get("info", {})
wails_ver = info.get("productVersion", "")

if not (update_ver == pkg_ver == wails_ver):
    raise SystemExit(f"Version mismatch: update.go={update_ver}, package.json={pkg_ver}, wails.json={wails_ver}")

vals = [
    update_ver,
    info.get("productName") or wails_cfg.get("name", "AppleVault"),
    info.get("copyright") or "AppleVault",
    info.get("comments") or "AppleVault",
]
print("\x1f".join(vals))
PY
)

echo "==> [1/5] Building AppleVault v${VERSION} for macOS (${ARCH})..."

if [[ "$SKIP_FRONTEND" == false ]]; then
  pushd "$REPO_ROOT/src/frontend" >/dev/null
  if command -v pnpm >/dev/null 2>&1; then
    pnpm install --frozen-lockfile
    pnpm run build
  else
    npm install
    npm run build
  fi
  popd >/dev/null
fi

WORK_TMP="$(mktemp -d)"
cleanup() {
  rm -rf "$WORK_TMP"
}
trap cleanup EXIT

echo "==> [2/5] Compiling Go desktop binary..."
export CGO_ENABLED=1
export CGO_CFLAGS="${CGO_CFLAGS:-} -mmacosx-version-min=11.0"
export CGO_LDFLAGS="${CGO_LDFLAGS:-} -framework UniformTypeIdentifiers -mmacosx-version-min=11.0"

build_arch() {
  local target_arch="$1"
  local out_bin="$2"
  GOOS=darwin GOARCH="$target_arch" go build -buildvcs=false -trimpath -tags "desktop,production" -ldflags "-w -s" -o "$out_bin" ./src
}

BIN_PATH="$WORK_TMP/AppleVault"
if [[ "$ARCH" == "universal" ]]; then
  build_arch arm64 "$WORK_TMP/AppleVault-arm64"
  build_arch amd64 "$WORK_TMP/AppleVault-amd64"
  lipo -create -output "$BIN_PATH" "$WORK_TMP/AppleVault-arm64" "$WORK_TMP/AppleVault-amd64"
else
  build_arch "$ARCH" "$BIN_PATH"
fi

echo "==> [3/5] Assembling AppleVault.app bundle..."
BIN_DIR="$REPO_ROOT/build/bin"
RELEASE_DIR="$REPO_ROOT/build/release"
mkdir -p "$BIN_DIR" "$RELEASE_DIR"

APP_DIR="$BIN_DIR/AppleVault.app"
CONTENTS_DIR="$APP_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"
TOOLS_DIR="$RESOURCES_DIR/tools"

rm -rf "$APP_DIR"
mkdir -p "$MACOS_DIR" "$TOOLS_DIR"
cp "$BIN_PATH" "$MACOS_DIR/AppleVault"
chmod 0755 "$MACOS_DIR/AppleVault"

python3 - "$REPO_ROOT/build/darwin/Info.plist" "$CONTENTS_DIR/Info.plist" "$PRODUCT_NAME" "$VERSION" "$COMMENTS" "$COPYRIGHT" <<'PY'
import html, pathlib, re, sys

src, dst, product_name, version, comments, copyright_text = sys.argv[1:7]
text = pathlib.Path(src).read_text(encoding="utf-8")

def strip_template_if(content: str, expr: str) -> str:
    pattern = re.compile(r"\{\{if " + re.escape(expr) + r"\}\}")
    while True:
        m = pattern.search(content)
        if not m:
            break
        start = m.start()
        pos = m.end()
        depth = 1
        for tok in re.finditer(r"\{\{(if\b|range\b|with\b|end\}\})", content[pos:]):
            word = tok.group(1)
            if word.startswith(("if", "range", "with")):
                depth += 1
            else:
                depth -= 1
                if depth == 0:
                    end = pos + tok.end()
                    content = content[:start] + content[end:]
                    break
        else:
            raise SystemExit(f"Unclosed template block: {expr}")
    return content

text = strip_template_if(text, ".Info.FileAssociations")
text = strip_template_if(text, ".Info.Protocols")

replacements = {
    "{{.Info.ProductName}}": html.escape(product_name, quote=False),
    "{{.OutputFilename}}": "AppleVault",
    "{{.Name}}": "AppleVault",
    "{{.Info.ProductVersion}}": html.escape(version, quote=False),
    "{{.Info.Comments}}": html.escape(comments, quote=False),
    "{{.Info.Copyright}}": html.escape(copyright_text, quote=False),
}
for k, v in replacements.items():
    text = text.replace(k, v)

lines = [line for line in text.splitlines() if line.strip() != ""]
pathlib.Path(dst).write_text("\n".join(lines) + "\n", encoding="utf-8")
PY

plutil -lint "$CONTENTS_DIR/Info.plist"

ICONSET_DIR="$WORK_TMP/AppIcon.iconset"
mkdir -p "$ICONSET_DIR"
ICON_PNG="$REPO_ROOT/build/appicon.png"
render_icon() {
  local size="$1"
  local name="$2"
  sips -z "$size" "$size" "$ICON_PNG" --out "$ICONSET_DIR/$name" >/dev/null
}
render_icon 16   "icon_16x16.png"
render_icon 32   "icon_16x16@2x.png"
render_icon 32   "icon_32x32.png"
render_icon 64   "icon_32x32@2x.png"
render_icon 128  "icon_128x128.png"
render_icon 256  "icon_128x128@2x.png"
render_icon 256  "icon_256x256.png"
render_icon 512  "icon_256x256@2x.png"
render_icon 512  "icon_512x512.png"
render_icon 1024 "icon_512x512@2x.png"
iconutil -c icns "$ICONSET_DIR" -o "$RESOURCES_DIR/iconfile.icns"

echo "==> [4/5] Bundling go-ios device tool and signing..."
IOS_DST="$TOOLS_DIR/ios"
if [[ -f "$REPO_ROOT/tools/ios" ]]; then
  cp "$REPO_ROOT/tools/ios" "$IOS_DST"
else
  GO_IOS_ZIP="$WORK_TMP/go-ios-mac.zip"
  curl -fSL --retry 3 "https://github.com/danielpaulus/go-ios/releases/download/${GO_IOS_VERSION}/go-ios-mac.zip" -o "$GO_IOS_ZIP"
  unzip -q -o "$GO_IOS_ZIP" -d "$WORK_TMP/go-ios"
  IOS_SRC="$(find "$WORK_TMP/go-ios" -type f -name ios | head -n 1)"
  if [[ -z "$IOS_SRC" ]]; then
    echo "Failed to locate ios binary inside go-ios-mac.zip" >&2
    exit 1
  fi
  cp "$IOS_SRC" "$IOS_DST"
fi
chmod 0755 "$IOS_DST"

if [[ -f "$REPO_ROOT/third_party/ipatool-LICENSE.txt" ]]; then
  cp "$REPO_ROOT/third_party/ipatool-LICENSE.txt" "$RESOURCES_DIR/ipatool-LICENSE.txt"
fi
if [[ -f "$REPO_ROOT/LICENSE" ]]; then
  cp "$REPO_ROOT/LICENSE" "$RESOURCES_DIR/LICENSE"
fi

codesign --force --sign - "$IOS_DST"
codesign --force --deep --sign - "$APP_DIR"

echo "==> [5/5] Creating .zip and .dmg packages..."
ZIP_NAME="AppleVault-v${VERSION}-macos-${ARCH}.zip"
DMG_NAME="AppleVault-v${VERSION}-macos-${ARCH}.dmg"
ZIP_PATH="$RELEASE_DIR/$ZIP_NAME"
DMG_PATH="$RELEASE_DIR/$DMG_NAME"

rm -f "$ZIP_PATH" "$ZIP_PATH.sha256" "$DMG_PATH" "$DMG_PATH.sha256"
ditto -c -k --sequesterRsrc --keepParent "$APP_DIR" "$ZIP_PATH"

DMG_STAGE="$WORK_TMP/dmg"
mkdir -p "$DMG_STAGE"
cp -R "$APP_DIR" "$DMG_STAGE/AppleVault.app"
ln -s /Applications "$DMG_STAGE/Applications"
if [[ -f "$REPO_ROOT/README.md" ]]; then
  cp "$REPO_ROOT/README.md" "$DMG_STAGE/README.md"
fi
if [[ -f "$REPO_ROOT/LICENSE" ]]; then
  cp "$REPO_ROOT/LICENSE" "$DMG_STAGE/LICENSE"
fi
cat <<'EOF' > "$DMG_STAGE/首次打开说明.txt"
请将 AppleVault.app 拖入右侧 Applications（应用程序）文件夹后再打开。

若 macOS 提示“AppleVault 已损坏”或“无法验证开发者”（因为应用未经过 Apple 公证），请在终端运行以下命令后重新打开：
sudo xattr -cr /Applications/AppleVault.app
EOF

for attempt in 1 2 3; do
  if hdiutil create -volname "AppleVault v${VERSION}" -srcfolder "$DMG_STAGE" -ov -format UDZO "$DMG_PATH"; then
    break
  fi
  if [[ "$attempt" -eq 3 ]]; then
    echo "Failed to create DMG after 3 attempts" >&2
    exit 1
  fi
  sleep 3
done

(
  cd "$RELEASE_DIR"
  shasum -a 256 "$ZIP_NAME" > "${ZIP_NAME}.sha256"
  shasum -a 256 "$DMG_NAME" > "${DMG_NAME}.sha256"
)

echo "Done:"
ls -lh "$ZIP_PATH" "$DMG_PATH"
