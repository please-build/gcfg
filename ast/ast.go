package ast

import (
	"fmt"
	"log"
	"strings"
)

// A File is an AST representation of a config file.
type File struct {
	Name     string
	NumLines int
	Sections []Section
}

type Section struct {
	// Section header as read from the config file
	Header string
	// Canonicalised key for identifying the section
	Key    string
	Fields []Field
}

type Field struct {
	Key   string
	Value string

	// If a comment line is found, store it here
	// and leave Key and Value empty
	Comment string
}

func MakeSection(header string) Section {
	return Section{
		Header: header,
		Key:    GetSectionKeyFromString(header),
	}
}

func MakeField(s string) Field {
	if s == "" {
		// A field can be a blank line
		return Field{}
	} else if strings.HasPrefix(s, ";") || strings.HasPrefix(s, "#") {
		// or a comment
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

// GetSectionKeyFromString strips brackets from a section header and lowers it.
// E.g. '[Something "sub"]' -> 'something "sub"'
func GetSectionKeyFromString(s string) string {
	n := strings.Count(s, "[")
	n += strings.Count(s, "]")
	if n == 0 {
		return strings.ToLower(s)
	}
	if n != 2 {
		log.Fatalf("Invalid section header: %v", s)
	}
	stripped := strings.Replace(s, "[", "", -1)
	stripped = strings.Replace(stripped, "]", "", -1)
	stripped = strings.ToLower(stripped)
	return stripped
}

// ToBytes returns a correctly formatted section header as a byte slice.
// Needed for writing to output file.
func (s Section) ToBytes() []byte {
	if s.Key == "" {
		log.Fatalf("Tried to convert an empty section to byte slice")
	}
	return []byte(s.Header + "\n")
}

// ToBytes returns a field as a byte slice. Needed for writing
// to output file.
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

func (f Field) IsBlankLine() bool {
	return f.Comment == "" &&
		f.Key == "" &&
		f.Value == ""
}

// PrintDebug prints an entire AST File to help with debugging.
func (f File) PrintDebug() {
	log.Printf("Name: %v", f.Name)
	log.Printf("Lines: %v", f.NumLines)
	log.Printf("Contents:")
	for _, s := range f.Sections {
		log.Printf("%v", s.Header)
		for _, field := range s.Fields {
			if field.Key == "" && field.Value == "" && field.Comment == "" {
				log.Printf("")
			} else if field.Comment == "" {
				log.Printf("%v=%v", field.Key, field.Value)
			} else {
				log.Printf("%v", field.Comment)
			}
		}
	}
}

// MakeSectionHeader tries to form a textual section header
// from a string which may or may not have brackets already.
func MakeSectionHeader(s string) string {
	n := strings.Count(s, "[")
	n += strings.Count(s, "]")

	if n > 2 {
		panic(fmt.Sprintf("Badly-formed section header %v passed to MakeSectionHeader", s))
	}
	if n == 2 {
		if strings.HasPrefix(s, "[") && strings.HasPrefix(s, "]") {
			// This is a fully-formed section header
			return s
		}
		panic(fmt.Sprintf("Badly-formed section header %v passed to MakeSectionHeader", s))
	}
	if n == 1 {
		panic(fmt.Sprintf("Badly-formed section header %v passed to MakeSectionHeader", s))
	}

	return "[" + s + "]"
}
