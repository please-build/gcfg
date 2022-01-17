package ast

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// A File is an AST representation of a config file.
type File struct {
	name     string
	sections []section
}

type section struct {
	// str is the as-read section if the
	// section was read from a config file
	str        string
	name       string
	subsection string
	key        string
	fields     []field
}

type field struct {
	// str is the as-read field if the
	// field was read from a config file
	str   string
	key   string
	value string

	// If a comment line is found, store it here
	// and leave key and value empty
	comment string
}

// makeSection initialises a section in the ast given a section
// and a subsection
func makeSection(sect, subsection string) section {
	s := section{
		name:       sect,
		subsection: subsection,
		key:        makeSectionKey(sect, subsection)}

	if subsection == "" {
		s.str = "[" + sect + "]"
	} else {
		s.str = "[" + sect + " \"" + subsection + "\"]"
	}

	return s
}

// makeField initialises an ast field given a key and a value
func makeField(key, value string) field {
	return field{key: key, value: value}
}

// getStr returns the string associated with a field
func (f field) getStr() string {
	if f.str != "" {
		return f.str
	}

	return f.key + " = " + f.value
}

// getStr returns the string associated with a section
func (s section) getStr() string {
	if s.str != "" {
		return s.str
	}
	if s.subsection != "" {
		return "[" + s.name + " " + "\"" + s.subsection + "\"]"
	}
	return "[" + s.name + "]"
}

// makeSectionKey creates a key for identifying a section
func makeSectionKey(sect, subsection string) string {
	if subsection == "" {
		return strings.ToLower(sect)
	}
	return strings.ToLower(sect) + "&" + strings.ToLower(subsection)
}

// makeSectionFromString takes a string read from a config file and returns an ast section
func makeSectionFromString(str string) section {
	s := section{str: str}
	// Check first if str has a
	secAndSubReg := regexp.MustCompile(`^ *\[ *[a-zA-Z0-9_.-]+ *" *[a-zA-Z0-9_.-]+ *" *\] *$`)
	secOnlyReg := regexp.MustCompile(`^ *\[ *[a-zA-Z0-9_.-]+ *\] *$`)

	if secAndSubReg.MatchString(str) {
		subsectionReg := regexp.MustCompile(`(?:" *)([a-zA-Z0-9_.-]+)(?: *")`)
		s.subsection = subsectionReg.FindStringSubmatch(str)[1]
		sectionReg := regexp.MustCompile(`(?:\[ *)([a-zA-Z0-9_.-]+)(?: *")`)
		s.name = sectionReg.FindStringSubmatch(str)[1]
		s.key = makeSectionKey(s.name, s.subsection)
	} else if secOnlyReg.MatchString(str) {
		sectionReg := regexp.MustCompile(`(?:\[ *)([a-zA-Z0-9_.-]+)(?: *\])`)
		s.name = sectionReg.FindStringSubmatch(str)[1]
		s.key = makeSectionKey(s.name, "")
	}

	return s
}

// makeFieldFromString takes a string read from a config file and returns an ast field
func makeFieldFromString(s string) field {
	if s == "" {
		// A field can be a blank line
		return field{}
	} else if strings.HasPrefix(s, ";") || strings.HasPrefix(s, "#") {
		// or a comment
		return field{comment: s}
	}

	ret := field{str: s}
	keyValueReg := regexp.MustCompile(`(^ *[a-zA-Z0-9_ .-]+ *)(=)( *[a-zA-Z0-9_/ .-]+ *$)`)
	if !(keyValueReg.MatchString(s)) {
		log.Fatalf("Could not parse field %v", s)
	}

	ret.key = keyValueReg.FindStringSubmatch(s)[1]
	ret.value = keyValueReg.FindStringSubmatch(s)[3]

	return ret
}

// GetSectionKeyFromString strips brackets from a section header and lowers it.
// E.g. '[Something "sub"]' -> 'something "sub"'
func getKeyFromSectionAndSubsection(sect, subsection string) string {
	if subsection == "" {
		return strings.ToLower(sect)
	}
	return strings.ToLower(sect) + "&" + strings.ToLower(subsection)
}

// ToBytes returns a correctly formatted section header as a byte slice.
// Needed for writing to output file.
func (s section) toBytes() []byte {
	if s.key == "" {
		log.Fatalf("Tried to convert an empty section to byte slice")
	}
	str := s.getStr()
	return []byte(str + "\n")
}

// ToBytes returns a field as a byte slice. Needed for writing
// to output file.
func (f field) toBytes() []byte {
	if f.comment != "" {
		return []byte(f.comment + "\n")
	}
	if f.key == "" && f.value == "" {
		return []byte("\n")
	}
	if f.key == "" || f.value == "" {
		log.Fatalf("Key or value missing for field: %v", f)
	}

	s := f.getStr() + "\n"
	return []byte(s)
}

// isBlankLine returns whether f is a field representing a blank line
func (f field) isBlankLine() bool {
	return f.comment == "" &&
		f.key == "" &&
		f.value == ""
}

// PrintDebug prints an entire AST File to help with debugging.
func (f File) PrintDebug() {
	log.Printf("----------------")
	log.Printf("Name: %v", f.name)
	log.Printf("Lines: %v", f.numLines())
	log.Printf("Contents:")
	for _, s := range f.sections {
		log.Printf("%v", s.str)
		for _, field := range s.fields {
			log.Printf("%v", field.str)
		}
	}
	log.Printf("----------------")
}

// MakeSectionHeader tries to form a textual section header
// from a string which may or may not have brackets already.
func makeSectionHeader(s string) string {
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

// numLines returns the number of lines in the section including the section title
func (s section) numLines() int {
	if s.key == "_preamble" {
		return len(s.fields)
	}

	return len(s.fields) + 1
}

// numLines sums all section line counts and returns the number of lines in the file
func (f File) numLines() int {
	ret := 0
	for _, s := range f.sections {
		ret += s.numLines()
	}
	return ret
}
