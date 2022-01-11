package writer

import (
	"bufio"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/please-build/gcfg/ast"
)

// Inject a field into the ast
func InjectField(f ast.File, field ast.Field, section string, repeatable bool) ast.File {
	// Read the file so we know where to inject
	sectionKey := ast.GetSectionKeyFromString(section)

	// Format field correctly for injection
	if !(strings.HasSuffix(field.Key, " ")) {
		field.Key += " "
	}
	if !(strings.HasPrefix(field.Value, " ")) {
		field.Value = " " + field.Value
	}

	// If section exists, add field to section
	for i, s := range f.Sections {
		if s.Key == sectionKey {
			if repeatable {
				f.Sections[i].Fields = append(f.Sections[i].Fields, field)
				f.NumLines += 1
				return f
			}
			for j, k := range s.Fields {
				if k.Key == field.Key {
					f.Sections[i].Fields[j].Value = field.Value
					return f
				}
			}
			f.Sections[i].Fields = append(f.Sections[i].Fields, field)
			f.NumLines += 1
			return f
		}
	}

	// Couldn't find section so create new one
	// Check if file currently ends with a blank line first
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

	header := ast.MakeSectionHeader(section)
	astSection := ast.MakeSection(header)
	astSection.Fields = append(astSection.Fields, field)
	f.Sections = append(f.Sections, astSection)

	return f
}

func readFile(file *os.File) ast.File {
	return read(file, file.Name())
}

// Read file into a file struct
func read(file io.Reader, name string) ast.File {
	var f ast.File
	f.Name = name
	scanner := bufio.NewScanner(file)

	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") || line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			// Append field to the current section
			if currentSection == "" || currentSection == "_preamble" {
				if len(f.Sections) == 0 {
					// If we're in here, then we're picking up some preamble
					// which could be blank lines, comments, or something else.
					f.Sections = append(f.Sections, ast.MakeSection("[_preamble]"))
					currentSection = "_preamble"
				}
			}
			for i, s := range f.Sections {
				if s.Key == currentSection {
					f.Sections[i].Fields = append(f.Sections[i].Fields, ast.MakeField(line))
				}
			}
			f.NumLines += 1
		} else if matched, err := regexp.MatchString(`^ *\[.*\] *$`, line); err != nil {
			log.Panicf("Error matching regexp: %v", err)
		} else if matched {
			f.Sections = append(f.Sections, ast.MakeSection(line))
			currentSection = f.Sections[len(f.Sections)-1].Key
			f.NumLines += 1
		} else {
			f.NumLines += 1
		}
	}

	return f
}

func convertASTToBytes(f ast.File) []byte {
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
