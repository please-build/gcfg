package writer

import (
	"bufio"
	"log"
	"os"

	"github.com/please-build/gcfg/ast"
)

// Read file into a file struct
func readIntoStruct(in string) *ast.File {
	file, err := os.Open(in)
	if err != nil {
		log.Fatalf("Couldn't open file")
	}
	defer file.Close()

	var myfile ast.File
	myfile.Name = in

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text()[0] == '[' {
			myfile.Sections = append(myfile.Sections, ast.MakeSection(scanner.Text()))
			myfile.NumSections += 1
		}
	}

	return &myfile
}
