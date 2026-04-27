package db

import (
	"fmt"
	"regexp"
	"strings"
)

// Valid SQL metacharacters that could be used for injection
// Matches: semicolons, double dashes (SQL comment), block comment markers, script tags
var sqlMetacharacterPattern = regexp.MustCompile(`;|--|\/\*|\*\/|<script|`)

// Valid table columns allowlist for validation
var validTaskColumns = map[string]bool{
	"id":           true,
	"milestone":    true,
	"title":        true,
	"description":  true,
	"status":       true,
	"actor":        true,
	"last_updated": true,
}

// validTableNames is the allowlist for valid table names
var validTableNames = map[string]bool{
	"tasks": true,
}

// FilterOp represents a filter operator
type FilterOp string

const (
	// FilterOpEq represents equality comparison
	FilterOpEq FilterOp = "="
	// FilterOpNe represents not equal comparison
	FilterOpNe FilterOp = "!="
	// FilterOpLike represents LIKE pattern matching
	FilterOpLike FilterOp = "LIKE"
	// FilterOpIn represents IN clause
	FilterOpIn FilterOp = "IN"
	// FilterOpNotIn represents NOT IN clause
	FilterOpNotIn FilterOp = "NOT IN"
)

// FilterLogic represents the logical operator between filters
type FilterLogic string

const (
	// FilterLogicAnd represents AND logic
	FilterLogicAnd FilterLogic = "AND"
	// FilterLogicOr represents OR logic
	FilterLogicOr FilterLogic = "OR"
)

// FilterField represents the field to filter on
type FilterField string

const (
	// FilterFieldMilestone represents the milestone field
	FilterFieldMilestone FilterField = "milestone"
	// FilterFieldStatus represents the status field
	FilterFieldStatus FilterField = "status"
	// FilterFieldActor represents the actor field
	FilterFieldActor FilterField = "actor"
	// FilterFieldTitle represents the title field
	FilterFieldTitle FilterField = "title"
	// FilterFieldID represents the id field
	FilterFieldID FilterField = "id"
	// FilterFieldDescription represents the description field
	FilterFieldDescription FilterField = "description"
)

// Condition represents a single filter condition
type Condition struct {
	Field    FilterField
	Operator FilterOp
	Value    interface{}
}

// Filter represents a composable filter with conditions and logical operators
type Filter struct {
	Conditions []Condition
	Logic      FilterLogic
	SubFilters []*Filter
}

// NewFilter creates a new Filter with AND logic by default
func NewFilter() *Filter {
	return &Filter{
		Conditions: []Condition{},
		Logic:      FilterLogicAnd,
		SubFilters: []*Filter{},
	}
}

