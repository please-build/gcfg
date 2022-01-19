package ast

import (
	"log"
	"os"
)

// Write writes an AST file to a file on disk.
func Write(f File, output string) error {
	data := convertASTToBytes(f)
	if err := os.WriteFile(output, []byte(data), 0644); err != nil {
		return err
	}
	return nil
}

// convertASTToBytes converts an AST file to a byte slice.
func convertASTToBytes(f File) []byte {
	var data []byte
	for _, field := range f.Fields {
		data = append(data, field.toBytes()...)
	}
	for _, section := range f.Sections {
		data = append(data, section.toBytes()...)
		for _, field := range section.Fields {
			data = append(data, field.toBytes()...)
		}
	}
	for _, comment := range f.CommentsAfter {
		data = append(data, comment.toBytes()...)
	}

	return data
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
