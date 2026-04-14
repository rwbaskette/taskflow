package db

import (
	"testing"
)

func TestFilterFieldConstants(t *testing.T) {
	// Verify all filter field constants are correctly defined
	fields := []FilterField{
		FilterFieldMilestone,
		FilterFieldStatus,
		FilterFieldActor,
		FilterFieldTitle,
		FilterFieldID,
		FilterFieldDescription,
	}

	expectedFields := []FilterField{
		"milestone",
		"status",
		"actor",
		"title",
		"id",
		"description",
	}

	if len(fields) != len(expectedFields) {
		t.Fatalf("expected %d fields, got %d", len(expectedFields), len(fields))
	}

	for i, f := range fields {
		if f != expectedFields[i] {
			t.Errorf("expected field %s, got %s", expectedFields[i], f)
		}
	}
}

func TestFilterOpConstants(t *testing.T) {
	// Verify all filter operator constants are correctly defined
	ops := []FilterOp{
		FilterOpEq,
		FilterOpNe,
		FilterOpLike,
		FilterOpIn,
		FilterOpNotIn,
	}

	expectedOps := []FilterOp{
		"=",
		"!=",
		"LIKE",
		"IN",
		"NOT IN",
	}

	if len(ops) != len(expectedOps) {
		t.Fatalf("expected %d operators, got %d", len(expectedOps), len(ops))
	}

	for i, op := range ops {
		if op != expectedOps[i] {
			t.Errorf("expected operator %s, got %s", expectedOps[i], op)
		}
	}
}

func TestFilterLogicConstants(t *testing.T) {
	// Verify filter logic constants
	if FilterLogicAnd != "AND" {
		t.Errorf("expected AND, got %s", FilterLogicAnd)
	}
	if FilterLogicOr != "OR" {
		t.Errorf("expected OR, got %s", FilterLogicOr)
	}
}

func TestNewFilter(t *testing.T) {
	f := NewFilter()

	if f == nil {
		t.Fatal("NewFilter returned nil")
	}

	if len(f.Conditions) != 0 {
		t.Errorf("expected empty conditions, got %d", len(f.Conditions))
	}

	if f.Logic != FilterLogicAnd {
		t.Errorf("expected default AND logic, got %s", f.Logic)
	}

	if len(f.SubFilters) != 0 {
		t.Errorf("expected empty sub-filters, got %d", len(f.SubFilters))
	}
}

func TestFilterWithCondition(t *testing.T) {
	f := NewFilter()
	f.WithCondition(FilterFieldMilestone, FilterOpEq, "v1.0")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Field != FilterFieldMilestone {
		t.Errorf("expected field milestone, got %s", cond.Field)
	}
	if cond.Operator != FilterOpEq {
		t.Errorf("expected operator =, got %s", cond.Operator)
	}
	if cond.Value != "v1.0" {
		t.Errorf("expected value v1.0, got %v", cond.Value)
	}
}

func TestFilterWithAnd(t *testing.T) {
	f := NewFilter().WithOr().WithAnd()

	if f.Logic != FilterLogicAnd {
		t.Errorf("expected AND logic, got %s", f.Logic)
	}
}

func TestFilterWithOr(t *testing.T) {
	f := NewFilter().WithOr()

	if f.Logic != FilterLogicOr {
		t.Errorf("expected OR logic, got %s", f.Logic)
	}
}

func TestFilterWithSubFilter(t *testing.T) {
	f := NewFilter()
	sub := NewFilter().WithCondition(FilterFieldStatus, FilterOpEq, "open")
	f.WithSubFilter(sub)

	if len(f.SubFilters) != 1 {
		t.Fatalf("expected 1 sub-filter, got %d", len(f.SubFilters))
	}

	if f.SubFilters[0].Conditions[0].Field != FilterFieldStatus {
		t.Errorf("expected sub-filter to have status field")
	}
}

