package db

import (
	"strings"
	"testing"
)

func TestNewQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()

	if qb == nil {
		t.Fatal("NewQueryBuilder returned nil")
	}

	if qb.tableName != "tasks" {
		t.Errorf("expected table 'tasks', got '%s'", qb.tableName)
	}

	expectedCols := []string{"id", "milestone", "title", "description", "status", "actor", "last_updated"}
	if len(qb.selectCols) != len(expectedCols) {
		t.Errorf("expected %d columns, got %d", len(expectedCols), len(qb.selectCols))
	}

	if qb.orderBy != "last_updated" {
		t.Errorf("expected orderBy 'last_updated', got '%s'", qb.orderBy)
	}

	if !qb.orderDesc {
		t.Error("expected orderDesc to be true by default")
	}

	if len(qb.filters) != 0 {
		t.Errorf("expected empty filters, got %d", len(qb.filters))
	}
}

func TestQueryBuilderForTable(t *testing.T) {
	qb := NewQueryBuilder().ForTable("tasks")

	if qb.tableName != "tasks" {
		t.Errorf("expected table 'tasks', got '%s'", qb.tableName)
	}

	// Test invalid table name
	qb = NewQueryBuilder().ForTable("invalid_table")
	// The table should still be set but validation happens at Build time
	if qb.tableName != "invalid_table" {
		t.Errorf("expected table 'invalid_table', got '%s'", qb.tableName)
	}
}

func TestQueryBuilderSelect(t *testing.T) {
	qb := NewQueryBuilder().Select("id", "title")

	if len(qb.selectCols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(qb.selectCols))
	}

	// Test validation - invalid columns should be skipped
	qb = NewQueryBuilder().Select("id", "invalid_col", "title")
	// At least valid columns should be present
	if len(qb.selectCols) < 2 {
		t.Errorf("expected at least 2 valid columns, got %d", len(qb.selectCols))
	}
}

func TestQueryBuilderWhere(t *testing.T) {
	qb := NewQueryBuilder()
	filter := StatusEq("open")
	qb.Where(filter)

	if len(qb.filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(qb.filters))
	}

	// Test nil filter
	qb.Where(nil)
	if len(qb.filters) != 1 {
		t.Errorf("expected 1 filter after nil, got %d", len(qb.filters))
	}
}

func TestQueryBuilderOrderBy(t *testing.T) {
	qb := NewQueryBuilder().OrderBy("id", true)

	if qb.orderBy != "id" {
		t.Errorf("expected orderBy 'id', got '%s'", qb.orderBy)
	}

	if !qb.orderDesc {
		t.Error("expected orderDesc to be true")
	}

	// Test with DESC = false
	qb = NewQueryBuilder().OrderBy("title", false)
	if qb.orderDesc {
		t.Error("expected orderDesc to be false")
	}

	// Test invalid column - should fallback to default
	qb = NewQueryBuilder().OrderBy("invalid_col", true)
	if qb.orderBy != "last_updated" {
		t.Errorf("expected default orderBy after invalid, got '%s'", qb.orderBy)
	}
}

func TestQueryBuilderLimit(t *testing.T) {
	qb := NewQueryBuilder().Limit(10)

	if qb.limit != 10 {
		t.Errorf("expected limit 10, got %d", qb.limit)
	}
}

func TestQueryBuilderOffset(t *testing.T) {
	qb := NewQueryBuilder().Offset(20)

	if qb.offset != 20 {
		t.Errorf("expected offset 20, got %d", qb.offset)
	}
}

func TestQueryBuilderGroupBy(t *testing.T) {
	qb := NewQueryBuilder().GroupBy("status", "milestone")

	if len(qb.groupBy) != 2 {
		t.Errorf("expected 2 group by columns, got %d", len(qb.groupBy))
	}

	// Test validation - invalid columns should be skipped
	qb = NewQueryBuilder().GroupBy("status", "invalid_col", "milestone")
	if len(qb.groupBy) != 2 {
		t.Errorf("expected 2 valid group by columns, got %d", len(qb.groupBy))
	}
}

func TestQueryBuilderJoin(t *testing.T) {
	qb := NewQueryBuilder().Join("JOIN users ON tasks.user_id = users.id")

	if len(qb.joins) != 1 {
		t.Errorf("expected 1 join, got %d", len(qb.joins))
	}

	// Test invalid join - should be skipped
	qb = NewQueryBuilder().Join("JOIN users ON tasks.id = users.id; DROP TABLE tasks--")
	if len(qb.joins) != 0 {
		t.Errorf("expected 0 joins after invalid, got %d", len(qb.joins))
	}
}

