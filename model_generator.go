package gorep

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"text/template"
)

//go:embed model.template
var templateFileModel string

type ModelGenerator struct {
	templateModel string
}

func NewModelGenerator() *ModelGenerator {
	return &ModelGenerator{templateModel: templateFileModel}
}

func (g *ModelGenerator) Generate(packageName string, dtoFileContents string) (string, error) {
	if dtoFileContents == "" {
		return "", fmt.Errorf("dto file contents must not be empty")
	}

	if packageName == "" {
		return "", fmt.Errorf("package name must not be empty")
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "file.go", dtoFileContents, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("dto file contents parsing error: %w", err)
	}

	var structName string
	ast.Inspect(
		file, func(astNode ast.Node) bool {
			astTypeSpec, ok := astNode.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structName = g.removeDTOFromStructName(astTypeSpec.Name.Name)

			return false
		},
	)

	var modelFields []modelField
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

				fieldType := dtoFileContents[(field.Type.Pos() - 1):(field.Type.End() - 1)]

				modelFields = append(
					modelFields, modelField{
						Name:       currentField.Name,
						Type:       fieldType,
						StructName: structName,
					},
				)
			}

			return false
		},
	)

	if structName == "" {
		return "", fmt.Errorf("no DTO structure was found in DTO contents")
	}

	// TODO[petr]: if model file not exist

	if len(modelFields) == 0 {
		return "", fmt.Errorf("no fields found in DTO")
	}

	sort.Slice(
		modelFields, func(i, j int) bool {
			return modelFields[i].Name < modelFields[j].Name
		},
	)

	imports := g.createImports(modelFields)

	data := struct {
		PackageName string
		StructName  string
		Fields      []modelField
		Imports     []string
	}{
		PackageName: packageName,
		StructName:  structName,
		Fields:      modelFields,
		Imports:     imports,
	}

	templator, err := template.New("model.template").
		Funcs(
			template.FuncMap{
				"Uppercase": StringCaseConverter{}.SnakeCaseToCamelCase,
				"Lowercase": StringCaseConverter{}.Lowercase,
			},
		).
		Parse(g.templateModel)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = templator.Execute(&buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil

	// TODO[petr]: if model file exist
}

func (g *ModelGenerator) createImports(fields []modelField) []string {
	importsMap := map[string]string{
		"time.Time":       "time",
		"[]sql.NullByte":  "database/sql",
		"sql.NullBool":    "database/sql",
		"sql.NullFloat64": "database/sql",
		"sql.NullString":  "database/sql",
		"sql.NullTime":    "database/sql",
		"sql.NullInt64":   "database/sql",
	}

	alreadyImported := make(map[string]struct{})
	var imports []string
	for _, field := range fields {
		if importPackage, ok := importsMap[field.Type]; ok {
			if _, ok := alreadyImported[importPackage]; !ok {
				imports = append(imports, importPackage)
				alreadyImported[importPackage] = struct{}{}
			}
		}
	}

	return imports
}

func (g *ModelGenerator) removeDTOFromStructName(name string) string {
	return strings.ReplaceAll(name, "DTO", "")
}
