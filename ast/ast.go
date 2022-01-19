package ast

import (
	"log"
	"strings"
)

// Comment is any whitespace or comment that is ignored by this parser
type Comment struct {
	Str string
}

// File is an AST representation of a config file
type File struct {
	Sections      []*Section
	CommentsAfter []*Comment // Any comments or whitespace that come at the end of a file
	Fields        []*Field   // Fields that don't belong to a section
}

type Section struct {
	HeadingStr     string // The literal value of this section heading
	Name           string
	Subsection     string
	Key            string // Section identifier
	Fields         []*Field
	CommentsBefore []*Comment // Any comments or whitespace between this field and whatever came before
}

type Field struct {
	Str             string // The literal value of this line, including comments and whitespace e.g. Foo = Bar ; test
	Name            string
	Value           string
	TrailingComment string     // Any comments or whitespace after the field value
	CommentsBefore  []*Comment // Any comments or whitespace between this field and whatever came before
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

// getStr returns the string associated with a field
func (f Field) getStr() string {
	if f.Str != "" {
		return f.Str
	}

	return f.Name + " = " + f.Value
}

// getHeadingStr returns the string associated with a section
func (s Section) getHeadingStr() string {
	if s.HeadingStr != "" {
		return s.HeadingStr
	}
	if s.Subsection != "" {
		return "[" + s.Name + " " + "\"" + s.Subsection + "\"]"
	}
	return "[" + s.Name + "]"
}

// makeSectionKey creates a key for identifying a section
func makeSectionKey(name, subsection string) string {
	if subsection == "" {
		return strings.ToLower(name)
	}
	return strings.ToLower(name) + "&" + strings.ToLower(subsection)
}

// getSectionKeyFromString strips brackets from a section header and lowers it.
// E.g. '[Something "sub"]' -> 'something "sub"'
func getKeyFromSectionAndSubsection(sect, subsection string) string {
	if subsection == "" {
		return strings.ToLower(sect)
	}
	return strings.ToLower(sect) + "&" + strings.ToLower(subsection)
}

// toBytes returns a correctly formatted section header as a byte slice.
// Needed for writing to output file.
func (s Section) toBytes() []byte {
	if s.Name == "" {
		log.Fatalf("Tried to convert an empty section to byte slice")
	}
	var ret string
	for _, c := range s.CommentsBefore {
		ret += c.Str + "\n"
	}
	ret += s.getHeadingStr() + "\n"
	return []byte(ret)
}

// toBytes returns a field as a byte slice. Needed for writing
// to output file.
func (f Field) toBytes() []byte {
	var ret string
	for _, c := range f.CommentsBefore {
		ret += c.Str + "\n"
	}
	ret += f.getStr() + "\n"
	return []byte(ret)
}

func (c Comment) toBytes() []byte {
	return []byte(c.Str + "\n")
}

// printDebug prints an entire AST File to help with debugging.
func (f File) printDebug() {
	log.Printf("----------------")
	log.Printf("Lines: %v", f.numLines())
	log.Printf("Contents:")
	for _, field := range f.Fields {
		log.Printf("%v", field.getStr())
	}
	for _, s := range f.Sections {
		for _, c := range s.CommentsBefore {
			log.Printf("%v", c.Str)
		}
		log.Printf("%v", s.HeadingStr)
		for _, field := range s.Fields {
			for _, c := range field.CommentsBefore {
				log.Printf("%v", c.Str)
			}
			log.Printf("%v", field.Str)
		}
	}
	for _, c := range f.CommentsAfter {
		log.Printf("%v", c.Str)
	}
	log.Printf("----------------")
}

// numLines returns the number of lines in the section including the section title
func (s Section) numFields() int {
	return len(s.Fields)
}

// numLines sums all section line counts and returns the number of lines in the file
func (f File) numLines() int {
	ret := 0
	for _, s := range f.Sections {
		ret += 1 + len(s.CommentsBefore)
		for _, field := range s.Fields {
			ret += 1 + len(field.CommentsBefore)
		}
	}
	ret += len(f.CommentsAfter) + len(f.Fields)

	return ret
}