// WithCondition adds a condition to the filter
func (f *Filter) WithCondition(field FilterField, op FilterOp, value interface{}) *Filter {
	f.Conditions = append(f.Conditions, Condition{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return f
}

// WithAnd sets the logical operator to AND
func (f *Filter) WithAnd() *Filter {
	f.Logic = FilterLogicAnd
	return f
}

// WithOr sets the logical operator to OR
func (f *Filter) WithOr() *Filter {
	f.Logic = FilterLogicOr
	return f
}

// WithSubFilter adds a sub-filter (for nested conditions)
func (f *Filter) WithSubFilter(sub *Filter) *Filter {
	f.SubFilters = append(f.SubFilters, sub)
	return f
}

// MilestoneEq creates a filter for milestone equality
func MilestoneEq(value string) *Filter {
	return NewFilter().WithCondition(FilterFieldMilestone, FilterOpEq, value)
}

// MilestoneNe creates a filter for milestone inequality
func MilestoneNe(value string) *Filter {
	return NewFilter().WithCondition(FilterFieldMilestone, FilterOpNe, value)
}

// MilestoneIn creates a filter for milestone IN clause
func MilestoneIn(values ...string) *Filter {
	if len(values) == 1 {
		return NewFilter().WithCondition(FilterFieldMilestone, FilterOpEq, values[0])
	}
	return NewFilter().WithCondition(FilterFieldMilestone, FilterOpIn, values)
}

// StatusEq creates a filter for status equality
func StatusEq(value string) *Filter {
	return NewFilter().WithCondition(FilterFieldStatus, FilterOpEq, value)
}

// StatusNe creates a filter for status inequality
func StatusNe(value string) *Filter {
	return NewFilter().WithCondition(FilterFieldStatus, FilterOpNe, value)
}

// StatusIn creates a filter for status IN clause
func StatusIn(values ...string) *Filter {
	if len(values) == 1 {
		return NewFilter().WithCondition(FilterFieldStatus, FilterOpEq, values[0])
	}
	return NewFilter().WithCondition(FilterFieldStatus, FilterOpIn, values)
}

// ActorEq creates a filter for actor equality
func ActorEq(value string) *Filter {
	return NewFilter().WithCondition(FilterFieldActor, FilterOpEq, value)
}

// ActorNe creates a filter for actor inequality
func ActorNe(value string) *Filter {
	return NewFilter().WithCondition(FilterFieldActor, FilterOpNe, value)
}

// ActorIn creates a filter for actor IN clause
func ActorIn(values ...string) *Filter {
	if len(values) == 1 {
		return NewFilter().WithCondition(FilterFieldActor, FilterOpEq, values[0])
	}
	return NewFilter().WithCondition(FilterFieldActor, FilterOpIn, values)
}

// TitleLike creates a filter for title pattern matching
func TitleLike(pattern string) *Filter {
	return NewFilter().WithCondition(FilterFieldTitle, FilterOpLike, "%"+pattern+"%")
}

// IDIn creates a filter for id IN clause
func IDIn(values ...string) *Filter {
	if len(values) == 1 {
		return NewFilter().WithCondition(FilterFieldID, FilterOpEq, values[0])
	}
	return NewFilter().WithCondition(FilterFieldID, FilterOpIn, values)
}

// CombineFilters combines multiple filters with a logical operator
func CombineFilters(logic FilterLogic, filters ...*Filter) *Filter {
	combined := NewFilter().WithAnd()
	if logic == FilterLogicOr {
		combined.WithOr()
	}
	for _, f := range filters {
		if f != nil {
			combined.WithSubFilter(f)
		}
	}
	return combined
}

// And combines filters with AND logic
func And(filters ...*Filter) *Filter {
	return CombineFilters(FilterLogicAnd, filters...)
}

// Or combines filters with OR logic
func Or(filters ...*Filter) *Filter {
	return CombineFilters(FilterLogicOr, filters...)
}

// FromTaskFilter converts a TaskFilter to a composable Filter
func FromTaskFilter(tf TaskFilter) *Filter {
	filter := NewFilter()

	if tf.Milestone != "" {
		filter.WithCondition(FilterFieldMilestone, FilterOpEq, tf.Milestone)
	}

	if tf.Status != "" {
		filter.WithCondition(FilterFieldStatus, FilterOpEq, tf.Status)
	}

	if tf.Actor != "" {
		filter.WithCondition(FilterFieldActor, FilterOpEq, tf.Actor)
	}

	return filter
}

// validateColumnName validates a column name against the allowlist
func validateColumnName(col string) error {
	if col == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	// First check against allowlist (fast path for valid columns)
	if validTaskColumns[col] {
		return nil
	}

	// If not in allowlist, check for SQL injection patterns
	if sqlMetacharacterPattern.MatchString(col) {
		return fmt.Errorf("column name contains invalid characters: %s", col)
	}

	return fmt.Errorf("invalid column name: %s", col)
}

// validateTableName validates a table name against the allowlist
func validateTableName(table string) error {
	if table == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// First check against allowlist (fast path for valid tables)
	if validTableNames[table] {
		return nil
	}

	// If not in allowlist, check for SQL injection patterns
	if sqlMetacharacterPattern.MatchString(table) {
		return fmt.Errorf("table name contains invalid characters: %s", table)
	}

	return fmt.Errorf("invalid table name: %s", table)
}

// validateJoinClause validates a raw JOIN SQL string for SQL injection
func validateJoinClause(join string) error {
	if join == "" {
		return fmt.Errorf("join clause cannot be empty")
	}

	// Reject if contains semicolons (statement separation)
	if strings.Contains(join, ";") {
		return fmt.Errorf("join clause contains invalid character (semicolon): %s", join)
	}

	// Reject if contains SQL comment markers
	if strings.Contains(join, "--") || strings.Contains(join, "/*") || strings.Contains(join, "*/") {
		return fmt.Errorf("join clause contains SQL comment markers: %s", join)
	}

	return nil
}

// normalizeFieldName converts a FilterField to its SQL column name
func normalizeFieldName(field FilterField) string {
	return string(field)
}

// parseValue converts a value for use in SQL parameters
func parseValue(op FilterOp, value interface{}) (interface{}, bool) {
	switch op {
	case FilterOpLike:
		if v, ok := value.(string); ok {
			return v, true
		}
		return value, true
	case FilterOpIn, FilterOpNotIn:
		if values, ok := value.([]string); ok && len(values) > 0 {
			// Return as is for IN clause handling
			return values, true
		}
		if slice, ok := value.([]interface{}); ok && len(slice) > 0 {
			return slice, true
		}
		return nil, false
	default:
		return value, true
	}
}

// buildConditionSQL builds the SQL for a single condition
func buildConditionSQL(cond Condition, params *[]interface{}) (string, bool) {
	// Handle slice values for IN/NOT IN
	if cond.Operator == FilterOpIn || cond.Operator == FilterOpNotIn {
		if values, ok := cond.Value.([]string); ok && len(values) > 0 {
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = "?"
				*params = append(*params, values[i])
			}
			fieldName := normalizeFieldName(cond.Field)
			return fieldName + " " + string(cond.Operator) + " (" + strings.Join(placeholders, ", ") + ")", true
		}
		if values, ok := cond.Value.([]interface{}); ok && len(values) > 0 {
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = "?"
				*params = append(*params, values[i])
			}
			fieldName := normalizeFieldName(cond.Field)
			return fieldName + " " + string(cond.Operator) + " (" + strings.Join(placeholders, ", ") + ")", true
		}
		return "", false
	}

	// Handle LIKE operator - wrap value with % for pattern matching
	if cond.Operator == FilterOpLike {
		if v, ok := cond.Value.(string); ok {
			*params = append(*params, v)
			fieldName := normalizeFieldName(cond.Field)
			return fieldName + " " + string(cond.Operator) + " ?", true
		}
	}

	// Handle regular operators
	parsedValue, ok := parseValue(cond.Operator, cond.Value)
	if !ok {
		return "", false
	}

	*params = append(*params, parsedValue)
	fieldName := normalizeFieldName(cond.Field)
	return fieldName + " " + string(cond.Operator) + " ?", true
}
