# Gorep - Golang repository generator

![example branch parameter](https://github.com/vehsamrak/gorep/actions/workflows/main.yml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/vehsamrak/gorep/branch/main/graph/badge.svg?token=1wSNzO0Ds1)](https://codecov.io/gh/vehsamrak/gorep)

Tired of hand-writing DTO's and SQL for accessing database data?
Gorep aims to automate this tedious and repeatable task for you, using boilerplate code generation.

It can:

* Generate DTO structures from database tables
* [not ready for now] Generate Repository classes with fetching methods containing SQL and DTO-mapping boilerplate.

### What is repository

Repository mediates between the domain and data mapping layers using a collection-like interface 
for accessing domain objects. (c) [Martin Fowler](https://martinfowler.com/eaaCatalog/repository.html)

### Supported databases
Only PostgreSQL is supported for now.
Default schema is "public", and could be changed prefixing table name with schema. For example, to fetch use 
table named "table_name" and schema "schema_name" you should pass "schema_name.table_name" as table name. If no prefix
set to table name, default "public" schema would be used. Thereby "table_name" and "public.table_name" are equal.

## Usage

There are several ways to use this package:

1. Create new DTO Generator using `gorep.NewDtoGenerator()`, which has `Generate()` method to parse database 
and create DTO contents string. Then this string could be saved to file.

2. Create new application to call from command line or go:generate.

First, create file `dto_generator.go` with main function:

```go
package main

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/vehsamrak/gorep"
)

const (
	// use your database credentials
	databaseHost     = "localhost"
	databasePort     = 5432
	databaseUser     = "user"
	databasePassword = "password"
	databaseName     = "database_name"
)

func main() {
	// pass mandatory parameters as command line arguments
	packageName := os.Args[1]
	filename := os.Args[2]
	tableName := os.Args[3]

	// connect to database
	database, err := sqlx.Connect(
		"postgres", fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			databaseHost, databasePort, databaseUser, databasePassword, databaseName,
		),
	)
	if err != nil {
		panic(err)
	}

	// use DtoGenerator to create DTO contents string
	generatedDtoContents, err := gorep.NewDtoGenerator(database).Generate(packageName, tableName)
	if err != nil {
		panic(err)
	}

	// write DTO contents to file
	err = os.WriteFile(filename, []byte(generatedDtoContents), 0666)
	if err != nil {
		panic(err)
	}
}
```

Now it could be called with: `go run dto_generator main example_dto.go tablename`.

Or with `go:generate`:
```go
package main

//go:generate go run dto_generator main example_dto.go tablename
```

Output of generated `example_dto.go` now contains DTO structure, named according to table name.
Structure has parameters generated from database columns with their names and "db" tags for database mapping.
DTO properties has their own types, with respect for database nullables.
Generated file would look something like this:
```go
// Code was generated by GoRep. Please do not modify it!

package main

import (
    "database/sql"
    "time"
)

type TablenameDTO struct {
	Id int64 `db:"id"`
	Name sql.NullString `db:"name"`
	Description sql.NullString `db:"description"`
	StartTime time.Time `db:"start_time"`
	FinishTime sql.NullTime `db:"finish_time"`
	WasApproved bool `db:"was_approved"`
}
```

### Dependencies
* jmoiron/sqlx - to create DTO from database table
* mattn/go-sqlite3 - to perform SQL tests
* ory/dockertest - to test databases with docker
