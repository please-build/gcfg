package writer

import (
	"bufio"
	"go/ast"
	"log"
	"os"
)

// Read file into a file struct
func readIntoStruct(in string) *ast.File {
	file, err := os.Open(in)
	if err != nil {
		log.Fatalf("Couldn't open file")
	}
	defer file.Close()

	var myfile File
	myfile.name = in

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text()[0] == '[' {
			myfile.sections = append(myfile.sections, scanner.Text())
			myfile.numSections += 1
		}
	}

	return &myfile
}
