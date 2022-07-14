package gorep

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vehsamrak/gorep/test_tools"
)

func TestRepositoryGenerator_Generate(t *testing.T) {
	const (
		packageName            = "package_name"
		dtoFileNameNotGolang   = "test_data/test_dto.golden"
		dtoFileName            = "test_data/test_dto.go"
		repositoryContentsFile = "test_data/repository.golden"
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
			fileContents:  test_tools.GetFileContents(dtoFileNameNotGolang),
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
			fileContents:  test_tools.GetFileContents(dtoFileNameNotGolang),
			packageName:   packageName,
			expected:      "",
			expectedError: "file must be a valid go file",
		},
		{
			name:          "file is valid golang file with DTO structure, must return repository contents",
			fileContents:  test_tools.GetFileContents(dtoFileName),
			packageName:   packageName,
			expected:      test_tools.GetFileContents(repositoryContentsFile),
			expectedError: "",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				generator := NewRepositoryGenerator()

				result, err := generator.Generate(tt.fileContents, tt.packageName)

				if tt.expectedError == "" {
					assert.Nil(t, err)
				} else {
					assert.ErrorContains(t, err, tt.expectedError)
				}
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}
