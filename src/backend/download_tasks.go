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
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	if strings.TrimSpace(appName) == "" {
		appName = bundleId
	}
	if strings.TrimSpace(version) == "" || version == "未查询" {
		version = "最新版"
	}

	task := &DownloadTask{
		ID:        id,
		AppName:   appName,
		BundleID:  bundleId,
		AppID:     appId,
		Version:   version,
		VersionID: versionId,
		FileSize:  fileSize,
		Status:    "pending",
		Speed:     "等待下载",
		Progress:  0,
		CreatedAt: time.Now().Format("15:04:05"),
	}

	a.tasks = append([]*DownloadTask{task}, a.tasks...)
	a.saveTasksLocked()
	go a.runDownloadTask(task)

	return task, nil
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
		task.BundleID,
		task.AppID,
		task.VersionID,
		a.getDownloadsDir(),
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
