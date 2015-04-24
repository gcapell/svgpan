package main

import (
	"bufio"
	"bytes"
	"encoding/XML"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func scanXMLToken(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, io.EOF
	}
	// if data starts with '<', return everything up to (and including) closing '>'
	if data[0] == '<' {
		endPos := bytes.Index(data, []byte{'>'})
		if endPos == -1 {
			if !atEOF {
				return 0, nil, nil
			}
			return 0, data, fmt.Errorf("incomplete tag %q", string(data))
		}
		return endPos + 1, data[:endPos+1], nil
	}
	// return everything up to (but not including) opening '<'
	endPos := bytes.Index(data, []byte{'<'})
	if endPos == -1 {
		if !atEOF {
			return 0, nil, nil
		}
		return len(data), data, nil
	}
	return endPos, data[:endPos], nil
}

// change/add id, prepend script
func addPan(t string) string {
	d := xml.NewDecoder(strings.NewReader(t))
	tok, err := d.RawToken()
	if err != nil {
		log.Fatal("addPan", err)
	}
	s := tok.(xml.StartElement)
	for pos, a := range s.Attr {
		if a.Name.Local == "id" {
			a.Value = "viewport"
			s.Attr[pos] = a
			break
		}
	}
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := enc.EncodeToken(tok); err != nil {
		log.Fatal("encodeToken", err)
	}
	enc.Flush()

	script := `<script xlink:href="http://svgpan.googlecode.com/svn/trunk/SVGPan.js"/>`
	return script + "\n" + buf.String()
}

func filterPan(src io.Reader, dst io.Writer) error {
	s := bufio.NewScanner(src)
	s.Split(scanXMLToken)
	first := true
	for s.Scan() {
		tok := s.Text()
		if first {
			if strings.HasPrefix(tok, "<g") {
				first = false
				tok = addPan(tok)
			}
		}
		if _, err := io.WriteString(dst, tok); err != nil {
			return err
		}
	}
	return s.Err()
}

func main() {
	filterPan(os.Stdin, os.Stdout)
}
