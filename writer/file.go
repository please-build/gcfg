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
		if strings.Contains(line, "=") || line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			// Append field to the current section
			if currentSection == "" || currentSection == "_preamble" {
				if len(f.Sections) == 0 {
					// If we're in here, then we're picking up some preamble
					// which could be blank lines, or comments, or something else.
					f.Sections = append(f.Sections, ast.MakeSection("[_preamble]", f.NumLines))
					currentSection = "_preamble"
				}
			}
			for i, s := range f.Sections {
				if s.Key == currentSection {
					f.Sections[i].Fields = append(f.Sections[i].Fields, ast.MakeField(line))
				}
			}
			f.NumLines += 1
		} else if line[0] == '[' && line[len(line)-1] == ']' {
			f.Sections = append(f.Sections, ast.MakeSection(line, f.NumLines))
			currentSection = f.Sections[len(f.Sections)-1].Key
			f.NumLines += 1
		} else {
			f.NumLines += 1
		}
	}

	return f
}

func injectField(f ast.File, field ast.Field, section string) ast.File {
	// Read the file so we know where to inject
	sectionKey := ast.GetSectionKeyFromString(section)

	// Format field correctly for injection
	if !(strings.HasSuffix(field.Key, " ")) {
		field.Key += " "
	}
	if !(strings.HasPrefix(field.Value, " ")) {
		field.Value = " " + field.Value
	}

	// Does the section exist?
	exists := false
	for i, s := range f.Sections {
		if s.Key == sectionKey {
			exists = true
			f.Sections[i].Fields = append(f.Sections[i].Fields, field)
			f.NumLines += 1
		}
	}

	if !exists {
		log.Printf("This section doesn't exist. Need to create a new one")
		// Check if file currently ends with a blank line
		needToAppendSpace := true
		lastSection := &f.Sections[len(f.Sections)-1]
		if len(lastSection.Fields) > 0 {
			lastField := &lastSection.Fields[len(lastSection.Fields)-1]
			if lastField.IsBlankLine() {
				needToAppendSpace = false
			}
		}

		if needToAppendSpace {
			lastSection.Fields = append(lastSection.Fields, ast.Field{Key: "", Value: ""})
		}

		n := 4
		header := ast.MakeSectionHeader(section)
		astSection := ast.MakeSection(header, n)
		astSection.Fields = append(astSection.Fields, field)
		f.Sections = append(f.Sections, astSection)
	}

	return f
}

func convertASTToBytes(f ast.File) []byte {
	// Turn AST into a list of bytes
	var data []byte
	for _, section := range f.Sections {
		if section.Key != "_preamble" {
			data = append(data, section.ToBytes()...)
		}
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
