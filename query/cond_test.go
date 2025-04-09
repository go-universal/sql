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
	cond := query.NewCondition()
	cond.And("name = ?", "John").
		AndClosure("age > ? AND age < ?", 9, 31).
		OrIf(false, "age IS NULL").
		OrClosureIf(true, "membership @in", "admin", "manager", "accountant")

	expected := "SELECT COUNT(*) FROM `users` WHERE name = ? AND (age > ? AND age < ?) OR (membership IN (?, ?, ?));"
	result := cond.Build("SELECT COUNT(*) FROM `users` @where;")

	assert.Equal(t, expected, result, "Built SQL mismatch")
}
