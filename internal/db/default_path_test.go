package db

import (
	"os"
	"testing"
)

func TestDefaultDBPath_WithEnvVar(t *testing.T) {
	original := os.Getenv("TASKFLOW_DIR")
	defer func() { os.Setenv("TASKFLOW_DIR", original) }()

	os.Setenv("TASKFLOW_DIR", "/tmp/custom/taskflow")

	got := DefaultDBPath()
	want := "/tmp/custom/taskflow/tasks.db"
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

func TestDefaultDBPath_WithoutEnvVar(t *testing.T) {
	original := os.Getenv("TASKFLOW_DIR")
	defer func() { os.Setenv("TASKFLOW_DIR", original) }()

	os.Unsetenv("TASKFLOW_DIR")

	got := DefaultDBPath()
	want := ".taskflow/tasks.db"
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

func TestDefaultDBPath_WithEmptyEnvVar(t *testing.T) {
	original := os.Getenv("TASKFLOW_DIR")
	defer func() { os.Setenv("TASKFLOW_DIR", original) }()

	os.Setenv("TASKFLOW_DIR", "")

	got := DefaultDBPath()
	want := ".taskflow/tasks.db"
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

func TestDefaultDBPath_AbsolutePath(t *testing.T) {
	original := os.Getenv("TASKFLOW_DIR")
	defer func() { os.Setenv("TASKFLOW_DIR", original) }()

	os.Setenv("TASKFLOW_DIR", "/opt/data")

	got := DefaultDBPath()
	want := "/opt/data/tasks.db"
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

func TestDefaultDBPath_RelativePath(t *testing.T) {
	original := os.Getenv("TASKFLOW_DIR")
	defer func() { os.Setenv("TASKFLOW_DIR", original) }()

	os.Setenv("TASKFLOW_DIR", "mydata")

	got := DefaultDBPath()
	// Should resolve to absolute path
	if got == "mydata/tasks.db" {
		t.Errorf("DefaultDBPath() = %v, expected absolute path but got relative", got)
	}
	// Should end with tasks.db
	if got[len(got)-8:] != "tasks.db" {
		t.Errorf("DefaultDBPath() = %v, expected path ending with tasks.db", got)
	}
}

func TestDefaultDBPath_TrailingSlash(t *testing.T) {
	original := os.Getenv("TASKFLOW_DIR")
	defer func() { os.Setenv("TASKFLOW_DIR", original) }()

	os.Setenv("TASKFLOW_DIR", "/tmp/work/")

	got := DefaultDBPath()
	want := "/tmp/work/tasks.db"
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}
