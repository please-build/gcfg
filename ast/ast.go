package ast

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
	name string // file name as provided to AddFile
	base int    // Pos value range for this file is [base...base+size]
	size int    // file size as provided to AddFile

	// lines and infos are protected by set.mutex
	lines       []int
	infos       []lineInfo
	numSections int
	sections    []Section
}

type Section struct {
	title             string
	fields            []Field
	leadingWhiteSpace int
}

type Field struct {
	key   string
	value string
}
