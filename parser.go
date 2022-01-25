package main

import (
	"bytes"
)

// https://go.dev/ref/spec
type SourceFile struct {
	PackageClause []byte
	ImportDecl    [][]byte
	TopLevelDecl  []byte
}

func ParseSourceFile(src []byte) *SourceFile {
	section := 0

	var line []byte
	s := &SourceFile{}
	for _, ch := range src {
		line = append(line, ch)

		if ch == '\n' {
			switch section {
			case 0: // before import section
				s.PackageClause = append(s.PackageClause, line...)
				if bytes.Equal(line, []byte("import (\n")) {
					// start reading import block
					section = 1
				}
			case 1: // import section
				if bytes.Equal(line, []byte("\n")) {
					// ignore empty line
					break
				}
				if bytes.HasPrefix(line, []byte("\t//")) {
					// ignore commented line
					break
				}
				if bytes.Equal(line, []byte(")\n")) {
					s.TopLevelDecl = append(s.TopLevelDecl, line...)
					section = 2
					break
				}
				s.ImportDecl = append(s.ImportDecl, line)
			case 2: // after import section
				s.TopLevelDecl = append(s.TopLevelDecl, line...)
			}
			line = nil
		}
	}
	return s
}
