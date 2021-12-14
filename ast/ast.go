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
	Sections    []Section
}

type Section struct {
	Title             string
	Line              int
	Fields            []Field
	LeadingWhiteSpace int
}

type Field struct {
	key   string
	value string
}

func MakeSection(s string, line int) Section {
	// Strip brackets
	stripped := strings.Replace(s, "[", "", -1)
	stripped = strings.Replace(stripped, "]", "", -1)

	// Canonicalise title
	stripped = strings.ToLower(stripped)

	return Section{
		Title: stripped,
		Line:  line,
	}
}

func MakeField(s string) Field {
	split := strings.Split(s, "=")
	if len(split) != 2 {
		log.Fatalf("Got invalid field \"%v\"", s)
	}
	f := Field{
		key:   split[0],
		value: split[1],
	}
	return f
}
