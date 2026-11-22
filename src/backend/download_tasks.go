package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- Download Task Manager ---

func (a *App) currentDownloadPlatform() string {
	a.settingsMu.RLock()
	platform := strings.TrimSpace(a.settings.DefaultPlatform)
	a.settingsMu.RUnlock()
	if platform == "" {
		return "iphone"
	}
	return platform
}

func (a *App) AddDownloadTask(appName, bundleId string, appId int64, version, versionId, fileSize string) (DownloadTask, error) {
	accountID := a.activeAccountID()
	if accountID == "" {
		return DownloadTask{}, fmt.Errorf("尚未登录 Apple ID")
	}
	platform := a.currentDownloadPlatform()

	if strings.TrimSpace(appName) == "" {
		appName = bundleId
	}
	if strings.TrimSpace(version) == "" || version == "未查询" {
		version = "最新版"
	}

	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	task := newDownloadTask(accountID, id, appName, bundleId, appId, version, versionId, fileSize, platform)
	if a.hasActiveDuplicateLocked(task) {
		return DownloadTask{}, fmt.Errorf("该应用版本已在下载队列中")
	}

	a.tasks = append([]*DownloadTask{task}, a.tasks...)
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
		CreatedAt: time.Now().Format("15:04:05"),
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

	var source *DownloadTask
	for _, task := range a.tasks {
		if task.ID == id {
			source = task
			break
		}
	}
	retry, err := retryDownloadTaskFrom(source, fmt.Sprintf("%d", time.Now().UnixNano()), fallbackAccountID, fallbackPlatform)
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
	a.saveTasksLocked()
	a.startTaskLocked(retry)
	return *retry, nil
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
			item.BundleID == candidate.BundleID &&
			item.AppID == candidate.AppID &&
			item.VersionID == candidate.VersionID &&
			item.Platform == candidate.Platform {
			return true
		}
	}
	return false
}

func (a *App) startTaskLocked(task *DownloadTask) {
	ctx, cancel := context.WithCancel(context.Background())
	a.taskCancels[task.ID] = cancel
	go a.runDownloadTask(ctx, task)
}

func (a *App) runDownloadTask(ctx context.Context, task *DownloadTask) {
	defer func() {
		a.tasksMu.Lock()
		if cancel, ok := a.taskCancels[task.ID]; ok {
			cancel()
			delete(a.taskCancels, task.ID)
		}
		a.tasksMu.Unlock()
	}()

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
