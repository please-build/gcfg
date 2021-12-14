package writer

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/please-build/gcfg/ast"
)

// Read file into a file struct
func readIntoStruct(in *os.File) *ast.File {
	var f ast.File
	f.Name = in.Name()

	ioreader, err := os.Open(f.Name)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(ioreader)

	//TODO: Check for duplicate sections with a map, and collapse if found...?

	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			f.Lines += 1
		} else if line[0] == '[' && line[len(line)-1] == ']' {
			f.Sections = append(f.Sections, ast.MakeSection(line, f.Lines))
			currentSection = f.Sections[len(f.Sections)-1].Title
			f.NumSections += 1
			f.Lines += 1
		} else if strings.Contains(line, "=") {
			// This is a field, so append this to the current section
			for _, i := range f.Sections {
				if i.Title == currentSection {
					i.Fields = append(i.Fields, ast.MakeField(line))
				}
			}
			f.Lines += 1
		}
	}

	return &f
}
