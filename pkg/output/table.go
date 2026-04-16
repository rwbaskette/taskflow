package output

import (
	"fmt"
	"strings"

	"github.com/user/project/internal/service"
)

// OutputFormat defines the output format type
type OutputFormat string

const (
	FormatTable    OutputFormat = "table"
	FormatMarkdown OutputFormat = "markdown"
	FormatXML      OutputFormat = "xml"
)

// TaskTableRenderer renders tasks in table or markdown format
type TaskTableRenderer struct {
	format OutputFormat
}

// NewTaskTableRenderer creates a new table renderer
func NewTaskTableRenderer(format OutputFormat) *TaskTableRenderer {
	if format == "" {
		format = FormatTable
	}
	return &TaskTableRenderer{
		format: format,
	}
}

// Render renders the task list result
func (r *TaskTableRenderer) Render(result *service.ListTaskResult) {
	if result == nil {
		fmt.Println("No tasks found.")
		return
	}

	if len(result.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	switch r.format {
	case FormatMarkdown:
		r.renderMarkdown(result)
	case FormatXML:
		r.renderXML(result)
	default:
		r.renderTable(result)
	}
}

// renderTable renders tasks in table format
func (r *TaskTableRenderer) renderTable(result *service.ListTaskResult) {
	fmt.Println()
	fmt.Printf("%s%-10s %-30s %-12s %-20s %-10s %-15s %-19s %-19s%s\n",
		ColorBold, "ID", "Title", "Status", "Milestone", "Sprint", "Actor", "Created", "Last Updated", ColorReset)
	fmt.Println(strings.Repeat("-", 140))

	for _, task := range result.Tasks {
		// Truncate long titles
		title := task.Title
		if len(title) > 27 {
			title = title[:27] + "..."
		}

		// Set milestone to "-" if empty
		milestone := task.Milestone
		if milestone == "" {
			milestone = "-"
		}

		// Truncate milestone if too long
		if len(milestone) > 19 {
			milestone = milestone[:19] + "..."
		}

		// Set sprint to "-" if empty
		sprint := task.Sprint
		if sprint == "" {
			sprint = "-"
		}

		// Set actor to "-" if empty
		actor := task.Actor
		if actor == "" {
			actor = "-"
		}

		// Truncate actor if too long
		if len(actor) > 14 {
			actor = actor[:14] + "..."
		}

		// Color status based on value
		statusColor := ColorReset
		switch task.Status {
		case "done":
			statusColor = ColorGreen
		case "in_progress":
			statusColor = ColorYellow
		case "blocked":
			statusColor = ColorRed
		default:
			statusColor = ColorBlue
		}

		fmt.Printf("%-10s %-30s %s%-12s%s %-20s %-10s %-15s %-19s %-19s\n",
			task.ID,
			title,
			statusColor,
			task.Status,
			ColorReset,
			milestone,
			sprint,
			actor,
			task.Created,
			task.LastUpdated,
		)
	}

	fmt.Println()

	// Print pagination info
	r.renderPaginationInfo(result)
}

// renderMarkdown renders tasks in markdown format
func (r *TaskTableRenderer) renderMarkdown(result *service.ListTaskResult) {
	fmt.Println()

	// Print header
	fmt.Println("| ID | Title | Status | Milestone | Sprint | Actor | Created | Last Updated |")
	fmt.Println("|---|------|--------|-----------|--------|-------|-------------|---------------|")

	// Print rows
	for _, task := range result.Tasks {
		// Escape pipes in title
		title := strings.ReplaceAll(task.Title, "|", "\\|")

		// Truncate long titles
		if len(title) > 30 {
			title = title[:27] + "..."
		}

		// Set milestone to "-" if empty
		milestone := task.Milestone
		if milestone == "" {
			milestone = "-"
		}

		// Set sprint to "-" if empty
		sprint := task.Sprint
		if sprint == "" {
			sprint = "-"
		}

		// Set actor to "-" if empty
		actor := task.Actor
		if actor == "" {
			actor = "-"
		}

		// Format status with emoji indicator
		status := task.Status
		switch task.Status {
		case "done":
			status = "✅ " + task.Status
		case "in_progress":
			status = "🔄 " + task.Status
		case "blocked":
			status = "🚫 " + task.Status
		default:
			status = "📋 " + task.Status
		}

		fmt.Printf("| %s | %s | %s | %s | %s | %s | %s | %s |\n",
			task.ID,
			title,
			status,
			milestone,
			sprint,
			actor,
			task.Created,
			task.LastUpdated,
		)
	}

	fmt.Println()

	// Print pagination info
	r.renderMarkdownPaginationInfo(result)
}

// renderPaginationInfo prints pagination information for table format
func (r *TaskTableRenderer) renderPaginationInfo(result *service.ListTaskResult) {
	if result.Limit > 0 {
		start := result.Offset + 1
		end := result.Offset + len(result.Tasks)
		if result.Limit > len(result.Tasks) {
			end = result.Offset + len(result.Tasks)
		}

		if result.HasMore {
			fmt.Printf("Showing %d-%d of tasks (page %d). Use --limit to see more.\n",
				start, end, (result.Offset/result.Limit)+1)
		} else {
			fmt.Printf("Showing %d-%d of %d total tasks.\n",
				start, end, result.Total)
		}
	} else {
		fmt.Printf("Total: %d tasks\n", result.Total)
	}
}

// renderMarkdownPaginationInfo prints pagination information for markdown format
func (r *TaskTableRenderer) renderMarkdownPaginationInfo(result *service.ListTaskResult) {
	fmt.Println("**Pagination:**")
	if result.Limit > 0 {
		start := result.Offset + 1
		end := result.Offset + len(result.Tasks)
		if result.Limit > len(result.Tasks) {
			end = result.Offset + len(result.Tasks)
		}

		if result.HasMore {
			fmt.Printf("- Showing %d-%d (page %d)\n", start, end, (result.Offset/result.Limit)+1)
			fmt.Println("- Use `--limit` flag for more results")
		} else {
			fmt.Printf("- Showing %d-%d of %d total\n", start, end, result.Total)
		}
	} else {
		fmt.Printf("- Total: %d tasks\n", result.Total)
	}
}

// renderXML renders tasks in XML format
func (r *TaskTableRenderer) renderXML(result *service.ListTaskResult) {
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

// RenderTasks is a convenience function to render tasks with default format
func RenderTasks(tasks []service.TaskItem, format OutputFormat) {
	renderer := NewTaskTableRenderer(format)
	result := &service.ListTaskResult{
		Tasks:   tasks,
		Total:   len(tasks),
		Limit:   0,
		Offset:  0,
		HasMore: false,
	}
	renderer.Render(result)
}

// ParseOutputFormat parses the format string and returns OutputFormat
func ParseOutputFormat(formatStr string) (OutputFormat, error) {
	formatStr = strings.ToLower(strings.TrimSpace(formatStr))

	switch formatStr {
	case "table":
		return FormatTable, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	case "xml":
		return FormatXML, nil
	default:
		return "", fmt.Errorf("invalid output format: %s (valid options: table, markdown, xml)", formatStr)
	}
}

// GetValidFormats returns a list of valid format options
func GetValidFormats() []string {
	return []string{"table", "markdown", "xml"}
}
