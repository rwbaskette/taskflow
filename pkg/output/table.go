package output

import (
	"encoding/json"
	"fmt"

	"github.com/rwbaskette/taskflow/internal/service"
)

type TaskTableRenderer struct{}

func NewTaskTableRenderer() *TaskTableRenderer {
	return &TaskTableRenderer{}
}

func (r *TaskTableRenderer) Render(result *service.ListTaskResult) {
	if result == nil {
		fmt.Println("[]")
		return
	}

	data := struct {
		Tasks   []service.TaskItem `json:"tasks"`
		Total   int                `json:"total"`
		Limit   int                `json:"limit"`
		Offset  int                `json:"offset"`
		HasMore bool               `json:"hasMore"`
	}{
		Tasks:   result.Tasks,
		Total:   result.Total,
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("[]")
		return
	}

	fmt.Println(string(jsonData))
}

// RenderTasks renders a slice of tasks as JSON. The total is set to len(tasks)
// since this helper does not have access to the full filtered count. Callers
// that need accurate pagination metadata should use Render with a properly
// constructed ListTaskResult instead.
func RenderTasks(tasks []service.TaskItem) {
	renderer := NewTaskTableRenderer()
	result := &service.ListTaskResult{
		Tasks:   tasks,
		Total:   len(tasks),
		Limit:   0,
		Offset:  0,
		HasMore: false,
	}
	renderer.Render(result)
}