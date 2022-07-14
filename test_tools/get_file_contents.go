package test_tools

import (
	"fmt"
	"io/ioutil"
)

func GetFileContents(filePath string) string {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Errorf("golden file reading error: %w", err))
	}

	return string(fileContents)
}
