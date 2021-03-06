package gorep

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

const (
	databaseFieldTypeBigint           = "bigint"
	databaseFieldTypeBlob             = "blob"
	databaseFieldTypeBoolean          = "boolean"
	databaseFieldTypeBool             = "bool"
	databaseFieldTypeCharacter        = "character"
	databaseFieldTypeDate             = "date"
	databaseFieldTypeDatetime         = "datetime"
	databaseFieldTypeDecimal          = "decimal"
	databaseFieldTypeDouble           = "double"
	databaseFieldTypeDoublePrecision  = "double precision"
	databaseFieldTypeFloat            = "float"
	databaseFieldTypeFloat4           = "float4"
	databaseFieldTypeFloat8           = "float8"
	databaseFieldTypeInt              = "int"
	databaseFieldTypeInt2             = "int2"
	databaseFieldTypeInt4             = "int4"
	databaseFieldTypeInt8             = "int8"
	databaseFieldTypeInteger          = "integer"
	databaseFieldTypeMediumint        = "mediumint"
	databaseFieldTypeNumeric          = "numeric"
	databaseFieldTypeReal             = "real"
	databaseFieldTypeSerial           = "serial"
	databaseFieldTypeSmallint         = "smallint"
	databaseFieldTypeText             = "text"
	databaseFieldTypeTimestamp        = "timestamp"
	databaseFieldTypeTinyint          = "tinyint"
	databaseFieldTypeUnsignedBigInt   = "unsigned big int"
	databaseFieldTypeVarchar          = "varchar"
	databaseFieldTypeVaryingCharacter = "varying character"
)

//go:embed dto.template
var templateFile string

type DtoGenerator struct {
	database    Database
	templateDTO string
}

func NewDtoGenerator(database Database) *DtoGenerator {
	return &DtoGenerator{database: database, templateDTO: templateFile}
}

// Generate generates DTO for dtoPath as file content string
func (g *DtoGenerator) Generate(packageName string, tableName string) (string, error) {
	if len(packageName) == 0 {
		return "", errors.New("package name must not be empty")
	}

	if len(tableName) == 0 {
		return "", errors.New("table name must not be empty")
	}

	templator, err := template.New("dto.template").
		Funcs(
			template.FuncMap{
				"Uppercase": StringCaseConverter{}.SnakeCaseToCamelCase,
			},
		).
		Parse(g.templateDTO)

	if err != nil {
		return "", err
	}

	fields, err := g.fetchFields(tableName)
	if err != nil {
		return "", err
	}

	if len(fields) == 0 {
		return "", errors.New("table was not found or has no columns")
	}

	sort.Slice(
		fields, func(i, j int) bool {
			return fields[i].Name < fields[j].Name
		},
	)

	imports := g.createImports(fields)

	_, tableName = g.parseSchemaAndTableName(tableName)

	data := struct {
		PackageName string
		TableName   string
		Fields      []databaseField
		Imports     []string
	}{
		PackageName: packageName,
		TableName:   tableName,
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
	schema, tableName := g.parseSchemaAndTableName(tableName)

	rows, err := g.database.Query(
		fmt.Sprintf(
			"SELECT column_name, udt_name, is_nullable FROM information_schema.columns WHERE table_schema = '%s' AND table_name = '%s'",
			schema,
			tableName,
		),
	)
	if err != nil {
		return nil, err
	}

	var fields []databaseField
	for rows.Next() {
		var columnName string
		var columnType string
		var isNullableData string
		err = rows.Scan(&columnName, &columnType, &isNullableData)
		if err != nil {
			return nil, err
		}

		var isNullable bool
		if isNullableData == "YES" {
			isNullable = true
		}

		databaseTypeName := columnType
		databaseTypeName = strings.ToLower(databaseTypeName)

		databaseType := g.mapDatabaseType(databaseTypeName)
		if isNullable {
			databaseType = g.mapNullableTypeName(databaseType)
		}

		fields = append(
			fields, databaseField{
				Name: columnName,
				Type: databaseType,
			},
		)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func (*DtoGenerator) parseSchemaAndTableName(tableName string) (string, string) {
	schema := "public"

	tableNameParts := strings.Split(tableName, ".")
	if len(tableNameParts) > 1 {
		schema = tableNameParts[0]
		tableName = tableNameParts[1]
	}

	return schema, tableName
}

func (*DtoGenerator) mapDatabaseType(databaseTypeName string) string {
	typeMap := map[string]string{
		databaseFieldTypeBigint:           "int64",
		databaseFieldTypeBlob:             "[]byte",
		databaseFieldTypeBoolean:          "bool",
		databaseFieldTypeBool:             "bool",
		databaseFieldTypeCharacter:        "string",
		databaseFieldTypeDate:             "time.Time",
		databaseFieldTypeDatetime:         "time.Time",
		databaseFieldTypeDecimal:          "float64",
		databaseFieldTypeDouble:           "float64",
		databaseFieldTypeDoublePrecision:  "float64",
		databaseFieldTypeFloat:            "float64",
		databaseFieldTypeFloat4:           "float64",
		databaseFieldTypeFloat8:           "float64",
		databaseFieldTypeInt2:             "int64",
		databaseFieldTypeInt4:             "int64",
		databaseFieldTypeInt8:             "int64",
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

func (*DtoGenerator) mapNullableTypeName(typeName string) string {
	nullableTypeName, ok := map[string]string{
		"[]byte":    "[]sql.NullByte",
		"bool":      "sql.NullBool",
		"float64":   "sql.NullFloat64",
		"int64":     "sql.NullInt64",
		"string":    "sql.NullString",
		"time.Time": "sql.NullTime",
		"uint64":    "sql.NullInt64",
	}[typeName]
	if ok {
		typeName = nullableTypeName
	}

	return typeName
}

func (*DtoGenerator) createImports(fields []databaseField) []string {
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
