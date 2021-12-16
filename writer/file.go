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
		log.Printf("scanning line %v", f.Lines)
		line := scanner.Text()
		if line == "" {
			f.Lines += 1
		} else if line[0] == '[' && line[len(line)-1] == ']' {
			f.Sections = append(f.Sections, ast.MakeSection(line, f.Lines))
			// f.Sections[ast.GetSectionTitleFromString(line)] = ast.MakeSection(line, f.Lines)
			currentSection = f.Sections[len(f.Sections)-1].Key
			log.Printf("currentSection = \"%v\"", currentSection)
			f.NumSections += 1
			f.Lines += 1
		} else if strings.Contains(line, "=") {
			// This is a field, so append this to the current section
			log.Printf("Found field %v", line)
			log.Printf("looping through sections to find %v", currentSection)
			for i, s := range f.Sections {
				log.Printf("\ttrying section %v", s.Key)
				if s.Key == currentSection {
					log.Printf("section %v has %v fields", s.Key, len(f.Sections[i].Fields))
					f.Sections[i].Fields = append(f.Sections[i].Fields, ast.MakeField(line))
					log.Printf("appending field \"%v\" to section \"%v\"", line, s)
					log.Printf("so now section %v has %v fields", s.Key, len(f.Sections[i].Fields))
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
	fmt.Printf("in injectField\n")
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
		//fmt.Printf("astSection = %v\n", astSection)
		//astSection.Fields = append(astSection.Fields, field)
	} else {
		fmt.Printf("doesn't exist\n")
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
