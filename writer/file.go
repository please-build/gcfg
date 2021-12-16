package writer

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/please-build/gcfg/ast"
)

// Read file into a file struct
func readIntoStruct(file *os.File) ast.File {
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
			currentSection = f.Sections[len(f.Sections)-1].Key
			f.NumSections += 1
			f.Lines += 1
		} else if strings.Contains(line, "=") {
			// Append field to the current section
			for i, s := range f.Sections {
				if s.Key == currentSection {
					f.Sections[i].Fields = append(f.Sections[i].Fields, ast.MakeField(line))
				}
			}
			f.Lines += 1
		} else {
			f.Lines += 1
		}
	}

	return f
}

func injectField(f ast.File, field ast.Field, section string) ast.File {
	// Read the file so we know where to inject
	section = strings.ToLower(section)

	// Increment total lines in file
	f.Lines += 1

	// Does the section exist?
	exists := false
	for i, s := range f.Sections {
		if s.Key == section {
			exists = true
			f.Sections[i].Fields = append(f.Sections[i].Fields, field)
		}
	}

	if exists {
		//astSection.Fields = append(astSection.Fields, field)
	} else {
		// Create new section
	}

	return f
}

func convertASTToBytes(f ast.File) []byte {
	// Turn AST into a list of bytes
	var data []byte
	for _, section := range f.Sections {
		data = append(data, section.ToBytes()...)
		for _, field := range section.Fields {
			data = append(data, field.ToBytes()...)
		}
	}

	return data
}

func writeBytesToFile(data []byte, output string) {
	err := os.WriteFile(output, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write bytes to file: %v", err)
	}
}