func TestQueryBuilderBuild(t *testing.T) {
	qb := NewQueryBuilder()
	query := qb.Build()

	if query == nil {
		t.Fatal("Build returned nil query")
	}

	if query.SQL == "" {
		t.Error("Build returned empty SQL")
	}

	if !strings.Contains(query.SQL, "SELECT") {
		t.Error("Build SQL should contain SELECT")
	}

	if !strings.Contains(query.SQL, "FROM tasks") {
		t.Error("Build SQL should contain FROM tasks")
	}

	// Test with filters
	qb = NewQueryBuilder().Where(StatusEq("open"))
	query = qb.Build()

	if !strings.Contains(query.SQL, "WHERE") {
		t.Error("Build SQL should contain WHERE with filters")
	}

	// Test with limit
	qb = NewQueryBuilder().Limit(10)
	query = qb.Build()

	if !strings.Contains(query.SQL, "LIMIT") {
		t.Error("Build SQL should contain LIMIT")
	}

	// Test with offset
	qb = NewQueryBuilder().Offset(20)
	query = qb.Build()

	if !strings.Contains(query.SQL, "OFFSET") {
		t.Error("Build SQL should contain OFFSET")
	}

	// Test with group by
	qb = NewQueryBuilder().GroupBy("status")
	query = qb.Build()

	if !strings.Contains(query.SQL, "GROUP BY") {
		t.Error("Build SQL should contain GROUP BY")
	}
}

func TestQueryBuilderBuild_SQLInjectionPrevention(t *testing.T) {
	// Test that SQL injection attempts are caught
	qb := NewQueryBuilder().
		ForTable("tasks; DROP TABLE tasks--").
		Select("id", "title; DELETE FROM tasks--").
		OrderBy("id; DROP TABLE tasks--", false).
		GroupBy("id; DROP TABLE tasks--")

	query := qb.Build()

	// Should return empty SQL due to validation failure
	if query.SQL != "" {
		t.Error("Build should return empty SQL for invalid inputs")
	}
}

func TestQueryBuilderBuild_WithValidColumns(t *testing.T) {
	qb := NewQueryBuilder().
		Select("id", "title", "status").
		OrderBy("id", true).
		GroupBy("status")

	query := qb.Build()

	if !strings.Contains(query.SQL, "SELECT id, title, status") {
		t.Errorf("Expected SELECT with valid columns, got: %s", query.SQL)
	}

	if !strings.Contains(query.SQL, "ORDER BY id DESC") {
		t.Errorf("Expected ORDER BY id DESC, got: %s", query.SQL)
	}

	if !strings.Contains(query.SQL, "GROUP BY status") {
		t.Errorf("Expected GROUP BY status, got: %s", query.SQL)
	}
}

func TestBuildFilterSQL(t *testing.T) {
	params := []interface{}{}

	// Test single condition with AND
	f := StatusEq("open")
	sql, ok := buildFilterSQL(f, &params)
	if !ok {
		t.Error("buildFilterSQL returned false for valid filter")
	}
	if !strings.Contains(sql, "status =") {
		t.Errorf("expected 'status =', got '%s'", sql)
	}

	// Test OR logic with multiple conditions
	f = NewFilter().WithOr().
		WithCondition(FilterFieldStatus, FilterOpEq, "open").
		WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0")
	sql, ok = buildFilterSQL(f, &params)
	if !ok {
		t.Error("buildFilterSQL returned false for OR filter")
	}
	if !strings.Contains(sql, " OR ") {
		t.Errorf("expected ' OR ', got '%s'", sql)
	}

	// Test nested sub-filters
	f = NewFilter().
		WithCondition(FilterFieldStatus, FilterOpEq, "open").
		WithSubFilter(NewFilter().WithCondition(FilterFieldActor, FilterOpEq, "user1"))
	params = []interface{}{}
	sql, ok = buildFilterSQL(f, &params)
	if !ok {
		t.Error("buildFilterSQL returned false for nested filter")
	}
	if !strings.Contains(sql, "(") || !strings.Contains(sql, ")") {
		t.Errorf("expected sub-filter in parentheses, got '%s'", sql)
	}

	// Test nil filter
	sql, ok = buildFilterSQL(nil, &params)
	if ok {
		t.Error("buildFilterSQL should return false for nil filter")
	}

	// Test empty filter
	f = NewFilter()
	sql, ok = buildFilterSQL(f, &params)
	if ok {
		t.Error("buildFilterSQL should return false for empty filter")
	}
}

func TestBuildTaskQuery(t *testing.T) {
	// Test with empty filter
	tf := TaskFilter{}
	query := BuildTaskQuery(tf)

	if query == nil {
		t.Fatal("BuildTaskQuery returned nil")
	}

	if !strings.Contains(query.SQL, "SELECT") {
		t.Error("BuildTaskQuery SQL should contain SELECT")
	}

	// Test with filters
	tf = TaskFilter{
		Status:    "open",
		Milestone: "v1.0",
		Actor:     "user1",
		Limit:     10,
		Offset:    5,
	}
	query = BuildTaskQuery(tf)

	if !strings.Contains(query.SQL, "WHERE") {
		t.Error("BuildTaskQuery SQL should contain WHERE")
	}

	if !strings.Contains(query.SQL, "LIMIT") {
		t.Error("BuildTaskQuery SQL should contain LIMIT")
	}

	if !strings.Contains(query.SQL, "OFFSET") {
		t.Error("BuildTaskQuery SQL should contain OFFSET")
	}
}

