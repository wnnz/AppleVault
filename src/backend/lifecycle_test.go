package backend

import (
	"context"
	"testing"
)

func TestActiveDownloadCount(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	pending := newDownloadTask("acc", "1", "A", "com.a", 1, "1.0", "1", "1 MB", "iphone")
	downloading := newDownloadTask("acc", "2", "B", "com.b", 2, "1.0", "2", "1 MB", "iphone")
	downloading.Status = "downloading"
	completed := newDownloadTask("acc", "3", "C", "com.c", 3, "1.0", "3", "1 MB", "iphone")
	completed.Status = "completed"
	errored := newDownloadTask("acc", "4", "D", "com.d", 4, "1.0", "4", "1 MB", "iphone")
	errored.Status = "error"
	canceled := newDownloadTask("acc", "5", "E", "com.e", 5, "1.0", "5", "1 MB", "iphone")
	canceled.Status = "canceled"

	app.tasksMu.Lock()
	app.tasks = []*DownloadTask{pending, downloading, completed, errored, canceled}
	app.tasksMu.Unlock()

	if got := app.activeDownloadCount(); got != 2 {
		t.Fatalf("expected activeDownloadCount = 2, got %d", got)
	}
}

func TestShutdownInterruptsActiveTasks(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	pending := newDownloadTask("acc", "1", "A", "com.a", 1, "1.0", "1", "1 MB", "iphone")
	downloading := newDownloadTask("acc", "2", "B", "com.b", 2, "1.0", "2", "1 MB", "iphone")
	downloading.Status = "downloading"
	completed := newDownloadTask("acc", "3", "C", "com.c", 3, "1.0", "3", "1 MB", "iphone")
	completed.Status = "completed"
	completed.Speed = "已完成"

	canceledCalled := false
	app.tasksMu.Lock()
	app.tasks = []*DownloadTask{pending, downloading, completed}
	app.taskCancels["2"] = func() { canceledCalled = true }
	app.tasksMu.Unlock()

	app.shutdown(context.Background())

	if !canceledCalled {
		t.Fatal("expected registered task cancel function to be invoked on shutdown")
	}

	app.tasksMu.RLock()
	defer app.tasksMu.RUnlock()
	if len(app.taskCancels) != 0 {
		t.Fatalf("expected taskCancels to be cleared, got %d entries", len(app.taskCancels))
	}
	if pending.Status != "canceled" || pending.Speed != "已中断" {
		t.Fatalf("expected pending task canceled/已中断, got %q/%q", pending.Status, pending.Speed)
	}
	if downloading.Status != "canceled" || downloading.Speed != "已中断" {
		t.Fatalf("expected downloading task canceled/已中断, got %q/%q", downloading.Status, downloading.Speed)
	}
	if completed.Status != "completed" || completed.Speed != "已完成" {
		t.Fatalf("completed task should remain untouched, got %q/%q", completed.Status, completed.Speed)
	}
}
