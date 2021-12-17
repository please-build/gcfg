package ast

import (
	"log"
	"strings"
)

// A File is a handle for a file belonging to a FileSet.
// A File has a name, size, and line offset table.
//
type File struct {
	// set  *FileSet
	Name string // file name as provided to AddFile
	Base int    // Pos value range for this file is [base...base+size]
	Size int    // file size as provided to AddFile

	// lines and infos are protected by set.mutex
	Lines    int
	Sections []Section
}

type Section struct {
	// This is the section header as read from the config file
	Header string
	// Canonicalised key for identifying the section
	Key               string
	Line              int
	Fields            []Field
	LeadingWhiteSpace int
}

type Field struct {
	Key   string
	Value string

	// If a comment line is found, store it here
	// and leave Key and Value empty
	Comment string
}

func MakeSection(s string, line int) Section {
	return Section{
		Header: s,
		Key:    getSectionKeyFromString(s),
		Line:   line,
	}
}

func MakeField(s string) Field {
	// A field can be a blank line
	if s == "" {
		return Field{
			Key:   "",
			Value: "",
		}
	} else if strings.HasPrefix(s, ";") || strings.HasPrefix(s, "#") {
		return Field{
			Key:     "",
			Value:   "",
			Comment: s,
		}
	}

	split := strings.Split(s, "=")
	if len(split) != 2 {
		log.Fatalf("Got invalid field \"%v\"", s)
	}
	return Field{
		Key:   split[0],
		Value: split[1],
	}
}

func getSectionKeyFromString(s string) string {
	n := strings.Count(s, "[")
	n += strings.Count(s, "]")
	if n != 2 {
		log.Fatalf("Invalid section header: %v", s)
	}
	stripped := strings.Replace(s, "[", "", -1)
	stripped = strings.Replace(stripped, "]", "", -1)
	stripped = strings.ToLower(stripped)
	return stripped
}

func (s Section) ToBytes() []byte {
	if s.Key == "" {
		log.Fatalf("Tried to convert an empty section to byte slice")
	}
	return []byte(s.Header + "\n")
}

func (f Field) ToBytes() []byte {
	if f.Comment != "" {
		return []byte(f.Comment + "\n")
	}
	if f.Key == "" && f.Value == "" {
		return []byte("\n")
	}
	if f.Key == "" || f.Value == "" {
		log.Fatalf("Key or value missing for field: %v", f)
	}

	s := f.Key + "=" + f.Value + "\n"
	return []byte(s)
}
