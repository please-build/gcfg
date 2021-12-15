package ast

import (
	"log"
	"strings"
)

// A lineInfo object describes alternative file and line number
// information (such as provided via a //line comment in a .go
// file) for a given file offset.
type lineInfo struct {
	// fields are exported to make them accessible to gob
	Offset   int
	Filename string
	Line     int
}

// A File is a handle for a file belonging to a FileSet.
// A File has a name, size, and line offset table.
//
type File struct {
	// set  *FileSet
	Name string // file name as provided to AddFile
	Base int    // Pos value range for this file is [base...base+size]
	Size int    // file size as provided to AddFile

	// lines and infos are protected by set.mutex
	Lines       int
	Infos       []lineInfo
	NumSections int
	// Sections    map[string]Section
	Sections []Section
}

type Section struct {
	Title             string
	Key               string
	Line              int
	Fields            []Field
	LeadingWhiteSpace int
}

type Field struct {
	Key   string
	Value string
}

func MakeSection(s string, line int) Section {
	return Section{
		Title: getSectionTitleFromString(s),
		Line:  line,
	}
}

func MakeField(s string) Field {
	split := strings.Split(s, "=")
	if len(split) != 2 {
		log.Fatalf("Got invalid field \"%v\"", s)
	}
	f := Field{
		Key:   split[0],
		Value: split[1],
	}
	return f
}

func getSectionTitleFromString(s string) string {
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

	return []byte(s.Title)
}

func (f Field) ToBytes() []byte {
	if f.Key == "" || f.Value == "" {
		log.Fatalf("Key or value missing for field: %v", f)
	}

	s := f.Key + "=" + f.Value
	return []byte(s)
}
