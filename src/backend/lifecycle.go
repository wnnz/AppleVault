package backend

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) activeDownloadCount() int {
	a.tasksMu.RLock()
	defer a.tasksMu.RUnlock()
	count := 0
	for _, t := range a.tasks {
		if t.Status == "pending" || t.Status == "downloading" {
			count++
		}
	}
	return count
}

func (a *App) beforeClose(ctx context.Context) bool {
	count := a.activeDownloadCount()
	if count == 0 {
		return false
	}
	choice, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "确认退出",
		Message:       fmt.Sprintf("还有 %d 个下载任务正在进行，退出将中断下载。确定要退出吗？", count),
		DefaultButton: "No",
	})
	if err != nil {
		return false
	}
	return choice != "Yes" && choice != "退出"
}

func (a *App) shutdown(ctx context.Context) {
	a.CancelRunningCommand()
	a.tasksMu.Lock()
	for id, cancel := range a.taskCancels {
		cancel()
		delete(a.taskCancels, id)
	}
	for _, t := range a.tasks {
		if t.Status == "pending" || t.Status == "downloading" {
			t.Status = "canceled"
			t.Speed = "已中断"
		}
	}
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	a.invalidateStoreSessions("")
}

func (a *App) onSecondInstanceLaunch(_ options.SecondInstanceData) {
	if a.ctx == nil {
		return
	}
	runtime.WindowUnminimise(a.ctx)
	runtime.Show(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
}

func BeforeCloseHandler(a *App) func(context.Context) bool {
	return a.beforeClose
}

func ShutdownHandler(a *App) func(context.Context) {
	return a.shutdown
}

func SecondInstanceHandler(a *App) func(options.SecondInstanceData) {
	return a.onSecondInstanceLaunch
}
