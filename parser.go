package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// https://go.dev/ref/spec
type SourceFile struct {
	PackageClause []byte
	ImportDecl    [][]byte
	TopLevelDecl  []byte
}

func ParseSourceFile(r io.Reader) (*SourceFile, error) {
	section := 0

	s := &SourceFile{}

	rd := bufio.NewReader(r)
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return s, nil
			}
			return nil, fmt.Errorf("read: %s", err)
		}

		switch section {
		case 0: // before import section
			s.PackageClause = append(s.PackageClause, line...)
			if bytes.Equal(line, []byte("import (\n")) {
				// start import block
				section = 1
			}
		case 1: // import section
			if bytes.Equal(line, []byte(")\n")) {
				// end import block
				section = 2

				s.TopLevelDecl = append(s.TopLevelDecl, line...)
				break
			}
			s.ImportDecl = append(s.ImportDecl, line)
		case 2: // after import section
			s.TopLevelDecl = append(s.TopLevelDecl, line...)
		}
	}
}
