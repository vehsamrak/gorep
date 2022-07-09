package gorep

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"github.com/jmoiron/sqlx"
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

	data := struct {
		PackageName string
		TableName   string
		Fields      []databaseField
	}{
		PackageName: packageName,
		TableName:   g.uppercaseFirstLetter(tableName),
		Fields:      fields,
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
		"int":     "int64",
		"varchar": "string",
	}

	databaseTypeName = strings.ToLower(databaseTypeName)

	typeName, ok := typeMap[databaseTypeName]
	if !ok {
		typeName = "string"
	}

	return typeName
}

func (g *DtoGenerator) uppercaseFirstLetter(text string) string {
	letters := []rune(text)
	letters[0] = unicode.ToUpper(letters[0])

	return string(letters)
}