func TestMilestoneEq(t *testing.T) {
	f := MilestoneEq("v1.0")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Field != FilterFieldMilestone {
		t.Errorf("expected field milestone, got %s", cond.Field)
	}
	if cond.Operator != FilterOpEq {
		t.Errorf("expected operator =, got %s", cond.Operator)
	}
	if cond.Value != "v1.0" {
		t.Errorf("expected value v1.0, got %v", cond.Value)
	}
}

func TestMilestoneNe(t *testing.T) {
	f := MilestoneNe("v1.0")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Operator != FilterOpNe {
		t.Errorf("expected operator !=, got %s", cond.Operator)
	}
}

func TestMilestoneIn(t *testing.T) {
	// Test single value - should use equality
	f := MilestoneIn("v1.0")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpEq {
		t.Errorf("expected equality for single value, got %s", f.Conditions[0].Operator)
	}

	// Test multiple values - should use IN
	f = MilestoneIn("v1.0", "v2.0")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpIn {
		t.Errorf("expected IN for multiple values, got %s", f.Conditions[0].Operator)
	}
}

func TestStatusEq(t *testing.T) {
	f := StatusEq("open")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Field != FilterFieldStatus {
		t.Errorf("expected field status, got %s", cond.Field)
	}
	if cond.Operator != FilterOpEq {
		t.Errorf("expected operator =, got %s", cond.Operator)
	}
	if cond.Value != "open" {
		t.Errorf("expected value open, got %v", cond.Value)
	}
}

func TestStatusNe(t *testing.T) {
	f := StatusNe("closed")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Operator != FilterOpNe {
		t.Errorf("expected operator !=, got %s", cond.Operator)
	}
}

func TestStatusIn(t *testing.T) {
	// Test single value
	f := StatusIn("open")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpEq {
		t.Errorf("expected equality for single value, got %s", f.Conditions[0].Operator)
	}

	// Test multiple values
	f = StatusIn("open", "in_progress")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpIn {
		t.Errorf("expected IN for multiple values, got %s", f.Conditions[0].Operator)
	}
}

func TestActorEq(t *testing.T) {
	f := ActorEq("user1")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Field != FilterFieldActor {
		t.Errorf("expected field actor, got %s", cond.Field)
	}
	if cond.Operator != FilterOpEq {
		t.Errorf("expected operator =, got %s", cond.Operator)
	}
	if cond.Value != "user1" {
		t.Errorf("expected value user1, got %v", cond.Value)
	}
}

func TestActorNe(t *testing.T) {
	f := ActorNe("user1")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Operator != FilterOpNe {
		t.Errorf("expected operator !=, got %s", cond.Operator)
	}
}

func TestActorIn(t *testing.T) {
	// Test single value
	f := ActorIn("user1")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpEq {
		t.Errorf("expected equality for single value, got %s", f.Conditions[0].Operator)
	}

	// Test multiple values
	f = ActorIn("user1", "user2")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpIn {
		t.Errorf("expected IN for multiple values, got %s", f.Conditions[0].Operator)
	}
}

func TestTitleLike(t *testing.T) {
	f := TitleLike("bug")

	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}

	cond := f.Conditions[0]
	if cond.Field != FilterFieldTitle {
		t.Errorf("expected field title, got %s", cond.Field)
	}
	if cond.Operator != FilterOpLike {
		t.Errorf("expected operator LIKE, got %s", cond.Operator)
	}
	// Value should be wrapped with %
	if cond.Value != "%bug%" {
		t.Errorf("expected value %%bug%%, got %v", cond.Value)
	}
}

func TestIDIn(t *testing.T) {
	// Test single value
	f := IDIn("1")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpEq {
		t.Errorf("expected equality for single value, got %s", f.Conditions[0].Operator)
	}

	// Test multiple values
	f = IDIn("1", "2", "3")
	if len(f.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(f.Conditions))
	}
	if f.Conditions[0].Operator != FilterOpIn {
		t.Errorf("expected IN for multiple values, got %s", f.Conditions[0].Operator)
	}
}

