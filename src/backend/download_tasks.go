package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- Download Task Manager ---

func (a *App) AddDownloadTask(appName, bundleId string, appId int64, version, versionId, fileSize string) (*DownloadTask, error) {
	accountID := a.activeAccountID()
	if accountID == "" {
		return nil, fmt.Errorf("尚未登录 Apple ID")
	}
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	if strings.TrimSpace(appName) == "" {
		appName = bundleId
	}
	if strings.TrimSpace(version) == "" || version == "未查询" {
		version = "最新版"
	}

	task := newDownloadTask(accountID, id, appName, bundleId, appId, version, versionId, fileSize)

	a.tasks = append([]*DownloadTask{task}, a.tasks...)
	a.saveTasksLocked()
	go a.runDownloadTask(task)

	return task, nil
}

func newDownloadTask(accountID, id, appName, bundleID string, appID int64, version, versionID, fileSize string) *DownloadTask {
	return &DownloadTask{
		ID:        id,
		AccountID: accountID,
		AppName:   appName,
		BundleID:  bundleID,
		AppID:     appID,
		Version:   version,
		VersionID: versionID,
		FileSize:  fileSize,
		Status:    "pending",
		Speed:     "等待下载",
		Progress:  0,
		CreatedAt: time.Now().Format("15:04:05"),
	}
}

func retryDownloadTaskFrom(source *DownloadTask, id, fallbackAccountID string) (*DownloadTask, error) {
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
	return newDownloadTask(accountID, id, source.AppName, source.BundleID, source.AppID, source.Version, source.VersionID, source.FileSize), nil
}

func (a *App) RetryDownloadTask(id string) (*DownloadTask, error) {
	a.tasksMu.Lock()
	var source *DownloadTask
	for _, task := range a.tasks {
		if task.ID == id {
			source = task
			break
		}
	}
	retry, err := retryDownloadTaskFrom(source, fmt.Sprintf("%d", time.Now().UnixNano()), a.activeAccountID())
	if err != nil {
		a.tasksMu.Unlock()
		return nil, err
	}
	if _, ok := a.accountByID(retry.AccountID); !ok {
		a.tasksMu.Unlock()
		return nil, fmt.Errorf("原下载账号已被移除，无法重试")
	}
	a.tasks = append([]*DownloadTask{retry}, a.tasks...)
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	go a.runDownloadTask(retry)
	return retry, nil
}

func (a *App) runDownloadTask(task *DownloadTask) {
	ctx, cancel := context.WithCancel(context.Background())
	a.tasksMu.Lock()
	a.taskCancels[task.ID] = cancel
	task.Status = "downloading"
	task.Speed = "连接中..."
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	a.emitTaskUpdated(task)

	defer func() {
		a.tasksMu.Lock()
		delete(a.taskCancels, task.ID)
		a.tasksMu.Unlock()
	}()

	a.settingsMu.RLock()
	platform := a.settings.DefaultPlatform
	a.settingsMu.RUnlock()

	result, err := a.downloadFromStore(
		ctx,
		task.AccountID,
		task.BundleID,
		task.AppID,
		task.VersionID,
		a.getDownloadsDirForAccount(task.AccountID),
		platform,
		true,
		func(progress downloadProgress) {
			a.tasksMu.Lock()
			task.CurrBytes = progress.Current
			task.Progress = progress.Percent
			if progress.Percent > 0 {
				task.TotalBytes = progress.Current * 100 / int64(progress.Percent)
			}
			if progress.Speed > 0 {
				task.Speed = formatBytes(progress.Speed) + "/s"
			} else {
				task.Speed = "处理中..."
			}
			a.tasksMu.Unlock()
			a.emitTaskUpdated(task)
		},
	)

	if ctx.Err() != nil {
		a.tasksMu.Lock()
		task.Status = "canceled"
		task.Speed = "已取消"
		a.saveTasksLocked()
		a.tasksMu.Unlock()
		a.emitTaskUpdated(task)
		return
	}
	if err != nil {
		a.failTask(task, err.Error())
		return
	}

	a.tasksMu.Lock()
	task.Status = "completed"
	task.Progress = 100
	task.Speed = "已完成"
	task.OutputPath = result.Output
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	a.emitTaskUpdated(task)
}

func (a *App) failTask(task *DownloadTask, msg string) {
	a.tasksMu.Lock()
	task.Status = "error"
	task.ErrorMessage = msg
	task.Speed = "下载失败"
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	a.emitTaskUpdated(task)
}

func (a *App) emitTaskUpdated(task *DownloadTask) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-task-updated", task)
	}
}

func (a *App) GetDownloadTasks() []*DownloadTask {
	a.tasksMu.RLock()
	defer a.tasksMu.RUnlock()
	res := make([]*DownloadTask, len(a.tasks))
	copy(res, a.tasks)
	return res
}

func (a *App) CancelDownloadTask(id string) {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()
	if cancel, ok := a.taskCancels[id]; ok {
		cancel()
	}
	for _, t := range a.tasks {
		if t.ID == id && (t.Status == "downloading" || t.Status == "pending") {
			t.Status = "canceled"
			t.Speed = "已取消"
			a.saveTasksLocked()
			a.emitTaskUpdated(t)
			break
		}
	}
}

func (a *App) DeleteDownloadTask(id string) {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()
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
