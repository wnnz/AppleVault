package backend

import (
	"context"
	"errors"
	"testing"
)

func TestHasActiveDuplicateTask(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	base := newDownloadTask("acc1", "t1", "WeChat", "com.tencent.xin", 100, "8.0.50", "800", "200 MB", "iphone")
	completed := newDownloadTask("acc1", "t2", "WeChat", "com.tencent.xin", 100, "8.0.49", "799", "199 MB", "iphone")
	completed.Status = "completed"

	app.tasksMu.Lock()
	defer app.tasksMu.Unlock()
	app.tasks = []*DownloadTask{base, completed}

	samePending := newDownloadTask("acc1", "t3", "WeChat", "com.tencent.xin", 100, "8.0.50", "800", "200 MB", "iphone")
	if !app.hasActiveDuplicateLocked(samePending) {
		t.Fatal("expected pending task with same account/app/version/platform to be flagged as duplicate")
	}

	sameAsCompleted := newDownloadTask("acc1", "t4", "WeChat", "com.tencent.xin", 100, "8.0.49", "799", "199 MB", "iphone")
	if app.hasActiveDuplicateLocked(sameAsCompleted) {
		t.Fatal("completed task should not block adding a new task for that version")
	}

	diffVersion := newDownloadTask("acc1", "t5", "WeChat", "com.tencent.xin", 100, "8.0.51", "801", "201 MB", "iphone")
	if app.hasActiveDuplicateLocked(diffVersion) {
		t.Fatal("different VersionID should not be flagged as duplicate")
	}

	diffPlatform := newDownloadTask("acc1", "t6", "WeChat", "com.tencent.xin", 100, "8.0.50", "800", "200 MB", "ipad")
	if app.hasActiveDuplicateLocked(diffPlatform) {
		t.Fatal("different Platform should not be flagged as duplicate")
	}
}

func TestRunDownloadTaskSkipsCanceledTask(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	task := newDownloadTask("acc1", "t1", "WeChat", "com.tencent.xin", 100, "8.0.50", "800", "200 MB", "iphone")
	task.Status = "canceled"
	task.Speed = "已取消"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	app.tasksMu.Lock()
	app.tasks = []*DownloadTask{task}
	app.taskCancels[task.ID] = cancel
	app.tasksMu.Unlock()

	app.runDownloadTask(ctx, task)

	app.tasksMu.RLock()
	defer app.tasksMu.RUnlock()
	if task.Status != "canceled" {
		t.Fatalf("expected canceled task to stay canceled, got %q", task.Status)
	}
	if len(app.taskCancels) != 0 {
		t.Fatalf("expected taskCancels to be cleaned up, got %d entries", len(app.taskCancels))
	}
}

func TestFinishDownloadTaskIgnoresDeletedTask(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	task := newDownloadTask("acc1", "t1", "WeChat", "com.tencent.xin", 100, "8.0.50", "800", "200 MB", "iphone")
	task.Status = "downloading"

	app.finishDownloadTask(context.Background(), task, DownloadResult{Output: "out.ipa", Success: true}, nil)

	app.tasksMu.RLock()
	defer app.tasksMu.RUnlock()
	if len(app.tasks) != 0 {
		t.Fatalf("expected tasks slice to remain empty, got %d", len(app.tasks))
	}
	if task.Status != "downloading" {
		t.Fatalf("deleted task should not be mutated on finish, got %q", task.Status)
	}
}

func TestFinishDownloadTaskStates(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	completedTask := newDownloadTask("acc1", "t-ok", "App", "com.example.ok", 1, "1.0", "1", "10 MB", "iphone")
	completedTask.Status = "downloading"

	errorTask := newDownloadTask("acc1", "t-err", "App", "com.example.err", 2, "1.0", "2", "10 MB", "iphone")
	errorTask.Status = "downloading"

	canceledTask := newDownloadTask("acc1", "t-cancel", "App", "com.example.cancel", 3, "1.0", "3", "10 MB", "iphone")
	canceledTask.Status = "downloading"

	app.tasksMu.Lock()
	app.tasks = []*DownloadTask{completedTask, errorTask, canceledTask}
	app.tasksMu.Unlock()

	app.finishDownloadTask(context.Background(), completedTask, DownloadResult{Output: "/tmp/app.ipa", Success: true}, nil)
	if completedTask.Status != "completed" || completedTask.Progress != 100 || completedTask.OutputPath != "/tmp/app.ipa" {
		t.Fatalf("unexpected completed state: %#v", completedTask)
	}

	app.finishDownloadTask(context.Background(), errorTask, DownloadResult{}, errors.New("network failure"))
	if errorTask.Status != "error" || errorTask.ErrorMessage != "network failure" {
		t.Fatalf("unexpected error state: %#v", errorTask)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	app.finishDownloadTask(canceledCtx, canceledTask, DownloadResult{}, errors.New("context canceled"))
	if canceledTask.Status != "canceled" {
		t.Fatalf("unexpected canceled state: %#v", canceledTask)
	}
}

func TestGetDownloadTasksReturnsCopies(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	task := newDownloadTask("acc1", "t1", "App", "com.example", 1, "1.0", "1", "10 MB", "iphone")
	app.tasksMu.Lock()
	app.tasks = []*DownloadTask{task}
	app.tasksMu.Unlock()

	got := app.GetDownloadTasks()
	if len(got) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got))
	}
	got[0].Status = "error"
	got[0].Progress = 99

	app.tasksMu.RLock()
	defer app.tasksMu.RUnlock()
	if app.tasks[0].Status != "pending" || app.tasks[0].Progress != 0 {
		t.Fatalf("mutating GetDownloadTasks result affected internal task: %#v", app.tasks[0])
	}
}