func TestCombineFilters(t *testing.T) {
	f1 := StatusEq("open")
	f2 := ActorEq("user1")

	// Test AND logic
	combined := CombineFilters(FilterLogicAnd, f1, f2)
	if combined.Logic != FilterLogicAnd {
		t.Errorf("expected AND logic, got %s", combined.Logic)
	}
	if len(combined.SubFilters) != 2 {
		t.Errorf("expected 2 sub-filters, got %d", len(combined.SubFilters))
	}

	// Test OR logic
	combined = CombineFilters(FilterLogicOr, f1, f2)
	if combined.Logic != FilterLogicOr {
		t.Errorf("expected OR logic, got %s", combined.Logic)
	}
}

func TestAnd(t *testing.T) {
	f1 := StatusEq("open")
	f2 := ActorEq("user1")
	f := And(f1, f2)

	if f.Logic != FilterLogicAnd {
		t.Errorf("expected AND logic, got %s", f.Logic)
	}
	if len(f.SubFilters) != 2 {
		t.Errorf("expected 2 sub-filters, got %d", len(f.SubFilters))
	}
}

func TestOr(t *testing.T) {
	f1 := StatusEq("open")
	f2 := ActorEq("user1")
	f := Or(f1, f2)

	if f.Logic != FilterLogicOr {
		t.Errorf("expected OR logic, got %s", f.Logic)
	}
	if len(f.SubFilters) != 2 {
		t.Errorf("expected 2 sub-filters, got %d", len(f.SubFilters))
	}
}

func TestFromTaskFilter(t *testing.T) {
	// Test with all fields set
	tf := TaskFilter{
		Milestone: "v1.0",
		Status:    "open",
		Actor:     "user1",
	}

	f := FromTaskFilter(tf)

	if f == nil {
		t.Fatal("FromTaskFilter returned nil")
	}

	if len(f.Conditions) != 3 {
		t.Errorf("expected 3 conditions, got %d", len(f.Conditions))
	}

	// Test with empty filter
	tf = TaskFilter{}
	f = FromTaskFilter(tf)

	if len(f.Conditions) != 0 {
		t.Errorf("expected 0 conditions for empty filter, got %d", len(f.Conditions))
	}
}

func TestBuildConditionSQL(t *testing.T) {
	params := []interface{}{}

	// Test equality condition
	cond := Condition{
		Field:    FilterFieldMilestone,
		Operator: FilterOpEq,
		Value:    "v1.0",
	}

	sql, ok := buildConditionSQL(cond, &params)
	if !ok {
		t.Error("buildConditionSQL returned false for valid condition")
	}
	if sql != "milestone = ?" {
		t.Errorf("expected 'milestone = ?', got '%s'", sql)
	}
	if len(params) != 1 || params[0] != "v1.0" {
		t.Errorf("expected params [v1.0], got %v", params)
	}

	// Test NOT IN condition
	params = []interface{}{}
	cond = Condition{
		Field:    FilterFieldStatus,
		Operator: FilterOpNotIn,
		Value:    []string{"open", "closed"},
	}

	sql, ok = buildConditionSQL(cond, &params)
	if !ok {
		t.Error("buildConditionSQL returned false for IN condition")
	}
	if sql != "status NOT IN (?, ?)" {
		t.Errorf("expected 'status NOT IN (?, ?)', got '%s'", sql)
	}
	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}

	// Test LIKE condition
	params = []interface{}{}
	cond = Condition{
		Field:    FilterFieldTitle,
		Operator: FilterOpLike,
		Value:    "bug",
	}

	sql, ok = buildConditionSQL(cond, &params)
	if !ok {
		t.Error("buildConditionSQL returned false for LIKE condition")
	}
	if sql != "title LIKE ?" {
		t.Errorf("expected 'title LIKE ?', got '%s'", sql)
	}

	// Test IN with interface{} slice
	params = []interface{}{}
	cond = Condition{
		Field:    FilterFieldID,
		Operator: FilterOpIn,
		Value:    []interface{}{"1", "2", "3"},
	}

	sql, ok = buildConditionSQL(cond, &params)
	if !ok {
		t.Error("buildConditionSQL returned false for IN condition with []interface{}")
	}
	if sql != "id IN (?, ?, ?)" {
		t.Errorf("expected 'id IN (?, ?, ?)', got '%s'", sql)
	}
	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}

	// Test invalid IN condition (empty slice)
	params = []interface{}{}
	cond = Condition{
		Field:    FilterFieldStatus,
		Operator: FilterOpIn,
		Value:    []string{},
	}

	sql, ok = buildConditionSQL(cond, &params)
	if ok {
		t.Error("buildConditionSQL should return false for empty IN values")
	}
}

