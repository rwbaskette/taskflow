package db

import (
	"fmt"
	"strings"
)

// Query represents a built SQL query with parameterized arguments
type Query struct {
	SQL    string
	Params []interface{}
}

// QueryBuilder builds SQL queries with filter support
type QueryBuilder struct {
	tableName  string
	selectCols []string
	filters    []*Filter
	orderBy    string
	orderDesc  bool
	limit      int
	offset     int
	groupBy    []string
	joins      []string
}

// NewQueryBuilder creates a new query builder for the tasks table
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		tableName:  "tasks",
		selectCols: []string{"id", "milestone", "title", "description", "status", "actor", "last_updated"},
		filters:    []*Filter{},
		orderBy:    "last_updated",
		orderDesc:  true,
		limit:      0,
		offset:     0,
		groupBy:    []string{},
		joins:      []string{},
	}
}

// ForTable sets the table name with validation
func (qb *QueryBuilder) ForTable(table string) *QueryBuilder {
	// Validate table name
	if err := validateTableName(table); err != nil {
		// Return with invalid table - validation will fail on Build
		qb.tableName = table
		return qb
	}
	qb.tableName = table
	return qb
}

// Select sets the columns to select with validation
func (qb *QueryBuilder) Select(cols ...string) *QueryBuilder {
	if len(cols) > 0 {
		validCols := []string{}
		// Validate all column names
		for _, col := range cols {
			if err := validateColumnName(col); err != nil {
				// Skip invalid columns - validation will fail on Build
				continue
			}
			validCols = append(validCols, col)
		}
		// Only replace columns if all columns were valid, otherwise keep defaults
		if len(validCols) == len(cols) && len(validCols) > 0 {
			qb.selectCols = validCols
		}
	}
	return qb
}

// Where adds a filter to the query
func (qb *QueryBuilder) Where(filter *Filter) *QueryBuilder {
	if filter != nil {
		qb.filters = append(qb.filters, filter)
	}
	return qb
}

// OrderBy sets the ordering column and direction with validation
func (qb *QueryBuilder) OrderBy(col string, desc bool) *QueryBuilder {
	// Validate column name
	if err := validateColumnName(col); err != nil {
		// Use default column on validation failure
		qb.orderBy = "last_updated"
		qb.orderDesc = desc
		return qb
	}
	qb.orderBy = col
	qb.orderDesc = desc
	return qb
}

// Limit sets the result limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the result offset for pagination
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// GroupBy adds a GROUP BY clause with validation
func (qb *QueryBuilder) GroupBy(cols ...string) *QueryBuilder {
	validCols := []string{}
	for _, col := range cols {
		if err := validateColumnName(col); err == nil {
			validCols = append(validCols, col)
		}
	}
	qb.groupBy = validCols
	return qb
}

// Join adds a JOIN clause with validation
func (qb *QueryBuilder) Join(join string) *QueryBuilder {
	// Validate JOIN clause
	if err := validateJoinClause(join); err != nil {
		// Silently skip invalid JOIN clauses to maintain backward compatibility
		return qb
	}
	qb.joins = append(qb.joins, join)
	return qb
}

// validateQueryBuilder validates the QueryBuilder state before building
func (qb *QueryBuilder) validateQueryBuilder() error {
	// Validate table name
	if err := validateTableName(qb.tableName); err != nil {
		return err
	}

	// Validate select columns
	for _, col := range qb.selectCols {
		if err := validateColumnName(col); err != nil {
			return err
		}
	}

	// Validate order by column
	if qb.orderBy != "" && qb.orderBy != "last_updated" {
		if err := validateColumnName(qb.orderBy); err != nil {
			return err
		}
	}

	// Validate group by columns
	for _, col := range qb.groupBy {
		if err := validateColumnName(col); err != nil {
			return err
		}
	}

	// Validate JOIN clauses
	for _, join := range qb.joins {
		if err := validateJoinClause(join); err != nil {
			return err
		}
	}

	return nil
}

// Build constructs the final Query with parameterized SQL
func (qb *QueryBuilder) Build() *Query {
	params := []interface{}{}
	conditions := []string{}

	// Validate query builder state
	if err := qb.validateQueryBuilder(); err != nil {
		// Return an error query with empty SQL
		return &Query{
			SQL:    "",
			Params: params,
		}
	}

	// Process all filters
	for _, f := range qb.filters {
		if f == nil {
			continue
		}
		condSQL, ok := buildFilterSQL(f, &params)
		if ok && condSQL != "" {
			conditions = append(conditions, condSQL)
		}
	}

	// Build SELECT clause
	selectClause := "*"
	if len(qb.selectCols) > 0 {
		selectClause = strings.Join(qb.selectCols, ", ")
	}

	// Build FROM clause
	fromClause := qb.tableName

	// Build JOINs clause
	joinClause := ""
	if len(qb.joins) > 0 {
		joinClause = " " + strings.Join(qb.joins, " ")
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " ")
	} else {
		// Always include WHERE with 1=1 for consistent SQL generation
		whereClause = " WHERE 1=1"
	}

	// Build GROUP BY clause
	groupByClause := ""
	if len(qb.groupBy) > 0 {
		groupByClause = " GROUP BY " + strings.Join(qb.groupBy, ", ")
	}

	// Build ORDER BY clause
	orderByClause := ""
	if qb.orderBy != "" {
		orderDir := "ASC"
		if qb.orderDesc {
			orderDir = "DESC"
		}
		orderByClause = " ORDER BY " + qb.orderBy + " " + orderDir
	}

	// Build LIMIT clause
	limitClause := ""
	if qb.limit > 0 {
		limitClause = " LIMIT ?"
		params = append(params, qb.limit)
	}

	// Build OFFSET clause (requires LIMIT in SQLite)
	offsetClause := ""
	if qb.offset > 0 {
		// SQLite requires LIMIT when using OFFSET
		// If no limit is set, use a large default
		if qb.limit == 0 {
			limitClause = " LIMIT ?"
			params = append(params, 10000) // Large default for OFFSET-only queries
		}
		offsetClause = " OFFSET ?"
		params = append(params, qb.offset)
	}

	// Construct final SQL
	sql := fmt.Sprintf("SELECT %s FROM %s%s%s%s%s%s%s",
		selectClause,
		fromClause,
		joinClause,
		whereClause,
		groupByClause,
		orderByClause,
		limitClause,
		offsetClause,
	)

	return &Query{
		SQL:    sql,
		Params: params,
	}
}

