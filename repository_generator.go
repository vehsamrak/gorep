package gorep

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type RepositoryGenerator struct {
}

func NewRepositoryGenerator() *RepositoryGenerator {
	return &RepositoryGenerator{}
}

func (g *RepositoryGenerator) Generate(dtoFileContents string, packageName string) (string, error) {
	if dtoFileContents == "" {
		return "", fmt.Errorf("dto file contents must not be empty")
	}

	if packageName == "" {
		return "", fmt.Errorf("package name must not be empty")
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "file.go", dtoFileContents, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("dto file contents parsing error: %w", err)
	}

	ast.Inspect(
		file, func(astNode ast.Node) bool {
			astStruct, ok := astNode.(*ast.StructType)
			if !ok {
				return true
			}

			for _, field := range astStruct.Fields.List {
				currentField := field.Names[0]
				if !currentField.IsExported() {
					continue
				}

				// fieldType := dtoFileContents[(field.Type.Pos() - 1):(field.Type.End() - 1)]
			}

			return false
		},
	)

	// TODO[petr]: if repository file not exist
	// TODO[petr]: if repository file exist

	return "", fmt.Errorf("file must be a valid go file")
}