func TestBuildCountQuery(t *testing.T) {
	// Test with no filters
	qb := NewQueryBuilder()
	query := qb.BuildCountQuery()

	if !strings.Contains(query.SQL, "SELECT COUNT(*) FROM tasks") {
		t.Errorf("expected 'SELECT COUNT(*) FROM tasks', got '%s'", query.SQL)
	}

	// Test with filters - should use OR logic
	qb = NewQueryBuilder().Where(StatusEq("open").WithOr())
	query = qb.BuildCountQuery()

	if !strings.Contains(query.SQL, "WHERE") {
		t.Error("BuildCountQuery SQL should contain WHERE")
	}

	// Test with AND logic filter
	qb = NewQueryBuilder().Where(StatusEq("open").WithAnd())
	query = qb.BuildCountQuery()

	// Should contain the condition
	if !strings.Contains(query.SQL, "status =") {
		t.Errorf("expected 'status =' in SQL, got '%s'", query.SQL)
	}
}

func TestBuildExistsQuery(t *testing.T) {
	// Test with no filters
	qb := NewQueryBuilder()
	query := qb.BuildExistsQuery()

	if !strings.Contains(query.SQL, "SELECT EXISTS(SELECT 1 FROM tasks") {
		t.Errorf("expected 'SELECT EXISTS(SELECT 1 FROM tasks', got '%s'", query.SQL)
	}

	// Test with filters - should use OR logic
	qb = NewQueryBuilder().Where(StatusEq("open").WithOr())
	query = qb.BuildExistsQuery()

	if !strings.Contains(query.SQL, "WHERE") {
		t.Error("BuildExistsQuery SQL should contain WHERE")
	}

	// Test with AND logic filter
	qb = NewQueryBuilder().Where(StatusEq("open").WithAnd())
	query = qb.BuildExistsQuery()

	// Should contain the condition
	if !strings.Contains(query.SQL, "status =") {
		t.Errorf("expected 'status =' in SQL, got '%s'", query.SQL)
	}
}

func TestQueryToSQL(t *testing.T) {
	// Test with string parameter
	q := &Query{
		SQL:    "SELECT * FROM tasks WHERE status = ?",
		Params: []interface{}{"open"},
	}

	result := q.ToSQL()
	if !strings.Contains(result, "'open'") {
		t.Errorf("expected interpolated 'open', got '%s'", result)
	}

	// Test with int parameter
	q = &Query{
		SQL:    "SELECT * FROM tasks WHERE id = ?",
		Params: []interface{}{123},
	}

	result = q.ToSQL()
	if !strings.Contains(result, "123") {
		t.Errorf("expected interpolated 123, got '%s'", result)
	}

	// Test with nil query
	q = nil
	result = q.ToSQL()
	if result != "" {
		t.Errorf("expected empty string for nil query, got '%s'", result)
	}

	// Test with empty params
	q = &Query{
		SQL:    "SELECT * FROM tasks",
		Params: []interface{}{},
	}

	result = q.ToSQL()
	if result != "SELECT * FROM tasks" {
		t.Errorf("expected 'SELECT * FROM tasks', got '%s'", result)
	}
}

func TestQueryBuilderBuild_FullQuery(t *testing.T) {
	qb := NewQueryBuilder().
		Select("id", "title", "status").
		Where(StatusEq("open")).
		OrderBy("id", true).
		Limit(10).
		Offset(5).
		GroupBy("status")

	query := qb.Build()

	// Verify all parts are present
	if !strings.Contains(query.SQL, "SELECT id, title, status") {
		t.Error("Missing SELECT clause")
	}
	if !strings.Contains(query.SQL, "FROM tasks") {
		t.Error("Missing FROM clause")
	}
	if !strings.Contains(query.SQL, "WHERE") {
		t.Error("Missing WHERE clause")
	}
	if !strings.Contains(query.SQL, "ORDER BY id DESC") {
		t.Error("Missing ORDER BY clause")
	}
	if !strings.Contains(query.SQL, "LIMIT") {
		t.Error("Missing LIMIT clause")
	}
	if !strings.Contains(query.SQL, "OFFSET") {
		t.Error("Missing OFFSET clause")
	}
	if !strings.Contains(query.SQL, "GROUP BY status") {
		t.Error("Missing GROUP BY clause")
	}

	// Verify parameters - filter adds 1 param (status), plus limit and offset
	if len(query.Params) != 3 {
		t.Errorf("Expected 3 params (status, limit, offset), got %d", len(query.Params))
	}
}