// buildFilterSQL builds the SQL for a Filter with all its conditions and sub-filters
func buildFilterSQL(f *Filter, params *[]interface{}) (string, bool) {
	if f == nil {
		return "", false
	}

	// Handle case with no conditions and no sub-filters - return empty
	if len(f.Conditions) == 0 && len(f.SubFilters) == 0 {
		return "", false
	}

	logicStr := " " + string(f.Logic) + " "

	// Build conditions
	condParts := []string{}
	for _, cond := range f.Conditions {
		condSQL, ok := buildConditionSQL(cond, params)
		if ok {
			condParts = append(condParts, condSQL)
		}
	}

	// Process sub-filters
	for _, sub := range f.SubFilters {
		if sub == nil {
			continue
		}
		subSQL, ok := buildFilterSQL(sub, params)
		if ok && subSQL != "" {
			condParts = append(condParts, "("+subSQL+")")
		}
	}

	if len(condParts) == 0 {
		return "", false
	}

	// Join all conditions with the logical operator
	return strings.Join(condParts, logicStr), true
}

// BuildTaskQuery builds a query for tasks table using optional filters
func BuildTaskQuery(tf TaskFilter) *Query {
	qb := NewQueryBuilder()

	// Convert TaskFilter to Filter if any filter fields are set
	filter := FromTaskFilter(tf)
	if filter != nil {
		qb.Where(filter)
	}

	// Apply pagination from TaskFilter
	if tf.Limit > 0 {
		qb.Limit(tf.Limit)
	}
	if tf.Offset > 0 {
		qb.Offset(tf.Offset)
	}

	return qb.Build()
}

// BuildCountQuery builds a count query with filters
func (qb *QueryBuilder) BuildCountQuery() *Query {
	params := []interface{}{}
	conditions := []string{}

	// Validate query builder state
	if err := qb.validateQueryBuilder(); err != nil {
		return &Query{
			SQL:    "",
			Params: params,
		}
	}

	// Process all filters respecting their Logic field
	for _, f := range qb.filters {
		if f == nil {
			continue
		}
		condSQL, ok := buildFilterSQL(f, &params)
		if ok && condSQL != "" {
			conditions = append(conditions, condSQL)
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		// Use OR as default for count queries to maximize results
		whereClause = " WHERE " + strings.Join(conditions, " OR ")
	} else {
		whereClause = " WHERE 1=1"
	}

	// Build count SQL
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s%s",
		qb.tableName,
		whereClause,
	)

	return &Query{
		SQL:    sql,
		Params: params,
	}
}

// BuildExistsQuery builds an EXISTS query with filters
func (qb *QueryBuilder) BuildExistsQuery() *Query {
	params := []interface{}{}
	conditions := []string{}

	// Validate query builder state
	if err := qb.validateQueryBuilder(); err != nil {
		return &Query{
			SQL:    "",
			Params: params,
		}
	}

	// Process all filters respecting their Logic field
	for _, f := range qb.filters {
		if f == nil {
			continue
		}
		condSQL, ok := buildFilterSQL(f, &params)
		if ok && condSQL != "" {
			conditions = append(conditions, condSQL)
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		// Use OR as default for EXISTS queries to maximize results
		whereClause = " WHERE " + strings.Join(conditions, " OR ")
	} else {
		whereClause = " WHERE 1=1"
	}

	// Build EXISTS SQL
	sql := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s%s LIMIT 1)",
		qb.tableName,
		whereClause,
	)

	return &Query{
		SQL:    sql,
		Params: params,
	}
}

// ToSQL returns the SQL string with parameters interpolated for debugging
func (q *Query) ToSQL() string {
	if q == nil {
		return ""
	}

	result := q.SQL
	for _, p := range q.Params {
		switch v := p.(type) {
		case string:
			result = strings.Replace(result, "?", fmt.Sprintf("'%s'", v), 1)
		case int, int64:
			result = strings.Replace(result, "?", fmt.Sprintf("%d", v), 1)
		default:
			result = strings.Replace(result, "?", "?", 1)
		}
	}
	return result
}
