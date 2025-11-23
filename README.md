# Go SQL Utility Documentation

![GitHub Tag](https://img.shields.io/github/v/tag/go-universal/sql?sort=semver&label=version)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-universal/sql.svg)](https://pkg.go.dev/github.com/go-universal/sql)
[![License](https://img.shields.io/badge/license-ISC-blue.svg)](https://github.com/go-universal/sql/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-universal/sql)](https://goreportcard.com/report/github.com/go-universal/sql)
![Contributors](https://img.shields.io/github/contributors/go-universal/sql)
![Issues](https://img.shields.io/github/issues/go-universal/sql)

`sql` is a Go library designed to simplify database interactions, migrations, and query management.

## Packages

### Query Builder

The QueryBuilder provide functions for dynamically constructing SQL conditions.

#### Basic Conditions

```go
import "github.com/go-universal/sql/query"

func main() {
    cond := query.NewCondition(query.NumbericResolver)
    cond.And("name = ?", "John").
        AndClosure("age > ? AND age < ?", 9, 31).
        OrIf(false, "age IS NULL").
        OrClosureIf(true, "membership @in", "admin", "manager", "accountant")

    // Result: "name = $1 AND (age > $2 AND age < $3) OR (membership IN ($4, $5, $6))"
}
```

#### Nested Conditions

The QueryBuilder supports nested conditions using `AndNested` and `OrNested` for complex query logic.

##### AndNested

`AndNested` appends a nested group of conditions joined with AND:

```go
func main() {
    q := query.NewCondition(nil)
    q.And("deleted_at IS NULL").
        AndNested(func(qb query.QueryBuilder) {
            qb.Or("name = ?", "John").
                Or("family = ?", "Doe")
        })

    // Result: "deleted_at IS NULL AND (name = ? OR family = ?)"
}
```

##### OrNested

`OrNested` appends a nested group of conditions joined with OR:

```go
func main() {
    q := query.NewCondition(nil)
    q.And("deleted_at IS NULL").
        OrNested(func(qb query.QueryBuilder) {
            qb.And("age > ?", 18).
                And("status = ?", "active")
        })

    // Result: "deleted_at IS NULL OR (age > ? AND status = ?)"
}
```

##### Deep Nesting

Nested builders can contain other nested builders for complex multi-level conditions:

```go
func main() {
    q := query.NewCondition(nil)
    q.And("deleted_at IS NULL").
        AndNested(func(qb query.QueryBuilder) {
            qb.Or("name = ?", "John").
                OrNested(func(nested query.QueryBuilder) {
                    nested.And("status = ?", "active").
                        And("role = ?", "admin")
                })
        })

    // Result: "deleted_at IS NULL AND (name = ? OR (status = ? AND role = ?))"
}
```

### Query Manager

The `query` package provides tools for managing and generating SQL queries.

#### Example

```go
import (
    "github.com/go-universal/sql/query"
    "github.com/go-universal/fs"
)
func main() {
    queriesFS := fs.NewDir("database/queries")
    queryManager, _err_ := query.NewQueryManager(fs, query.WithRoot("database"))

    usersList :=  queryManager.Get("queries/users/users_list")
    usersTrash :=  queryManager.Get("queries/users/deleted users")
    customers, exists :=  queryManager.Find("queries/customers/list")
    customers, exists :=  queryManager.Find("queries/customers/deleted") // "", false
}
```

Query files style:

```sql
-- users.sql

-- { query: users_list }
SELECT * FROM users WHERE `deleted_at` IS NULL AND `name` LIKE ?;

-- { query: deleted users }
SELECT * FROM users WHERE deleted_at IS NOT NULL;


-- customers.sql
-- { query: list }
SELECT * from customers;
```

### Postgres Package

The `postgres` package provides tools for constructing and executing SQL commands specifically for PostgreSQL databases. Query placeholders must `?`.

```go
package main

import (
    "context"

    "github.com/go-universal/sql/postgres"
    "github.com/jackc/pgx/v5/pgconn"
)

func main() {
    ctx := context.Background()
    config := postgres.NewConfig().
        Host("localhost").
        Port(5432).
        User("postgres").
        Password("password").
        Database("test").
        SSLMode("disable").
        MinConns(2)
    conn, err := postgres.New(
        ctx, config.Build(),
        func(c *pgxpool.Config) { c.MaxConns = 7 },
    )
    defer conn.Close(ctx)

    cmd := postgres.NewCmd(conn)
    result, err := cmd.Command("INSERT INTO users (name) VALUES (?)").Exec(ctx, "John Doe")
}
```

### MySQL Package

The `mysql` package provides tools for constructing and executing SQL commands specifically for MySQL databases.

```go
package main

import (
    "context"

    "github.com/go-universal/sql/mysql"
)

func main() {
    conn, err := mysql.New(
        context.Background(),
        mysql.NewConfig().Database("test").Password("root").Build(),
    )
    if err != nil {
        log.Fatal(err)
    }

    cmd := mysql.NewCmd(conn)
    result, err := cmd.Command("INSERT INTO users (name) VALUES (?)").Exec(context.Background(), "John Doe")
}
```

### Migration Package

The `migration` package provides tools for managing database migrations by stage.

```go
package main

import (
    "log"

    "github.com/go-universal/fs"
    "github.com/go-universal/sql/migration"
    "github.com/go-universal/sql/mysql"
)

func main() {
    conn := CreateConnection()
    fs := CreateFS()
    mig, err := migration.NewMigration(
        migration.NewMySQLSource(conn),
        fs,
        migration.WithRoot("migrations"),
    )

    err := mig.Up([]string{"table", "index", "seed"})
    if err != nil {
        log.Fatal(err)
    }
}
```

Migration files style:

```sql
-- 1741791024-create-users-table.sql
-- { up: table } table is sectin name
CREATE TABLE IF NOT EXISTS ...

-- { down: table }

-- { up: index }
...

-- { down: index }
...

-- { up: seed }
...

-- { down: seed }
...
```

## License

This library is licensed under the ISC License. See the [LICENSE](LICENSE) file for details.