func TestQueryBuilderBuild_EmptyConditions(t *testing.T) {
	// Test with no conditions - should have 1=1
	qb := NewQueryBuilder()
	query := qb.Build()

	if !strings.Contains(query.SQL, "WHERE 1=1") {
		t.Error("Expected WHERE 1=1 for empty conditions")
	}
}

func TestQueryBuilderBuild_WithJoin(t *testing.T) {
	qb := NewQueryBuilder().
		Join("JOIN users ON tasks.user_id = users.id").
		Where(StatusEq("open"))

	query := qb.Build()

	if !strings.Contains(query.SQL, "JOIN users ON tasks.user_id = users.id") {
		t.Error("Missing JOIN clause")
	}

	if !strings.Contains(query.SQL, "WHERE") {
		t.Error("Missing WHERE clause")
	}
}

func TestFilterLogicRespected(t *testing.T) {
	// Test that AND logic is respected in Build
	qb := NewQueryBuilder().
		Where(StatusEq("open").WithAnd().WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0"))

	query := qb.Build()

	// Should contain AND
	if !strings.Contains(query.SQL, " AND ") {
		t.Errorf("Expected AND in SQL, got: %s", query.SQL)
	}

	// Test that OR logic is respected in Build
	qb = NewQueryBuilder().
		Where(StatusEq("open").WithOr().WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0"))

	query = qb.Build()

	// Should contain OR
	if !strings.Contains(query.SQL, " OR ") {
		t.Errorf("Expected OR in SQL, got: %s", query.SQL)
	}
}

func TestBuildCountQuery_LogicRespected(t *testing.T) {
	// Test BuildCountQuery respects filter's Logic field
	// Currently it uses OR logic by default, should respect AND when set
	qb := NewQueryBuilder().
		Where(StatusEq("open").WithAnd().WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0"))

	query := qb.BuildCountQuery()

	// Should use the filter's logic (AND)
	if !strings.Contains(query.SQL, " AND ") {
		t.Errorf("Expected AND in count query SQL, got: %s", query.SQL)
	}

	// Test with OR logic
	qb = NewQueryBuilder().
		Where(StatusEq("open").WithOr().WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0"))

	query = qb.BuildCountQuery()

	// Should use OR logic
	if !strings.Contains(query.SQL, " OR ") {
		t.Errorf("Expected OR in count query SQL, got: %s", query.SQL)
	}
}

func TestBuildExistsQuery_LogicRespected(t *testing.T) {
	// Test BuildExistsQuery respects filter's Logic field
	qb := NewQueryBuilder().
		Where(StatusEq("open").WithAnd().WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0"))

	query := qb.BuildExistsQuery()

	// Should use the filter's logic (AND)
	if !strings.Contains(query.SQL, " AND ") {
		t.Errorf("Expected AND in exists query SQL, got: %s", query.SQL)
	}

	// Test with OR logic
	qb = NewQueryBuilder().
		Where(StatusEq("open").WithOr().WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0"))

	query = qb.BuildExistsQuery()

	// Should use OR logic
	if !strings.Contains(query.SQL, " OR ") {
		t.Errorf("Expected OR in exists query SQL, got: %s", query.SQL)
	}
}

func TestMultipleFilters(t *testing.T) {
	// Test multiple filters with different logic
	qb := NewQueryBuilder().
		Where(StatusEq("open")).
		Where(ActorEq("user1"))

	query := qb.Build()

	// Should have both conditions separated by space (default AND)
	if !strings.Contains(query.SQL, "status =") {
		t.Error("Missing status condition")
	}
	if !strings.Contains(query.SQL, "actor =") {
		t.Error("Missing actor condition")
	}
}

func TestNestedFilterComposition(t *testing.T) {
	// Test complex nested filter composition
	outer := NewFilter().WithOr()
	inner1 := StatusEq("open")
	inner2 := ActorEq("user1")
	outer.WithSubFilter(inner1)
	outer.WithSubFilter(inner2)

	qb := NewQueryBuilder().Where(outer)
	query := qb.Build()

	// Should have OR between sub-filters
	if !strings.Contains(query.SQL, " OR ") {
		t.Errorf("Expected OR in nested filter SQL, got: %s", query.SQL)
	}

	// Test with AND as outer
	outer = NewFilter().WithAnd()
	outer.WithSubFilter(StatusEq("open"))
	outer.WithSubFilter(ActorEq("user1"))

	qb = NewQueryBuilder().Where(outer)
	query = qb.Build()

	// Should have AND between sub-filters
	if !strings.Contains(query.SQL, " AND ") {
		t.Errorf("Expected AND in nested filter SQL, got: %s", query.SQL)
	}
}
