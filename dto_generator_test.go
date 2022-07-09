package gorep

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestDtoGenerator_Generate(t *testing.T) {
	const (
		tableName                               = "test"
		packageName                             = "package_name"
		testDtoGoldenExampleFilePath            = "test_data/test_dto.golden"
		testDtoWithImportsGoldenExampleFilePath = "test_data/test_dto_with_imports.golden"
	)

	database := createDatabase("test_data/database.db")
	defer database.Close()

	expectedDto, err := ioutil.ReadFile(testDtoGoldenExampleFilePath)
	if err != nil {
		t.Errorf("Generate() reading golden file error: %v", err)
	}
	expectedDtoWithImports, err := ioutil.ReadFile(testDtoWithImportsGoldenExampleFilePath)
	if err != nil {
		t.Errorf("Generate() reading golden file error: %v", err)
	}

	type arguments struct {
		database    *sqlx.DB
		tableName   string
		packageName string
	}
	tests := []struct {
		name          string
		arguments     arguments
		mockBehaviour func()
		expected      string
		expectedError bool
	}{
		{
			name: "table with int and varchar fields, must return correct DTO as string",
			arguments: arguments{
				database:    database,
				packageName: packageName,
				tableName:   tableName,
			},
			mockBehaviour: func() {
				createTable(database, tableName, map[string]string{"id": "int", "value": "varchar"})
			},
			expected:      string(expectedDto),
			expectedError: false,
		},
		{
			name: "table with fields and import types, must return correct DTO as string",
			arguments: arguments{
				database:    database,
				packageName: packageName,
				tableName:   tableName,
			},
			mockBehaviour: func() {
				createTable(
					database, tableName, map[string]string{
						"value_int":               "int",
						"value_integer":           "integer",
						"value_tinyint":           "tinyint",
						"value_smallint":          "smallint",
						"value_mediumint":         "mediumint",
						"value_bigint":            "bigint",
						"value_unsigned_big_int":  "unsigned big int",
						"value_int2":              "int2",
						"value_int8":              "int8",
						"value_character":         "character",
						"value_varchar":           "varchar",
						"value_varying_character": "varying character",
						"value_blob":              "blob",
						"value_text":              "text",
						"value_real":              "real",
						"value_double":            "double",
						"value_double_precision":  "double precision",
						"value_float":             "float",
						"value_numeric":           "numeric",
						"value_decimal":           "decimal",
						"value_boolean":           "boolean",
						"value_date":              "date",
						"value_datetime":          "datetime",
						"value_timestamp":         "timestamp",
					},
				)
			},
			expected:      string(expectedDtoWithImports),
			expectedError: false,
		},
		{
			name: "empty package name, must return error",
			arguments: arguments{
				database:    database,
				packageName: "",
				tableName:   tableName,
			},
			mockBehaviour: func() {},
			expected:      "",
			expectedError: true,
		},
		{
			name: "empty table name, must return error",
			arguments: arguments{
				database:    database,
				packageName: packageName,
				tableName:   "",
			},
			mockBehaviour: func() {},
			expected:      "",
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				dropAllTables(tt.arguments.database)
				tt.mockBehaviour()

				generator := NewDtoGenerator(tt.arguments.database)
				result, err := generator.Generate(tt.arguments.packageName, tt.arguments.tableName)

				if (err != nil) != tt.expectedError {
					t.Errorf("Generate() error: %v, expected error: %v", err, tt.expectedError)
					return
				}

				if result != tt.expected {
					t.Errorf(
						"Generate() result is not as expected:\n%v",
						diff.LineDiff(result, tt.expected),
					)
				}
			},
		)
	}
}

func createDatabase(databaseFile string) *sqlx.DB {
	database, err := sqlx.Connect("sqlite3", databaseFile)
	if err != nil {
		panic(fmt.Errorf("[create_database] error: %w", err))
	}

	return database
}

func createTable(database *sqlx.DB, tableName string, columnsMap map[string]string) {
	if len(columnsMap) == 0 {
		panic(fmt.Errorf("[create_table] no columns specified for create table operation"))
	}

	columns := make([]string, 0, len(columnsMap))
	for columnName, columnType := range columnsMap {
		columns = append(columns, fmt.Sprintf("%s %s", columnName, columnType))
	}

	fff := fmt.Sprintf("create table %s (%s)", tableName, strings.Join(columns, ", "))
	_, err := database.Exec(fff)
	if err != nil {
		panic(fmt.Errorf("[create_table] error: %w", err))
	}
}

func dropAllTables(database *sqlx.DB) {
	rows, err := database.Query("SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// sqlite_sequence table could not be dropped in SQLite
	inapplicableTableNames := map[string]struct{}{
		"sqlite_sequence": {},
	}

	var tableNames []string
	for rows.Next() {
		var query string
		err := rows.Scan(&query)
		if err != nil {
			panic(fmt.Errorf("[drop_all_tables] error: %w", err))
		}

		tableNames = append(tableNames, query)
	}

	for _, tableName := range tableNames {
		if _, ok := inapplicableTableNames[tableName]; ok {
			continue
		}

		_, err = database.Exec(fmt.Sprintf("DROP TABLE %s", tableName))
		if err != nil {
			panic(fmt.Errorf("[drop_all_tables] error: %w", err))
		}
	}
}
