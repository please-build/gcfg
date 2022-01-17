package ast

import (
	"strings"
)

// InjectField injects a field into the AST and returns the modified file.
// If multiple sections exist with the same name, we insert into the first
// one we find.
func InjectField(f File, key, value, section, subsection string, repeatable bool) File {
	// Work out which section we want to inject into
	sectionKey := getKeyFromSectionAndSubsection(section, subsection)

	fi := makeField(key, value)

	// If file is empty, we can insert the section and field without any further checks
	if len(f.sections) == 0 {
		s := makeSection(section, subsection)
		s.fields = append(s.fields, makeField(key, value))
		f.sections = append(f.sections, s)
		return f
	}

	// If section exists, add field to section
	for i, s := range f.sections {
		if s.key == sectionKey {
			if !repeatable {
				for j, k := range s.fields {
					if k.key == fi.key || strings.TrimSpace(k.key) == fi.key {
						f.sections[i].fields[j].value = fi.value
						f.sections[i].fields[j].str = k.key + "= " + fi.value
						return f
					}
				}
			}
			f.sections[i].fields = append(f.sections[i].fields, fi)
			return f
		}
	}

	// Couldn't find section so create new one
	// Check if file currently ends with a blank line first
	needToAppendSpace := true
	lastSection := &f.sections[len(f.sections)-1]
	if len(lastSection.fields) > 0 {
		lastField := &lastSection.fields[len(lastSection.fields)-1]
		if lastField.isBlankLine() {
			needToAppendSpace = false
		}
	}

	if needToAppendSpace {
		f.sections[len(f.sections)-1].fields = append(f.sections[len(f.sections)-1].fields, field{})
	}

	s := makeSection(section, subsection)
	s.fields = append(s.fields, makeField(key, value))
	f.sections = append(f.sections, s)
	return f
}

// DeleteSection deletes all sections found in the AST with the name sect
func DeleteSection(file File, sect, subsection string) File {
	// Work out which section we want to delete
	s := makeSection(sect, subsection)

	for i := 0; i < len(file.sections); i++ {
		if file.sections[i].key == s.key {
			if len(file.sections) > i+1 {
				file.sections = append(file.sections[:i], file.sections[i+1:]...)
				i--
			} else {
				file.sections = file.sections[:i]
			}

		}
	}
	return file
}
