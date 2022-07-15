# Gorep - Golang repository generator

![CI](https://github.com/vehsamrak/gorep/actions/workflows/main.yml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/vehsamrak/gorep/branch/main/graph/badge.svg?token=1wSNzO0Ds1)](https://codecov.io/gh/vehsamrak/gorep)
[![DeepSource](https://deepsource.io/gh/vehsamrak/gorep.svg/?label=active+issues&show_trend=true&token=iYOMH3keC-KA6pmDSfrsqQ0V)](https://deepsource.io/gh/vehsamrak/gorep/)

Tired of hand-writing DTO's and SQL for accessing database data?
Gorep aims to automate this tedious and repeatable task for you, using boilerplate code generation.

It can:

* Generate DTO structures from database tables
* Generate Model classes with constructor, closed fields and opened getters from DTOs
* **[not ready for now]** Generate Repository classes with fetching methods containing SQL and DTO-mapping boilerplate.

### What is repository

Repository mediates between the domain and data mapping layers using a collection-like interface 
for accessing domain objects. (c) [Martin Fowler](https://martinfowler.com/eaaCatalog/repository.html)

### Supported databases
Only **PostgreSQL** is supported for now.
Default schema is "public", which could be changed by prefixing table name with schema name. For example, to fetch
table named "table_name" and schema "schema_name" - you should pass "schema_name.table_name" as table name. If no
prefix set to table name, then default "public" schema would be used. Thereby "table_name" and "public.table_name"
are equal.

## Usage

There are several ways to use this package:

1. Create new DTO Generator using `gorep.NewDtoGenerator()`, which has `Generate()` method to parse database
   and create DTO contents string. Then this string could be saved to file.

2. Create new Model Generator using `gorep.NewModelGenerator()`, which also has `Generate()` method to parse DTO
   file and create model contents string.

3. Create new application to call from command line or go:generate.

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
	dtoFilename := os.Args[2]
	modelFilename := os.Args[3]
	tableName := os.Args[4]

	// connect to database
	database, _ := sqlx.Connect(
		"postgres", fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			databaseHost, databasePort, databaseUser, databasePassword, databaseName,
		),
	)

	// use DtoGenerator to create DTO contents string
	generatedDtoContents, _ := gorep.NewDtoGenerator(database).Generate(packageName, tableName)

	// write DTO contents to file
	os.WriteFile(dtoFilename, []byte(generatedDtoContents), 0666)

	// use ModelGenerator to create Model contents string from DTO file
	generatedModelContents, _ := gorep.NewModelGenerator().Generate(packageName, generatedDtoContents)

	// write Model contents to file
	os.WriteFile(modelFilename, []byte(generatedModelContents), 0666)
}
```

Now it could be called with: `go run dto_generator package_name example_dto.go example_model.go tablename`.

Or with `go:generate`:

```go
package main

//go:generate go run dto_generator package_name example_dto.go example_model.go tablename
```

Output of generated `example_dto.go` now contains DTO structure, named according to table name.
Structure has parameters generated from database columns with their names and "db" tags for database mapping.
DTO properties has their own types, with respect for database nullables.

Generated file for DTO would look something like this:

```go
// Code was generated by GoRep. Please do not modify it!

package package_name

import (
	"database/sql"
	"time"
)

type TablenameDTO struct {
	Id          int64          `db:"id"`
	Name        sql.NullString `db:"name"`
	Description sql.NullString `db:"description"`
	StartTime   time.Time      `db:"start_time"`
	FinishTime  sql.NullTime   `db:"finish_time"`
	WasApproved bool           `db:"was_approved"`
}
```

Generated file for Model would be:

```go
// Code was generated by GoRep. Please do not modify it!

package package_name

import (
	"database/sql"
	"time"
)

type Tablename struct {
	description sql.NullString
	finishTime  sql.NullTime
	id          int64
	name        sql.NullString
	startTime   time.Time
	wasApproved bool
}

func NewTablename(
	description sql.NullString,
	finishTime sql.NullTime,
	id int64,
	name sql.NullString,
	startTime time.Time,
	wasApproved bool,
) *Tablename {
	return &Tablename{
		description: description,
		finishTime:  finishTime,
		id:          id,
		name:        name,
		startTime:   startTime,
		wasApproved: wasApproved,
	}
}

func (m *Tablename) Description() sql.NullString {
	return m.description
}

func (m *Tablename) FinishTime() sql.NullTime {
	return m.finishTime
}

func (m *Tablename) Id() int64 {
	return m.id
}

func (m *Tablename) Name() sql.NullString {
	return m.name
}

func (m *Tablename) StartTime() time.Time {
	return m.startTime
}

func (m *Tablename) WasApproved() bool {
	return m.wasApproved
}
```

### Dependencies

* jmoiron/sqlx - to create DTO from database table
* mattn/go-sqlite3 - to perform SQL tests
* ory/dockertest - to test databases with docker
