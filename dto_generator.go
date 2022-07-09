package gorep

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/jmoiron/sqlx"
)

const (
	databaseFieldTypeBigint           = "bigint"
	databaseFieldTypeBlob             = "blob"
	databaseFieldTypeBoolean          = "boolean"
	databaseFieldTypeCharacter        = "character"
	databaseFieldTypeDate             = "date"
	databaseFieldTypeDatetime         = "datetime"
	databaseFieldTypeDecimal          = "decimal"
	databaseFieldTypeDouble           = "double"
	databaseFieldTypeDoublePrecision  = "double precision"
	databaseFieldTypeFloat            = "float"
	databaseFieldTypeInt              = "int"
	databaseFieldTypeInt2             = "int2"
	databaseFieldTypeInt8             = "int8"
	databaseFieldTypeInteger          = "integer"
	databaseFieldTypeMediumint        = "mediumint"
	databaseFieldTypeNumeric          = "numeric"
	databaseFieldTypeReal             = "real"
	databaseFieldTypeSmallint         = "smallint"
	databaseFieldTypeText             = "text"
	databaseFieldTypeTimestamp        = "timestamp"
	databaseFieldTypeTinyint          = "tinyint"
	databaseFieldTypeUnsignedBigInt   = "unsigned big int"
	databaseFieldTypeVarchar          = "varchar"
	databaseFieldTypeVaryingCharacter = "varying character"
)

type DtoGenerator struct {
	database *sqlx.DB
}

func NewDtoGenerator(database *sqlx.DB) *DtoGenerator {
	return &DtoGenerator{database: database}
}

// Generate generates DTO for tableName as file content string
func (g *DtoGenerator) Generate(packageName string, tableName string) (string, error) {
	if len(packageName) == 0 {
		return "", errors.New("package name must not be empty")
	}

	if len(tableName) == 0 {
		return "", errors.New("table name must not be empty")
	}

	funcMap := template.FuncMap{
		"UppercaseFirstLetter": g.uppercaseFirstLetter,
	}

	templator, err := template.New("dto.template").Funcs(funcMap).ParseFiles("dto.template")
	if err != nil {
		return "", err
	}

	fields, err := g.fetchFields(tableName)
	if err != nil {
		return "", err
	}

	sort.Slice(
		fields, func(i, j int) bool {
			return fields[i].Name > fields[j].Name
		},
	)

	imports := g.createImports(fields)

	data := struct {
		PackageName string
		TableName   string
		Fields      []databaseField
		Imports     []string
	}{
		PackageName: packageName,
		TableName:   g.uppercaseFirstLetter(tableName),
		Fields:      fields,
		Imports:     imports,
	}

	var buffer bytes.Buffer
	err = templator.Execute(&buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (g *DtoGenerator) fetchFields(tableName string) ([]databaseField, error) {
	rows, err := g.database.Queryx(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return nil, err
	}

	var fields []databaseField
	columnTypes, err := rows.ColumnTypes()
	for _, columnType := range columnTypes {
		fields = append(
			fields, databaseField{
				Name: columnType.Name(),
				Type: mapDatabaseType(columnType.DatabaseTypeName()),
			},
		)
	}

	return fields, nil
}

func mapDatabaseType(databaseTypeName string) string {
	typeMap := map[string]string{
		databaseFieldTypeBigint:           "int64",
		databaseFieldTypeBlob:             "[]byte",
		databaseFieldTypeBoolean:          "bool",
		databaseFieldTypeCharacter:        "string",
		databaseFieldTypeDate:             "time.Time",
		databaseFieldTypeDatetime:         "time.Time",
		databaseFieldTypeDecimal:          "float64",
		databaseFieldTypeDouble:           "float64",
		databaseFieldTypeDoublePrecision:  "float64",
		databaseFieldTypeFloat:            "float64",
		databaseFieldTypeInt2:             "int8",
		databaseFieldTypeInt8:             "int8",
		databaseFieldTypeInt:              "int64",
		databaseFieldTypeInteger:          "int64",
		databaseFieldTypeMediumint:        "int64",
		databaseFieldTypeNumeric:          "int64",
		databaseFieldTypeReal:             "float64",
		databaseFieldTypeSmallint:         "int64",
		databaseFieldTypeText:             "string",
		databaseFieldTypeTimestamp:        "time.Time",
		databaseFieldTypeTinyint:          "int64",
		databaseFieldTypeUnsignedBigInt:   "uint64",
		databaseFieldTypeVarchar:          "string",
		databaseFieldTypeVaryingCharacter: "string",
	}

	databaseTypeName = strings.ToLower(databaseTypeName)

	typeName, ok := typeMap[databaseTypeName]
	if !ok {
		typeName = "[]byte"
	}

	return typeName
}

func (g *DtoGenerator) uppercaseFirstLetter(text string) string {
	letters := []rune(text)
	letters[0] = unicode.ToUpper(letters[0])

	return string(letters)
}

func (g *DtoGenerator) createImports(fields []databaseField) []string {
	importsMap := map[string]string{
		"time.Time": "time",
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
