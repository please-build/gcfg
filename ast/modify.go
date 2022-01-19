package ast

import (
	"strings"
)

// InjectField injects a field into the AST and returns the modified file.
// If multiple sections exist with the same name, we insert into the first
// one we find.
func InjectField(f File, fieldName, fieldValue, sectionName, subsectionName string, repeatable bool) File {
	// Work out which section we want to inject into
	sectionKey := getKeyFromSectionAndSubsection(sectionName, subsectionName)

	fi := makeField(fieldName, fieldValue)

	// If file is empty, we can insert the section and field without any further checks
	if len(f.Sections) == 0 {
		s := makeSection(sectionName, subsectionName)
		s.Fields = append(s.Fields, &fi)
		f.Sections = append(f.Sections, &s)
		return f
	}

	// If section exists, add field to section
	for i, s := range f.Sections {
		if s.Key == sectionKey {
			if !repeatable {
				for j, k := range s.Fields {
					if k.Name == fi.Name || strings.TrimSpace(k.Name) == fieldName {
						f.Sections[i].Fields[j].Value = fi.Value
						f.Sections[i].Fields[j].Str = k.Name + "= " + fi.Value
						return f
					}
				}
			}
			f.Sections[i].Fields = append(f.Sections[i].Fields, &fi)
			return f
		}
	}

	s := makeSection(sectionName, subsectionName)

	// Append blank line if file does not end with blank line
	if len(f.CommentsAfter) == 0 {
		s.CommentsBefore = append(s.CommentsBefore, &Comment{})
	}

	s.Fields = append(s.Fields, &fi)
	f.Sections = append(f.Sections, &s)
	return f
}

// DeleteSection deletes all sections found in the AST with the name sect
func DeleteSection(file File, sect, subsection string) File {
	// Work out which section we want to delete
	s := makeSection(sect, subsection)

	for i := 0; i < len(file.Sections); i++ {
		if file.Sections[i].Key == s.Key {
			if len(file.Sections) > i+1 {
				file.Sections = append(file.Sections[:i], file.Sections[i+1:]...)
				i--
			} else {
				file.Sections = file.Sections[:i]
			}

		}
	}
	return file
}
