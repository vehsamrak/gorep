package gorep

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func TestDtoGenerator_Generate(t *testing.T) {
	const (
		packageName                             = "package_name"
		tableName                               = "test"
		testDtoGoldenExampleFilePath            = "test_data/test_dto.golden"
		testDtoWithImportsGoldenExampleFilePath = "test_data/test_dto_with_imports.golden"
	)

	database := createPostgresDatabase()
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
				createTable(
					database, tableName, map[string]string{
						"id":    makeNotNullable(databaseFieldTypeInt),
						"value": makeNotNullable(databaseFieldTypeVarchar),
					},
				)
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
						"value_bigint":                 databaseFieldTypeBigint,
						"value_boolean":                databaseFieldTypeBoolean,
						"value_date":                   databaseFieldTypeDate,
						"value_decimal":                databaseFieldTypeDecimal,
						"value_double_precision":       databaseFieldTypeDoublePrecision,
						"value_float":                  databaseFieldTypeFloat,
						"value_int":                    databaseFieldTypeInt,
						"value_int2":                   databaseFieldTypeInt2,
						"value_int8":                   databaseFieldTypeInt8,
						"value_integer":                databaseFieldTypeInteger,
						"value_numeric":                databaseFieldTypeNumeric,
						"value_real":                   databaseFieldTypeReal,
						"value_serial":                 databaseFieldTypeSerial,
						"value_smallint":               databaseFieldTypeSmallint,
						"value_text":                   databaseFieldTypeText,
						"value_timestamp":              databaseFieldTypeTimestamp,
						"value_timestamp_not_nullable": makeNotNullable(databaseFieldTypeTimestamp),
						"value_varchar":                databaseFieldTypeVarchar,
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
				dropTable(tt.arguments.database, tableName)
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

func makeNotNullable(typeName string) string {
	return fmt.Sprintf("%s NOT NULL", typeName)
}

func createPostgresDatabase() *sqlx.DB {
	database, err := sqlx.Connect(
		"postgres", fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			"localhost", 5433, "2Vw0AAtV2svu", "010dYSkjHMlY", "mud2",
		),
	)
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

	query := fmt.Sprintf("create table %s (%s)", tableName, strings.Join(columns, ", "))
	_, err := database.Exec(query)
	if err != nil {
		panic(fmt.Errorf("[create_table] error: %w", err))
	}
}

func dropTable(database *sqlx.DB, tableName string) {
	_, err := database.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		panic(fmt.Errorf("[drop_table] error: %w", err))
	}
}
