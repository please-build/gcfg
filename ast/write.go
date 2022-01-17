package ast

import (
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
	for _, section := range f.sections {
		if section.key != "_preamble" {
			data = append(data, section.toBytes()...)
		}
		for _, field := range section.fields {
			data = append(data, field.toBytes()...)
		}
	}

	return data
}
