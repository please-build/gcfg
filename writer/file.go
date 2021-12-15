package writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/please-build/gcfg/ast"
)

// Read file into a file struct
func readIntoStruct(file *os.File) *ast.File {
	var f ast.File
	f.Name = file.Name()

	ioreader, err := os.Open(f.Name)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(ioreader)

	//TODO: Check for duplicate sections and collapse/warn if found?

	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			f.Lines += 1
		} else if line[0] == '[' && line[len(line)-1] == ']' {
			f.Sections = append(f.Sections, ast.MakeSection(line, f.Lines))
			// f.Sections[ast.GetSectionTitleFromString(line)] = ast.MakeSection(line, f.Lines)
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
		} else {
			f.Lines += 1
		}
	}

	return &f
}

func injectField(f *ast.File, field ast.Field, section string) *ast.File {
	// Read the file so we know where to inject
	section = strings.ToLower(section)

	// Increment total lines in file
	f.Lines += 1

	// Does the section exist?
	exists := false
	var astSection ast.Section
	for _, s := range f.Sections {
		if s.Title == section {
			exists = true
			astSection = s
		}
	}

	if exists {
		fmt.Printf("astSection = %v", astSection)
		astSection.Fields = append(astSection.Fields, field)
	}

	return f
}

func writeASTToFile(f *ast.File, output string) *os.File {
	// Turn AST into a list of bytes
	for _, section := range f.Sections {
		for _, field := range section.Fields {

		}
	}
}
