package ast

import (
	"strings"
)

// InjectField injects a field into the AST and returns the modified file.
// If multiple sections exist with the same name, we insert into the first
// one we find.
func InjectField(f File, field Field, section string, repeatable bool) File {
	// Work out which section we want to inject into
	sectionKey := GetSectionKeyFromString(section)

	// Format field correctly for injection
	if !(strings.HasSuffix(field.Key, " ")) {
		field.Key += " "
	}
	if !(strings.HasPrefix(field.Value, " ")) {
		field.Value = " " + field.Value
	}

	// If file is empty, we can insert the section and field without any further checks
	if len(f.Sections) == 0 {
		return appendFieldToNewSection(f, field, section)
	}

	// If section exists, add field to section
	for i, s := range f.Sections {
		if s.Key == sectionKey {
			if !repeatable {
				for j, k := range s.Fields {
					if k.Key == field.Key {
						f.Sections[i].Fields[j].Value = field.Value
						return f
					}
				}
			}
			f.Sections[i].Fields = append(f.Sections[i].Fields, field)
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
		f.Sections[len(f.Sections)-1].Fields = append(f.Sections[len(f.Sections)-1].Fields, Field{})
	}

	return appendFieldToNewSection(f, field, section)
}

// appendFieldToNewSection appends a field to a new section and adds that section
// to the AST
func appendFieldToNewSection(f File, field Field, section string) File {
	header := makeSectionHeader(section)
	astSection := makeSection(header)
	astSection.Fields = append(astSection.Fields, field)
	f.Sections = append(f.Sections, astSection)

	return f
}

// DeleteSection deletes all sections found in the AST with the name sect
func DeleteSection(file File, sect string) File {
	// Work out which section we want to delete
	sectionKey := GetSectionKeyFromString(sect)

	for i := 0; i < len(file.Sections); i++ {
		if file.Sections[i].Key == sectionKey {
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
