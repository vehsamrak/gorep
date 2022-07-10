package gorep

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var testDatabase *sqlx.DB

func TestMain(m *testing.M) {
	const (
		maxDockerWaitSeconds = 120
	)

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "11",
			Env: []string{
				"POSTGRES_PASSWORD=secret",
				"POSTGRES_USER=user_name",
				"POSTGRES_DB=dbname",
				"listen_addresses = '*'",
			},
		}, func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		panic(fmt.Errorf("[docker_test] could not start resource: %w", err))
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	err = resource.Expire(maxDockerWaitSeconds)
	if err != nil {
		panic(fmt.Errorf("[docker_test] expiration error: %w", err))
	}

	pool.MaxWait = maxDockerWaitSeconds * time.Second
	if err = pool.Retry(
		func() error {
			testDatabase, err = sqlx.Connect("postgres", databaseUrl)

			return err
		},
	); err != nil {
		panic(fmt.Errorf("[create_database] error: %w", err))
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("[docker_test] could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestDtoGenerator_Generate(t *testing.T) {
	const (
		packageName                             = "package_name"
		tableName                               = "public.test"
		testDtoGoldenExampleFilePath            = "test_data/test_dto.golden"
		testDtoWithImportsGoldenExampleFilePath = "test_data/test_dto_with_imports.golden"
	)

	expectedDto, err := ioutil.ReadFile(testDtoGoldenExampleFilePath)
	if err != nil {
		t.Errorf("golden file reading error: %v", err)
	}
	expectedDtoWithImports, err := ioutil.ReadFile(testDtoWithImportsGoldenExampleFilePath)
	if err != nil {
		t.Errorf("golden file reading error: %v", err)
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
				database:    testDatabase,
				packageName: packageName,
				tableName:   tableName,
			},
			mockBehaviour: func() {
				createTable(
					testDatabase, tableName, map[string]string{
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
				database:    testDatabase,
				packageName: packageName,
				tableName:   tableName,
			},
			mockBehaviour: func() {
				createTable(
					testDatabase, tableName, map[string]string{
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
				database:    testDatabase,
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
				database:    testDatabase,
				packageName: packageName,
				tableName:   "",
			},
			mockBehaviour: func() {},
			expected:      "",
			expectedError: true,
		},
		{
			name: "table not exists, must return error",
			arguments: arguments{
				database:    testDatabase,
				packageName: packageName,
				tableName:   "non_existing_table",
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
