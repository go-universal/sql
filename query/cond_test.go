package query_test

import (
	"testing"

	"github.com/go-universal/sql/query"
	"github.com/stretchr/testify/assert"
)

func TestConditionBuilder_SQL(t *testing.T) {
	cond := query.NewCondition(query.NumbericResolver)
	cond.And("name = ?", "John").
		AndClosure("age > ? AND age < ?", 9, 31).
		OrIf(false, "age IS NULL").
		OrClosureIf(true, "membership @in", "admin", "manager", "accountant")

	expected := "name = $1 AND (age > $2 AND age < $3) OR (membership IN ($4, $5, $6))"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "SQL output mismatch")
}

func TestConditionBuilder_Build(t *testing.T) {
	cond := query.NewCondition().SetQuote(query.DoubleQuoteResolver)
	cond.And("name = ?", "John").
		AndClosure("age > ? AND age < ?", 9, 31).
		OrIf(false, "age IS NULL").
		OrClosureIf(true, `membership "@in"`, "admin", "manager", "accountant")

	expected := "SELECT COUNT(*) FROM `users` WHERE name = ? AND (age > ? AND age < ?) OR (membership IN (?, ?, ?));"
	result := cond.Build("SELECT COUNT(*) FROM `users` @where;")

	assert.Equal(t, expected, result, "Built SQL mismatch")
}

func TestConditionBuilder_AndNested(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		AndNested(func(qb query.ConditionBuilder) {
			qb.Or("name = ?", "John").
				Or("family = ?", "Doe")
		})

	expected := "deleted_at IS NULL AND (name = ? OR family = ?)"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "AndNested SQL output mismatch")
}

func TestConditionBuilder_OrNested(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		OrNested(func(qb query.ConditionBuilder) {
			qb.And("age > ?", 18).
				And("status = ?", "active")
		})

	expected := "deleted_at IS NULL OR (age > ? AND status = ?)"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "OrNested SQL output mismatch")
}

func TestConditionBuilder_MultipleNested(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		AndNested(func(qb query.ConditionBuilder) {
			qb.Or("name = ?", "John").
				Or("family = ?", "Doe")
		}).
		OrNested(func(qb query.ConditionBuilder) {
			qb.And("status = ?", "active").
				And("role = ?", "admin")
		})

	expected := "deleted_at IS NULL AND (name = ? OR family = ?) OR (status = ? AND role = ?)"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "Multiple nested SQL output mismatch")
}

func TestConditionBuilder_EmptyAndNested(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		AndNested(func(qb query.ConditionBuilder) {
			// Empty nested builder
		})

	expected := "deleted_at IS NULL"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "Empty AndNested should be ignored")
}

func TestConditionBuilder_EmptyOrNested(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		OrNested(func(qb query.ConditionBuilder) {
			// Empty nested builder
		})

	expected := "deleted_at IS NULL"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "Empty OrNested should be ignored")
}

func TestConditionBuilder_DeeplyNested(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		AndNested(func(qb query.ConditionBuilder) {
			qb.Or("name = ?", "John").
				OrNested(func(nested query.ConditionBuilder) {
					nested.And("status = ?", "active").
						And("role = ?", "admin")
				})
		})

	expected := "deleted_at IS NULL AND (name = ? OR (status = ? AND role = ?))"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "Deeply nested SQL output mismatch")
}

func TestConditionBuilder_NestedWithNumericResolver(t *testing.T) {
	cond := query.NewCondition(query.NumbericResolver)
	cond.And("deleted_at IS NULL").
		AndNested(func(qb query.ConditionBuilder) {
			qb.Or("name = ?", "John").
				Or("family = ?", "Doe")
		})

	expected := "deleted_at IS NULL AND (name = $1 OR family = $2)"
	actual := cond.SQL()

	assert.Equal(t, expected, actual, "Nested with numeric resolver SQL output mismatch")
}

func TestConditionBuilder_NestedBuild(t *testing.T) {
	cond := query.NewCondition()
	cond.And("deleted_at IS NULL").
		AndNested(func(qb query.ConditionBuilder) {
			qb.Or("name = ?", "John").
				Or("family = ?", "Doe")
		})

	expected := "SELECT * FROM users WHERE deleted_at IS NULL AND (name = ? OR family = ?);"
	result := cond.Build("SELECT * FROM users @where;")

	assert.Equal(t, expected, result, "Nested Build result mismatch")
}
