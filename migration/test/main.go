package main

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-universal/console"
	"github.com/go-universal/fs"
	"github.com/go-universal/sql/migration"
	"github.com/go-universal/sql/mysql"
)

func main() {
	fs := CreateFS()
	conn := CreateConnection()
	mig := CreateMigration(fs, conn)
	cmd := migration.NewMigrationCLI(
		mig,
		migration.WithDefaultStages("table", "index"),
		migration.WithOutputPath("database/migrations"),
		migration.WithNewCMD(true),
	)

	cmd.ExecuteContext(context.Background())
}

func main2() {
	fs := CreateFS()
	conn := CreateConnection()
	mig := CreateMigration(fs, conn)

	fmt.Println(path.Dir("some/create users table"))

	// results, err := mig.Refresh(
	// 	[]string{"table", "index"},
	// 	migration.OnlyFiles("create users table"),
	// )
	// if err != nil {
	// 	console.Message().Indent().Red("FAIL").Print(err.Error())
	// } else {
	// 	for _, result := range results {
	// 		console.Message().Green("Done").
	// 			Tags("REFRESH", result.Stage).
	// 			Underline().
	// 			Print(result.Name)
	// 	}
	// }

	summary, err := mig.Summary()
	if err != nil {
		log.Fatal(err)
	} else {
		console.PrintF("@Bwb{ Migration Summery: }\n")
		for stage, files := range summary.GroupByStage() {
			console.PrintF("@BUb{%s} @b{Stage} @Ib{(%d Files)}:\n", strings.ToTitle(stage), len(files))
			for _, file := range files {
				console.PrintF("    @U{%s}: @I{%s}\n", file.Name, humanize.Time(file.CreatedAt))
			}

			fmt.Println()
		}
	}
}

func CreateConnection() mysql.Connection {
	conn, err := mysql.New(
		context.Background(),
		mysql.NewConfig().Database("test").Password("root").Build(),
	)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func CreateFS() fs.FlexibleFS {
	return fs.NewDir("database")
}

func CreateMigration(fs fs.FlexibleFS, conn mysql.Connection) migration.Migration {
	mig, err := migration.NewMigration(
		migration.NewMySQLSource(conn),
		fs,
		migration.WithRoot("migrations"),
	)

	if err != nil {
		log.Fatal(err)
	}

	return mig
}

func CreateMigrations() {
	migration.CreateMigrationFile("database/migrations", "create users table", "sql", "table", "index")
	migration.CreateMigrationFile("database/migrations", "create customer table", "sql", "table", "index")
}
