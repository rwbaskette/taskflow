package output

import (
	"fmt"
	"strings"

	"github.com/user/project/internal/service"
)

type XMLRenderer struct{}

func NewXMLRenderer() *XMLRenderer {
	return &XMLRenderer{}
}

func (r *XMLRenderer) Render(result *service.ListTaskResult) {
	if result == nil {
		fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		fmt.Println("<tasks/>")
		return
	}

	if len(result.Tasks) == 0 {
		fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		fmt.Println("<tasks/>")
		return
	}

	fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Println("<tasks>")
	fmt.Printf("  <pagination total=\"%d\" limit=\"%d\" offset=\"%d\" hasMore=\"%v\"/>\n",
		result.Total, result.Limit, result.Offset, result.HasMore)

	for _, task := range result.Tasks {
		fmt.Println("  <task>")
		fmt.Printf("    <id>%s</id>\n", EscapeXML(task.ID))
		fmt.Printf("    <title>%s</title>\n", EscapeXML(task.Title))
		fmt.Printf("    <description>%s</description>\n", EscapeXML(task.Description))
		fmt.Printf("    <status>%s</status>\n", EscapeXML(task.Status))
		fmt.Printf("    <milestone>%s</milestone>\n", EscapeXML(task.Milestone))
		fmt.Printf("    <sprint>%s</sprint>\n", EscapeXML(task.Sprint))
		fmt.Printf("    <actor>%s</actor>\n", EscapeXML(task.Actor))
		fmt.Printf("    <created>%s</created>\n", EscapeXML(task.Created))
		fmt.Printf("    <lastUpdated>%s</lastUpdated>\n", EscapeXML(task.LastUpdated))
		fmt.Println("  </task>")
	}

	fmt.Println("</tasks>")
}

func (r *XMLRenderer) RenderTask(task *service.TaskItem) {
	if task == nil {
		fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		fmt.Println("<task/>")
		return
	}

	fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Println("<task>")
	fmt.Printf("  <id>%s</id>\n", EscapeXML(task.ID))
	fmt.Printf("  <title>%s</title>\n", EscapeXML(task.Title))
	fmt.Printf("  <description>%s</description>\n", EscapeXML(task.Description))
	fmt.Printf("  <status>%s</status>\n", EscapeXML(task.Status))
	fmt.Printf("  <milestone>%s</milestone>\n", EscapeXML(task.Milestone))
	fmt.Printf("  <sprint>%s</sprint>\n", EscapeXML(task.Sprint))
	fmt.Printf("  <actor>%s</actor>\n", EscapeXML(task.Actor))
	fmt.Printf("  <created>%s</created>\n", EscapeXML(task.Created))
	fmt.Printf("  <lastUpdated>%s</lastUpdated>\n", EscapeXML(task.LastUpdated))
	fmt.Println("</task>")
}

func EscapeXML(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func RenderTasksXML(tasks []service.TaskItem) {
	renderer := NewXMLRenderer()
	result := &service.ListTaskResult{
		Tasks:   tasks,
		Total:   len(tasks),
		Limit:   0,
		Offset:  0,
		HasMore: false,
	}
	renderer.Render(result)
}
