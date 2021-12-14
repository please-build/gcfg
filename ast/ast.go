package ast

import "strings"

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
	Lines       []int
	Infos       []lineInfo
	NumSections int
	Sections    []Section
}

type Section struct {
	Title             string
	Fields            []Field
	LeadingWhiteSpace int
}

type Field struct {
	key   string
	value string
}

func MakeSection(s string) Section {
	// Strip brackets
	stripped := strings.Replace(s, "[", "", -1)
	stripped = strings.Replace(stripped, "]", "", -1)
	return Section{Title: stripped}
}
