package integration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/vehsamrak/gorep"
	"github.com/vehsamrak/gorep/test_data"
)

func TestDtoGenerator_Generate_testDatabase(t *testing.T) {
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
		database    gorep.Database
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
						"id":    makeNotNullable(gorep.DatabaseFieldTypeInt),
						"value": makeNotNullable(gorep.DatabaseFieldTypeVarchar),
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
						"value_bigint":                 gorep.DatabaseFieldTypeBigint,
						"value_boolean":                gorep.DatabaseFieldTypeBoolean,
						"value_date":                   gorep.DatabaseFieldTypeDate,
						"value_decimal":                gorep.DatabaseFieldTypeDecimal,
						"value_double_precision":       gorep.DatabaseFieldTypeDoublePrecision,
						"value_float":                  gorep.DatabaseFieldTypeFloat,
						"value_int":                    gorep.DatabaseFieldTypeInt,
						"value_int2":                   gorep.DatabaseFieldTypeInt2,
						"value_int8":                   gorep.DatabaseFieldTypeInt8,
						"value_integer":                gorep.DatabaseFieldTypeInteger,
						"value_numeric":                gorep.DatabaseFieldTypeNumeric,
						"value_real":                   gorep.DatabaseFieldTypeReal,
						"value_serial":                 gorep.DatabaseFieldTypeSerial,
						"value_smallint":               gorep.DatabaseFieldTypeSmallint,
						"value_text":                   gorep.DatabaseFieldTypeText,
						"value_timestamp":              gorep.DatabaseFieldTypeTimestamp,
						"value_timestamp_not_nullable": makeNotNullable(gorep.DatabaseFieldTypeTimestamp),
						"value_varchar":                gorep.DatabaseFieldTypeVarchar,
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

				generator := gorep.NewDtoGenerator(tt.arguments.database)
				result, err := generator.Generate(tt.arguments.packageName, tt.arguments.tableName)

				if (err != nil) != tt.expectedError {
					t.Errorf("Generate() error: %v, expected error: %v", err, tt.expectedError)
					return
				}
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestDtoGenerator_Generate_mockDatabase(t *testing.T) {
	const (
		packageName = "package_name"
		tableName   = "public.test"
	)

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockDatabase := test_data.NewMockDatabase(mockController)

	type arguments struct {
		database    gorep.Database
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
			name: "database query error, must return error",
			arguments: arguments{
				database:    mockDatabase,
				packageName: packageName,
				tableName:   tableName,
			},
			mockBehaviour: func() {
				mockDatabase.EXPECT().Query(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			},
			expected:      "",
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.mockBehaviour()

				generator := gorep.NewDtoGenerator(tt.arguments.database)
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

func TestDtoGenerator_Generate_InvalidTemplate(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run(
		"invalid template file, must return parse error", func(t *testing.T) {
			const (
				invalidTemplateContents = "{{}}"
			)
			database := test_data.NewMockDatabase(mockController)
			expectedErrorMessage := "template: dto.template:1: missing value for command"
			generator := gorep.NewDtoGenerator(database)
			generator.SetTemplate(invalidTemplateContents)

			_, err := generator.Generate("package_name", "table_name")

			if err.Error() != expectedErrorMessage {
				t.Errorf("Generate() must return error \"%s\", returned \"%s\"", expectedErrorMessage, err)
			}
		},
	)

	t.Run(
		"invalid template file, must return execution error", func(t *testing.T) {
			const (
				tableName               = "test"
				packageName             = "package_name"
				invalidTemplateContents = "{{ .nonexistent }}"
			)
			dropTable(testDatabase, tableName)
			createTable(
				testDatabase, tableName, map[string]string{
					"id": gorep.DatabaseFieldTypeSerial,
				},
			)
			expectedErrorMessage := "can't evaluate field nonexistent"
			generator := gorep.NewDtoGenerator(testDatabase)
			generator.SetTemplate(invalidTemplateContents)

			_, err := generator.Generate(packageName, tableName)

			if !strings.Contains(err.Error(), expectedErrorMessage) {
				t.Errorf("Generate() must return error \"%s\", returned \"%s\"", expectedErrorMessage, err)
			}
		},
	)
}

func makeNotNullable(typeName string) string {
	return fmt.Sprintf("%s NOT NULL", typeName)
}
