package query

import "strings"

// QueryBuilder builds SQL queries with conditional logic and replacements.
// Use '@in' to generate an IN(args1, args2, ...) SQL clause.
type QueryBuilder interface {
	// And appends a condition using AND.
	And(query string, args ...any) QueryBuilder

	// AndIf appends a condition using AND if 'cond' is true.
	AndIf(cond bool, query string, args ...any) QueryBuilder

	// AndClosure appends a nested condition using AND.
	AndClosure(query string, args ...any) QueryBuilder

	// AndClosureIf appends a nested condition using AND if 'cond' is true.
	AndClosureIf(cond bool, query string, args ...any) QueryBuilder

	// Or appends a condition using OR.
	Or(query string, args ...any) QueryBuilder

	// OrIf appends a condition using OR if 'cond' is true.
	OrIf(cond bool, query string, args ...any) QueryBuilder

	// OrClosure appends a nested condition using OR.
	OrClosure(query string, args ...any) QueryBuilder

	// OrClosureIf appends a nested condition using OR if 'cond' is true.
	OrClosureIf(cond bool, query string, args ...any) QueryBuilder

	// AndNested appends a nested group of conditions using AND.
	AndNested(cb func(qb QueryBuilder)) QueryBuilder

	// OrNested appends a nested group of conditions using OR.
	OrNested(cb func(qb QueryBuilder)) QueryBuilder

	// Replace swaps occurrences of 'old' with 'new' in the final SQL query.
	// Common placeholders include '@sort' and '@order'.
	Replace(old, new string) QueryBuilder

	// Build constructs the final SQL query string.
	// Replaces '@conditions' with SQL conditions and '@where' with WHERE conditions if applicable.
	Build() string

	// Arguments returns the list of query arguments.
	Arguments() []any
}

type queryItem struct {
	joiner    string
	query     string
	closure   bool
	arguments []any
}

type queryBuilder struct {
	sql          string
	resolver     PlaceholderResolver
	conditions   []queryItem
	replacements []string
}

func (b *queryBuilder) And(q string, args ...any) QueryBuilder {
	b.addItem(q, "AND", false, args...)
	return b
}

func (b *queryBuilder) AndIf(c bool, q string, args ...any) QueryBuilder {
	if c {
		b.addItem(q, "AND", false, args...)
	}
	return b
}

func (b *queryBuilder) AndClosure(q string, args ...any) QueryBuilder {
	b.addItem(q, "AND", true, args...)
	return b
}

func (b *queryBuilder) AndClosureIf(c bool, q string, args ...any) QueryBuilder {
	if c {
		b.addItem(q, "AND", true, args...)
	}
	return b
}

func (b *queryBuilder) Or(q string, args ...any) QueryBuilder {
	b.addItem(q, "OR", false, args...)
	return b
}

func (b *queryBuilder) OrIf(c bool, q string, args ...any) QueryBuilder {
	if c {
		b.addItem(q, "OR", false, args...)
	}
	return b
}

func (b *queryBuilder) OrClosure(q string, args ...any) QueryBuilder {
	b.addItem(q, "OR", true, args...)
	return b
}

func (b *queryBuilder) OrClosureIf(c bool, q string, args ...any) QueryBuilder {
	if c {
		b.addItem(q, "OR", true, args...)
	}
	return b
}

func (b *queryBuilder) AndNested(cb func(QueryBuilder)) QueryBuilder {
	nested := &queryBuilder{
		sql:        "",
		resolver:   b.resolver,
		conditions: []queryItem{},
	}

	cb(nested)

	if len(nested.conditions) == 0 {
		return b
	}

	nestedSQL := nested.sqlConditions()
	b.addItem(nestedSQL, "AND", true, nested.Arguments()...)

	return b
}

func (b *queryBuilder) OrNested(cb func(QueryBuilder)) QueryBuilder {
	nested := &queryBuilder{
		sql:        "",
		resolver:   b.resolver,
		conditions: []queryItem{},
	}

	cb(nested)

	if len(nested.conditions) == 0 {
		return b
	}

	nestedSQL := nested.sqlConditions()
	b.addItem(nestedSQL, "OR", true, nested.Arguments()...)

	return b
}

func (b *queryBuilder) Replace(o, n string) QueryBuilder {
	b.replacements = append(b.replacements, o, n)
	return b
}

func (b *queryBuilder) Build() string {
	conditions := b.sqlConditions()
	where := ""
	if conditions != "" {
		where = "WHERE " + conditions
	}

	return strings.NewReplacer(
		append(
			b.replacements,
			"@conditions", conditions,
			"@where", where,
		)...,
	).Replace(b.sql)
}

func (b *queryBuilder) Arguments() []any {
	args := make([]any, 0)
	for _, q := range b.conditions {
		args = append(args, q.arguments...)
	}
	return args
}

func (b *queryBuilder) addItem(query, joiner string, closure bool, args ...any) {
	if strings.TrimSpace(query) == "" {
		return
	}

	b.conditions = append(b.conditions, queryItem{
		joiner:    joiner,
		query:     query,
		closure:   closure,
		arguments: args,
	})
}

func (b *queryBuilder) sqlConditions() string {
	conditions := ""

	// Generate conditions
	for _, cond := range b.conditions {
		query := cond.query

		// Generate @in placeholder
		if strings.Contains(query, "@in") {
			placeholders := strings.TrimLeft(strings.Repeat(", ?", len(cond.arguments)), ", ")
			query = strings.Replace(query, "@in", "IN ("+placeholders+")", 1)
		}

		// Wrap subquery conditions in parentheses
		if cond.closure {
			query = "(" + query + ")"
		}

		if conditions == "" {
			conditions = query
		} else {
			conditions = conditions + " " + cond.joiner + " " + query
		}
	}

	if b.resolver == nil {
		return conditions
	}

	// Replace '?' placeholders with custom placeholders.
	counter := 0
	var builder strings.Builder
	builder.Grow(len(conditions) + 10)
	for i := range conditions {
		if conditions[i] == '?' {
			counter++
			builder.WriteString(b.resolver(counter))
		} else {
			builder.WriteByte(conditions[i])
		}
	}
	return builder.String()
}
