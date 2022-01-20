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
	fieldRegex := regexp.MustCompile(`^ *[a-zA-Z0-9_.-]+ *= *[a-zA-Z0-9_ /.-]`)
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
				log.Panicf("could not parse field '%v'. Check the field name does not contain spaces. If the field value contains spaces, make sure to enclose the string in double quotes.", line)
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
			log.Panicf("did not recognise line %v", line)
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

// tryMakeFieldFromString takes a field string and tries to return an AST field
func tryMakeFieldFromString(s string) (*Field, bool) {
	reg := regexp.MustCompile(`^ *([a-zA-Z0-9_.-]+) *= *([a-zA-Z0-9_@&,|"~<>/:= +.-]+)?( *;.*)?`)
	matches := reg.FindStringSubmatch(s)
	if len(matches) == 0 {
		return nil, false
	}

	return &Field{
		Str:             s,
		Name:            matches[1],
		Value:           matches[2],
		TrailingComment: matches[3]}, true
}

// tryMakeSectionFromString takes a string tries to make a Section
func tryMakeSectionFromString(line string) (Section, bool) {
	s := Section{HeadingStr: line}
	secAndSubReg := regexp.MustCompile(`^ *\[ *[a-zA-Z0-9_.-]+ *" *[a-zA-Z0-9_.-]+ *" *\].*`)
	secOnlyReg := regexp.MustCompile(`^ *\[ *[a-zA-Z0-9_.-]+ *\].*`)

	if secAndSubReg.MatchString(line) {
		subsectionReg := regexp.MustCompile(`(?:" *)([a-zA-Z0-9_.-]+)(?: *")`)
		s.Subsection = strings.ToLower(subsectionReg.FindStringSubmatch(line)[1])
		sectionReg := regexp.MustCompile(`(?:\[ *)([a-zA-Z0-9_.-]+)(?: *")`)
		s.Name = strings.ToLower(sectionReg.FindStringSubmatch(line)[1])
		s.Key = makeSectionKey(s.Name, s.Subsection)
		return s, true
	}

	if secOnlyReg.MatchString(line) {
		sectionReg := regexp.MustCompile(`(?:\[ *)([a-zA-Z0-9_.-]+)(?: *\])`)
		s.Name = strings.ToLower(sectionReg.FindStringSubmatch(line)[1])
		s.Key = makeSectionKey(s.Name, "")
		return s, true
	}

	return s, false
}
