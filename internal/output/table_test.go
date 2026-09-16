package output

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/rwbaskette/taskflow/internal/service"
)

// captureStdout runs fn while capturing everything it writes to os.Stdout.
// Render writes via fmt.Println, which resolves os.Stdout on each call, so
// swapping the variable is sufficient. Tests in this package do not run in
// parallel.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	defer func() {
		os.Stdout = old
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return string(out)
}

func twoTaskResult() *service.ListTaskResult {
	return &service.ListTaskResult{
		Tasks: []service.TaskItem{
			{
				ID:          "t-1",
				Milestone:   "m1",
				Sprint:      "sprint-1",
				Title:       "First task",
				Description: "Do the first thing",
				Status:      "todo",
				Actor:       "alice",
				BlockedBy:   []string{"t-2"},
				Created:     "2026-01-02 15:04:05",
				LastUpdated: "2026-01-03 10:00:00",
			},
			{
				ID:          "t-2",
				Milestone:   "m1",
				Sprint:      "sprint-1",
				Title:       "Second task",
				Description: "",
				Status:      "blocked",
				Actor:       "",
				BlockedBy:   nil,
				Created:     "2026-01-02 16:00:00",
				LastUpdated: "2026-01-03 11:00:00",
			},
		},
		Total:   2,
		Limit:   10,
		Offset:  0,
		HasMore: false,
	}
}

// TestRender pins the exact JSON contract of Render: top-level key names,
// per-task key names (TaskItem has no json tags, so they are the Go field
// names), value types, and pagination fields.
func TestRender(t *testing.T) {
	tests := []struct {
		name   string
		result *service.ListTaskResult
		check  func(t *testing.T, out string)
	}{
		{
			name:   "nil result prints empty JSON array",
			result: nil,
			check: func(t *testing.T, out string) {
				if out != "[]\n" {
					t.Errorf("Render(nil) output = %q, want %q", out, "[]\n")
				}
			},
		},
		{
			name:   "normal task list with all fields populated",
			result: twoTaskResult(),
			check: func(t *testing.T, out string) {
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(out), &m); err != nil {
					t.Fatalf("output is not valid JSON: %v\n%s", err, out)
				}

				wantTopKeys := []string{"tasks", "total", "limit", "offset", "hasMore"}
				if len(m) != len(wantTopKeys) {
					t.Errorf("top-level key count = %d (%v), want exactly %v", len(m), keysOf(m), wantTopKeys)
				}
				for _, k := range wantTopKeys {
					if _, ok := m[k]; !ok {
						t.Errorf("missing top-level key %q", k)
					}
				}

				if got := m["total"]; got != float64(2) {
					t.Errorf("total = %v, want 2", got)
				}
				if got := m["limit"]; got != float64(10) {
					t.Errorf("limit = %v, want 10", got)
				}
				if got := m["offset"]; got != float64(0) {
					t.Errorf("offset = %v, want 0", got)
				}
				if got := m["hasMore"]; got != false {
					t.Errorf("hasMore = %v, want false", got)
				}

				tasks, ok := m["tasks"].([]interface{})
				if !ok {
					t.Fatalf("tasks is %T, want array; output:\n%s", m["tasks"], out)
				}
				if len(tasks) != 2 {
					t.Fatalf("len(tasks) = %d, want 2", len(tasks))
				}

				first, ok := tasks[0].(map[string]interface{})
				if !ok {
					t.Fatalf("tasks[0] is %T, want object", tasks[0])
				}
				// TaskItem has no json tags: keys are the Go field names.
				wantTaskKeys := []string{
					"ID", "Milestone", "Sprint", "Title", "Description",
					"Status", "Actor", "BlockedBy", "Created", "LastUpdated",
				}
				if len(first) != len(wantTaskKeys) {
					t.Errorf("tasks[0] key count = %d (%v), want exactly %v", len(first), keysOf(first), wantTaskKeys)
				}
				for _, k := range wantTaskKeys {
					if _, ok := first[k]; !ok {
						t.Errorf("tasks[0] missing key %q", k)
					}
				}

				wantValues := map[string]interface{}{
					"ID":          "t-1",
					"Milestone":   "m1",
					"Sprint":      "sprint-1",
					"Title":       "First task",
					"Description": "Do the first thing",
					"Status":      "todo",
					"Actor":       "alice",
					"Created":     "2026-01-02 15:04:05",
					"LastUpdated": "2026-01-03 10:00:00",
				}
				for k, want := range wantValues {
					if got := first[k]; got != want {
						t.Errorf("tasks[0][%q] = %v, want %v", k, got, want)
					}
				}

				blockedBy, ok := first["BlockedBy"].([]interface{})
				if !ok {
					t.Fatalf("tasks[0][BlockedBy] is %T, want array", first["BlockedBy"])
				}
				if len(blockedBy) != 1 || blockedBy[0] != "t-2" {
					t.Errorf("tasks[0][BlockedBy] = %v, want [t-2]", blockedBy)
				}

				second, ok := tasks[1].(map[string]interface{})
				if !ok {
					t.Fatalf("tasks[1] is %T, want object", tasks[1])
				}
				// nil BlockedBy marshals as JSON null (no omitempty).
				if got := second["BlockedBy"]; got != nil {
					t.Errorf("tasks[1][BlockedBy] = %v, want nil (JSON null)", got)
				}
				if got := second["Status"]; got != "blocked" {
					t.Errorf("tasks[1][Status] = %v, want blocked", got)
				}
			},
		},
		{
			name: "hasMore true when offset+page does not reach total",
			result: &service.ListTaskResult{
				Tasks:   []service.TaskItem{{ID: "t-3", Title: "Third"}},
				Total:   5,
				Limit:   2,
				Offset:  2,
				HasMore: true,
			},
			check: func(t *testing.T, out string) {
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(out), &m); err != nil {
					t.Fatalf("output is not valid JSON: %v", err)
				}
				if got := m["hasMore"]; got != true {
					t.Errorf("hasMore = %v, want true", got)
				}
				if got := m["total"]; got != float64(5) {
					t.Errorf("total = %v, want 5", got)
				}
				if got := m["limit"]; got != float64(2) {
					t.Errorf("limit = %v, want 2", got)
				}
				if got := m["offset"]; got != float64(2) {
					t.Errorf("offset = %v, want 2", got)
				}
			},
		},
		{
			name: "empty non-nil task slice renders as empty array",
			result: &service.ListTaskResult{
				Tasks: []service.TaskItem{},
				Total: 0,
			},
			check: func(t *testing.T, out string) {
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(out), &m); err != nil {
					t.Fatalf("output is not valid JSON: %v\n%s", err, out)
				}
				arr, ok := m["tasks"].([]interface{})
				if !ok {
					t.Fatalf("tasks is %T (%v), want empty array [] not null; output:\n%s", m["tasks"], m["tasks"], out)
				}
				if len(arr) != 0 {
					t.Errorf("len(tasks) = %d, want 0", len(arr))
				}
				if got := m["total"]; got != float64(0) {
					t.Errorf("total = %v, want 0", got)
				}
			},
		},
		{
			name: "nil task slice renders as null",
			result: &service.ListTaskResult{
				Tasks: nil,
				Total: 0,
			},
			check: func(t *testing.T, out string) {
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(out), &m); err != nil {
					t.Fatalf("output is not valid JSON: %v\n%s", err, out)
				}
				if got, present := m["tasks"]; !present || got != nil {
					t.Errorf("tasks = %v (present=%t), want JSON null", got, present)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				NewTaskTableRenderer().Render(tt.result)
			})
			tt.check(t, out)
		})
	}
}

