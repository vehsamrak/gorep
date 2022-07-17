package gorep

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/vehsamrak/gorep/tests/test_tools"
)

func TestModelGenerator_Generate(t *testing.T) {
	const (
		packageName                 = "package_name"
		fileNameNotGolang           = "test_data/non_golang_file.txt"
		fileNameWithoutStruct       = "test_data/dto_without_structure.test"
		fileNameWithoutStructFields = "test_data/dto_without_struct_fields.test"
		fileNameDto                 = "test_data/test_dto.go"
		modelFileContents           = "test_data/test_model.golden"
	)
	tests := []struct {
		name          string
		fileContents  string
		packageName   string
		expected      string
		expectedError string
	}{
		{
			name:          "file is valid golang file with DTO structure, must return model file contents",
			fileContents:  test_tools.GetFileContents(fileNameDto),
			packageName:   packageName,
			expected:      test_tools.GetFileContents(modelFileContents),
			expectedError: "",
		},
		{
			name:          "package name is empty, must return error",
			fileContents:  test_tools.GetFileContents(modelFileContents),
			packageName:   "",
			expected:      "",
			expectedError: "package name must not be empty",
		},
		{
			name:          "dto contents is empty, must return error",
			fileContents:  "",
			packageName:   packageName,
			expected:      "",
			expectedError: "dto file contents must not be empty",
		},
		{
			name:          "file is not golang file, must return error",
			fileContents:  test_tools.GetFileContents(fileNameNotGolang),
			packageName:   packageName,
			expected:      "",
			expectedError: "dto file contents parsing error",
		},
		{
			name:          "file without structure, must return error",
			fileContents:  test_tools.GetFileContents(fileNameWithoutStruct),
			packageName:   packageName,
			expected:      "",
			expectedError: "no DTO structure was found in DTO contents",
		},
		{
			name:          "file without structure, must return error",
			fileContents:  test_tools.GetFileContents(fileNameWithoutStructFields),
			packageName:   packageName,
			expected:      "",
			expectedError: "no fields found in DTO",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				generator := NewModelGenerator()

				result, err := generator.Generate(tt.packageName, tt.fileContents)

				if tt.expectedError == "" {
					assert.Nil(t, err, err)
				} else {
					assert.ErrorContains(t, err, tt.expectedError)
				}
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestModelGenerator_Generate_InvalidTemplate(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run(
		"invalid template file, must return parse error", func(t *testing.T) {
			const (
				packageName             = "package_name"
				fileNameDto             = "test_data/test_dto.go"
				invalidTemplateContents = "{{}}"
			)
			fileContents := test_tools.GetFileContents(fileNameDto)
			expectedErrorMessage := "template: model.template:1: missing value for command"
			generator := NewModelGenerator()
			generator.templateModel = invalidTemplateContents

			_, err := generator.Generate(packageName, fileContents)

			assert.ErrorContains(t, err, expectedErrorMessage)
		},
	)

	t.Run(
		"invalid template file, must return execution error", func(t *testing.T) {
			const (
				packageName             = "package_name"
				fileNameDto             = "test_data/test_dto.go"
				invalidTemplateContents = "{{ .nonexistent }}"
			)
			fileContents := test_tools.GetFileContents(fileNameDto)
			expectedErrorMessage := "can't evaluate field nonexistent"
			generator := NewModelGenerator()
			generator.templateModel = invalidTemplateContents

			_, err := generator.Generate(packageName, fileContents)

			assert.ErrorContains(t, err, expectedErrorMessage)
		},
	)
}
