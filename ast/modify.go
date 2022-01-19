package ast

import (
	"strings"
)

// InjectField injects a field into the AST and returns the modified file. If multiple sections
// exist with the same name, the field is inserted into the first one found.
func InjectField(f File, fieldName, fieldValue, sectionName, subsectionName string, repeatable bool) File {
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
					if k.Name == fi.Name {
						f.Sections[i].Fields[j].Value = fi.Value
						f.Sections[i].Fields[j].Str = k.Name + " = " + fi.Value + k.TrailingComment
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

// DeleteAllFieldsWithName deletes all fields with the name fieldName in section [sectionName "subsectionName"]
func DeleteAllFieldsWithName(f File, fieldName, sectionName, subsectionName string) File {
	sectionKey := getKeyFromSectionAndSubsection(sectionName, subsectionName)

	for i, s := range f.Sections {
		if s.Key == sectionKey {
			for j := 0; j < len(s.Fields); j++ {
				if s.Fields[j].Name == fieldName {
					if j == len(s.Fields)-1 {
						f.Sections[i].Fields = f.Sections[i].Fields[:j]
					} else {
						f.Sections[i].Fields = append(f.Sections[i].Fields[:j], f.Sections[i].Fields[j+1:]...)
						j--
					}
				}
			}
		}
	}
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

// MergeAllDuplicateSections merges all sections with duplicate names, along with any fields and
// comments belonging to them
func MergeAllDuplicateSections(file File) File {
	type setInt struct {
		value int
		set   bool
	}
	m := make(map[string]setInt)
	for i := 0; i < len(file.Sections); i++ {
		if m[file.Sections[i].Key].set == true {
			idx := m[file.Sections[i].Key].value
			file.Sections[idx].Fields = append(file.Sections[idx].Fields, file.Sections[i].Fields...)
			file.Sections[idx].CommentsBefore = append(file.Sections[idx].CommentsBefore, file.Sections[i].CommentsBefore...)
			if i == len(file.Sections)-1 {
				file.Sections = file.Sections[:i]
			} else {
				file.Sections = append(file.Sections[:i], file.Sections[i+1:]...)
				i--
			}
		} else {
			m[file.Sections[i].Key] = setInt{set: true, value: i}
		}
	}

	return file
}

// makeSection initialises a section in the ast given a section
// and a subsection
func makeSection(sect, subsection string) Section {
	s := Section{
		Name:       sect,
		Subsection: subsection,
		Key:        makeSectionKey(sect, subsection)}

	if subsection == "" {
		s.HeadingStr = "[" + sect + "]"
	} else {
		s.HeadingStr = "[" + sect + " \"" + subsection + "\"]"
	}

	return s
}

// makeField initialises an ast field given a key and a value
func makeField(key, value string) Field {
	return Field{Name: key, Value: value}
}

// getSectionKeyFromString strips brackets from a section header and lowers it.
// E.g. '[Something "sub"]' -> 'something&sub'
func getKeyFromSectionAndSubsection(sect, subsection string) string {
	if subsection == "" {
		return strings.ToLower(sect)
	}
	return strings.ToLower(sect) + "&" + strings.ToLower(subsection)
}
