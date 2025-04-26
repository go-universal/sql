package query

import "strings"

// ConditionBuilder defines an interface for dynamically constructing SQL conditions.
// Use '@in' as a placeholder to generate an IN(args1, args2, ...) SQL clause.
type ConditionBuilder interface {
	// SetResolver assigns a custom resolver for handling placeholders in SQL queries.
	SetResolver(resolver PlaceholderResolver) ConditionBuilder

	// And appends a condition using AND.
	And(query string, args ...any) ConditionBuilder

	// AndIf appends a condition using AND if 'cond' is true.
	AndIf(cond bool, query string, args ...any) ConditionBuilder

	// AndClosure appends a nested condition using AND.
	AndClosure(query string, args ...any) ConditionBuilder

	// AndClosureIf appends a nested condition using AND if 'cond' is true.
	AndClosureIf(cond bool, query string, args ...any) ConditionBuilder

	// Or appends a condition using OR.
	Or(query string, args ...any) ConditionBuilder

	// OrIf appends a condition using OR if 'cond' is true.
	OrIf(cond bool, query string, args ...any) ConditionBuilder

	// OrClosure appends a nested condition using OR.
	OrClosure(query string, args ...any) ConditionBuilder

	// OrClosureIf appends a nested condition using OR if 'cond' is true.
	OrClosureIf(cond bool, query string, args ...any) ConditionBuilder

	// Replace substitutes occurrences of the specified old phrase with the new phrase
	// in the final SQL query (e.g., "@sort", "@order").
	Replace(old, new string) ConditionBuilder

	// SQL returns the constructed conditions as a raw SQL string.
	SQL() string

	// Build inserts the generated SQL conditions into the provided query string.
	// Replaces "@conditions" with the SQL conditions and "@where" with `WHERE` followed
	// by the conditions (if applicable).
	Build(query string) string

	// Arguments returns the list of arguments associated with the conditions.
	Arguments() []any
}

type conditionItem struct {
	joiner    string
	query     string
	closure   bool
	arguments []any
}

type conditionBuilder struct {
	resolver     PlaceholderResolver
	conditions   []conditionItem
	replacements []string
}

// NewCondition creates and returns a new ConditionBuilder instance.
// Accepts optional PlaceholderResolver for handling placeholders in SQL queries.
func NewCondition(resolver ...PlaceholderResolver) ConditionBuilder {
	return &conditionBuilder{
		resolver:     parseVariadic(nil, resolver...),
		conditions:   make([]conditionItem, 0),
		replacements: make([]string, 0),
	}
}

func (b *conditionBuilder) SetResolver(r PlaceholderResolver) ConditionBuilder {
	b.resolver = r
	return b
}

func (b *conditionBuilder) And(q string, args ...any) ConditionBuilder {
	b.addItem(q, "AND", false, args...)
	return b
}

func (b *conditionBuilder) AndIf(c bool, q string, args ...any) ConditionBuilder {
	if c {
		b.addItem(q, "AND", false, args...)
	}
	return b
}

func (b *conditionBuilder) AndClosure(q string, args ...any) ConditionBuilder {
	b.addItem(q, "AND", true, args...)
	return b
}

func (b *conditionBuilder) AndClosureIf(c bool, q string, args ...any) ConditionBuilder {
	if c {
		b.addItem(q, "AND", true, args...)
	}
	return b
}

func (b *conditionBuilder) Or(q string, args ...any) ConditionBuilder {
	b.addItem(q, "OR", false, args...)
	return b
}

func (b *conditionBuilder) OrIf(c bool, q string, args ...any) ConditionBuilder {
	if c {
		b.addItem(q, "OR", false, args...)
	}
	return b
}

func (b *conditionBuilder) OrClosure(q string, args ...any) ConditionBuilder {
	b.addItem(q, "OR", true, args...)
	return b
}

func (b *conditionBuilder) OrClosureIf(c bool, q string, args ...any) ConditionBuilder {
	if c {
		b.addItem(q, "OR", true, args...)
	}
	return b
}

func (b *conditionBuilder) Replace(o, n string) ConditionBuilder {
	b.replacements = append(b.replacements, o, n)
	return b
}

func (b *conditionBuilder) SQL() string {
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

func (b *conditionBuilder) Build(q string) string {
	conditions := b.SQL()
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
	).Replace(q)
}

func (b *conditionBuilder) Arguments() []any {
	args := make([]any, 0)
	for _, q := range b.conditions {
		args = append(args, q.arguments...)
	}
	return args
}

func (b *conditionBuilder) addItem(query, joiner string, closure bool, args ...any) {
	if strings.TrimSpace(query) == "" {
		return
	}

	b.conditions = append(b.conditions, conditionItem{
		joiner:    joiner,
		query:     query,
		closure:   closure,
		arguments: args,
	})
}
