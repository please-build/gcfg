package ast

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strings"
)

func Read(file io.Reader) File {
	var f File
	scanner := bufio.NewScanner(file)
	commentBuffer := make([]*Comment, 0, 64)
	fieldBuffer := make([]*Field, 0, 64)
	fieldRegex := regexp.MustCompile(`^ *[a-zA-Z0-9_ .-]+ *= *[a-zA-Z0-9_ /.-]`)
	sectionRegex := regexp.MustCompile(`^ *\[[a-zA-Z0-9_" .-]+\]`)
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 { // Blank line
			commentBuffer = append(commentBuffer, &Comment{})

		} else if l := strings.TrimSpace(line); strings.HasPrefix(l, ";") || strings.HasPrefix(l, "#") { // Comment
			commentBuffer = append(commentBuffer, &Comment{Str: line})
		} else if fieldRegex.MatchString(line) { // Field
			if field, ok := tryMakeFieldFromString(line); ok {
				// Attach comment buffer
				field.CommentsBefore = commentBuffer
				commentBuffer = nil
				fieldBuffer = append(fieldBuffer, field)
			} else {
				log.Panicf("Did not recognise line %v", line)
			}
		} else if sectionRegex.MatchString(line) { // Section
			if section, ok := tryMakeSectionFromString(line); ok {
				if len(f.Sections) > 0 {
					f.Sections[len(f.Sections)-1].Fields = fieldBuffer
				} else if len(f.Sections) == 0 && len(fieldBuffer) > 0 {
					f.Fields = fieldBuffer
					fieldBuffer = nil
				}
				fieldBuffer = nil
				section.CommentsBefore = commentBuffer
				commentBuffer = nil
				f.Sections = append(f.Sections, &section)
			}
		} else { // Unrecognised
			log.Panicf("Did not recognise line %v", line)
		}
	}

	// Deal with fields and comments still left in the buffers
	if len(fieldBuffer) > 0 && len(f.Sections) > 0 {
		f.Sections[len(f.Sections)-1].Fields = fieldBuffer
		fieldBuffer = nil
	} else if len(fieldBuffer) > 0 && len(f.Sections) == 0 {
		f.Fields = fieldBuffer
		fieldBuffer = nil
	}

	if len(commentBuffer) > 0 {
		f.CommentsAfter = commentBuffer
		commentBuffer = nil
	}

	return f
}

// Read reads a file into an ast.File struct.
// func oldRead(file io.Reader) File {
// 	var f File
// 	scanner := bufio.NewScanner(file)
//
// 	currentSection := ""
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if strings.Contains(line, "=") || line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
// 			// Append field to the current section
// 			if currentSection == "" || currentSection == "_preamble" {
// 				if len(f.Sections) == 0 {
// 					// If we're in here, then we're picking up some preamble
// 					// which could be blank lines, comments, or something else.
// 					sectionToAppend := makeSection("_preamble", "")
// 					f.Sections = append(f.Sections, &sectionToAppend)
// 					currentSection = "_preamble"
// 				}
// 			}
// 			for i, s := range f.Sections {
// 				if s.key == currentSection {
// 					fieldToAppend := makeFieldFromString(line)
// 					f.Sections[i].Fields = append(f.Sections[i].Fields, &fieldToAppend)
// 				}
// 			}
// 		} else if matched, err := regexp.MatchString(`^ *\[.*\] *$`, line); err != nil {
// 			log.Panicf("Error matching regexp: %v", err)
// 		} else if matched {
// 			// Matched a section title
// 			sectionToAppend := makeSectionFromString(line)
// 			f.Sections = append(f.Sections, &sectionToAppend)
// 			currentSection = f.Sections[len(f.Sections)-1].key
// 		}
// 	}
//
// 	return f
// }