func TestBuildConditionSQL_EmptyValues(t *testing.T) {
	params := []interface{}{}

	// Test IN with empty slice
	cond := Condition{
		Field:    FilterFieldStatus,
		Operator: FilterOpIn,
		Value:    []string{},
	}

	sql, ok := buildConditionSQL(cond, &params)
	if ok {
		t.Error("Expected false for empty IN values")
	}
	if sql != "" {
		t.Errorf("Expected empty SQL for empty IN values, got '%s'", sql)
	}

	// Test NOT IN with empty slice
	cond = Condition{
		Field:    FilterFieldStatus,
		Operator: FilterOpNotIn,
		Value:    []string{},
	}

	sql, ok = buildConditionSQL(cond, &params)
	if ok {
		t.Error("Expected false for empty NOT IN values")
	}
	if sql != "" {
		t.Errorf("Expected empty SQL for empty NOT IN values, got '%s'", sql)
	}
}

func TestValidateColumnName(t *testing.T) {
	tests := []struct {
		name    string
		col     string
		wantErr bool
	}{
		{"valid column", "id", false},
		{"valid column", "milestone", false},
		{"valid column", "title", false},
		{"valid column", "status", false},
		{"valid column", "actor", false},
		{"valid column", "last_updated", false},
		{"invalid column", "invalid_col", true},
		{"invalid column", "name; DROP TABLE tasks;--", true},
		{"invalid column", "name--", true},
		{"invalid column", "name/*comment*/", true},
		{"empty column", "", true},
		{"quote injection", "id'; DELETE FROM tasks--", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateColumnName(tt.col)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateColumnName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTableName(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		wantErr bool
	}{
		{"valid table", "tasks", false},
		{"invalid table", "users", true},
		{"invalid table", "tasks; DROP TABLE tasks;--", true},
		{"invalid table", "tasks--", true},
		{"empty table", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTableName(tt.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTableName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJoinClause(t *testing.T) {
	tests := []struct {
		name    string
		join    string
		wantErr bool
	}{
		{"valid join", "JOIN users ON tasks.user_id = users.id", false},
		{"valid join", "LEFT JOIN comments ON tasks.id = comments.task_id", false},
		{"invalid join with semicolon", "JOIN users ON tasks.id = users.id; DROP TABLE tasks--", true},
		{"invalid join with comment marker", "JOIN users ON tasks.id = users.id -- comment", true},
		{"invalid join with block comment", "JOIN users ON tasks.id = users.id /* comment */", true},
		{"empty join", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJoinClause(tt.join)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateJoinClause() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseValue(t *testing.T) {
	// Test LIKE operator
	val, ok := parseValue(FilterOpLike, "test")
	if !ok {
		t.Error("parseValue returned false for LIKE with string")
	}
	if val != "test" {
		t.Errorf("expected 'test', got %v", val)
	}

	// Test IN operator with string slice
	val, ok = parseValue(FilterOpIn, []string{"a", "b"})
	if !ok {
		t.Error("parseValue returned false for IN with []string")
	}
	if val == nil {
		t.Error("parseValue returned nil for IN with []string")
	}

	// Test IN operator with empty slice
	val, ok = parseValue(FilterOpIn, []string{})
	if ok {
		t.Error("parseValue should return false for empty IN values")
	}

	// Test default operator
	val, ok = parseValue(FilterOpEq, 123)
	if !ok {
		t.Error("parseValue returned false for Eq with int")
	}
	if val != 123 {
		t.Errorf("expected 123, got %v", val)
	}
}
