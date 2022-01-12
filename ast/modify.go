package ast

import "strings"

// InjectField injects a field into the AST and returns the modified file.
func InjectField(f File, field Field, section string, repeatable bool) File {
	// Read the file so we know where to inject
	sectionKey := GetSectionKeyFromString(section)

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
		lastSection.Fields = append(lastSection.Fields, Field{Key: "", Value: ""})
	}

	header := MakeSectionHeader(section)
	astSection := MakeSection(header)
	astSection.Fields = append(astSection.Fields, field)
	f.Sections = append(f.Sections, astSection)

	return f
}
