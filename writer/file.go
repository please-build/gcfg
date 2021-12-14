package writer

import (
	"bufio"
	"log"
	"os"
	"strings"

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

	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			myfile.Lines += 1
		} else if line[0] == '[' && line[len(line)-1] == ']' {
			myfile.Sections = append(myfile.Sections, ast.MakeSection(line, myfile.Lines))
			currentSection = myfile.Sections[len(myfile.Sections)-1].Title
			myfile.NumSections += 1
			myfile.Lines += 1
		} else if strings.Contains(line, "=") {
			// This is a field, so append this to the current section
			myfile.Lines += 1
		}
	}

	return &myfile
}
