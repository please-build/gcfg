package ast

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strings"
)

// Read reads a file into an ast.File struct.
func Read(file io.Reader, name string) File {
	var f File
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
					f.Sections = append(f.Sections, MakeSection("[_preamble]"))
					currentSection = "_preamble"
				}
			}
			for i, s := range f.Sections {
				if s.Key == currentSection {
					f.Sections[i].Fields = append(f.Sections[i].Fields, MakeField(line))
				}
			}
			f.NumLines += 1
		} else if matched, err := regexp.MatchString(`^ *\[.*\] *$`, line); err != nil {
			log.Panicf("Error matching regexp: %v", err)
		} else if matched {
			f.Sections = append(f.Sections, MakeSection(line))
			currentSection = f.Sections[len(f.Sections)-1].Key
			f.NumLines += 1
		} else {
			f.NumLines += 1
		}
	}

	return f
}