// TestRenderSnapshot is the single full-output snapshot: it pins the overall
// shape, the two-space MarshalIndent style, struct-field key order, and the
// trailing newline added by Println.
func TestRenderSnapshot(t *testing.T) {
	want := `{
  "tasks": [
    {
      "ID": "t-1",
      "Milestone": "m1",
      "Sprint": "sprint-1",
      "Title": "First task",
      "Description": "Do the first thing",
      "Status": "todo",
      "Actor": "alice",
      "BlockedBy": [
        "t-2"
      ],
      "Created": "2026-01-02 15:04:05",
      "LastUpdated": "2026-01-03 10:00:00"
    },
    {
      "ID": "t-2",
      "Milestone": "m1",
      "Sprint": "sprint-1",
      "Title": "Second task",
      "Description": "",
      "Status": "blocked",
      "Actor": "",
      "BlockedBy": null,
      "Created": "2026-01-02 16:00:00",
      "LastUpdated": "2026-01-03 11:00:00"
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0,
  "hasMore": false
}
`

	out := captureStdout(t, func() {
		NewTaskTableRenderer().Render(twoTaskResult())
	})
	if out != want {
		t.Errorf("Render output mismatch:\n got: %q\nwant: %q", out, want)
	}

	// Determinism: rendering the same input twice must produce byte-identical output.
	out2 := captureStdout(t, func() {
		NewTaskTableRenderer().Render(twoTaskResult())
	})
	if out2 != out {
		t.Error("Render is not deterministic: two renders of the same input differ")
	}
}

// TestRenderTasks pins the helper's fixed pagination metadata: total is
// len(tasks), limit/offset are 0, hasMore is false.
func TestRenderTasks(t *testing.T) {
	tests := []struct {
		name  string
		tasks []service.TaskItem
		check func(t *testing.T, m map[string]interface{})
	}{
		{
			name: "one task gets total=1, zero pagination",
			tasks: []service.TaskItem{
				{ID: "t-1", Title: "Only task", Status: "todo"},
			},
			check: func(t *testing.T, m map[string]interface{}) {
				if got := m["total"]; got != float64(1) {
					t.Errorf("total = %v, want 1", got)
				}
				if got := m["limit"]; got != float64(0) {
					t.Errorf("limit = %v, want 0", got)
				}
				if got := m["offset"]; got != float64(0) {
					t.Errorf("offset = %v, want 0", got)
				}
				if got := m["hasMore"]; got != false {
					t.Errorf("hasMore = %v, want false", got)
				}
				tasks, ok := m["tasks"].([]interface{})
				if !ok || len(tasks) != 1 {
					t.Fatalf("tasks = %v, want array of length 1", m["tasks"])
				}
			},
		},
		{
			name:  "nil task slice gets total=0 and null tasks",
			tasks: nil,
			check: func(t *testing.T, m map[string]interface{}) {
				if got := m["total"]; got != float64(0) {
					t.Errorf("total = %v, want 0", got)
				}
				if got := m["tasks"]; got != nil {
					t.Errorf("tasks = %v, want JSON null", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				RenderTasks(tt.tasks)
			})
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(out), &m); err != nil {
				t.Fatalf("output is not valid JSON: %v\n%s", err, out)
			}
			tt.check(t, m)
		})
	}
}

func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
