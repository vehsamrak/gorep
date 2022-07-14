package gorep

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vehsamrak/gorep/test_tools"
)

func TestModelGenerator_Generate(t *testing.T) {
	const (
		packageName       = "package_name"
		fileNameNotGolang = "test_data/non_golang_file.txt"
		fileNameDto       = "test_data/test_dto.go"
		modelFileContents = "test_data/test_model.golden"
	)
	tests := []struct {
		name          string
		fileContents  string
		packageName   string
		expected      string
		expectedError string
	}{
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
			name:          "file is valid golang file with DTO structure, must return model file contents",
			fileContents:  test_tools.GetFileContents(fileNameDto),
			packageName:   packageName,
			expected:      test_tools.GetFileContents(modelFileContents),
			expectedError: "",
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
