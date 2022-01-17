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
	f.name = name
	scanner := bufio.NewScanner(file)

	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") || line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			// Append field to the current section
			if currentSection == "" || currentSection == "_preamble" {
				if len(f.sections) == 0 {
					// If we're in here, then we're picking up some preamble
					// which could be blank lines, comments, or something else.
					f.sections = append(f.sections, makeSection("_preamble", ""))
					currentSection = "_preamble"
				}
			}
			for i, s := range f.sections {
				if s.key == currentSection {
					f.sections[i].fields = append(f.sections[i].fields, makeFieldFromString(line))
				}
			}
		} else if matched, err := regexp.MatchString(`^ *\[.*\] *$`, line); err != nil {
			log.Panicf("Error matching regexp: %v", err)
		} else if matched {
			// Matched a section title
			f.sections = append(f.sections, makeSectionFromString(line))
			currentSection = f.sections[len(f.sections)-1].key
		}
	}

	return f
}
