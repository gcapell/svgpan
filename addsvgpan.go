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
	"flag"
)

var (
	script = flag.String("script", "https://raw.githubusercontent.com/gcapell/svgpan/master/SVGPan.js", "URL/file for SVGPan script")
	outFile = flag.String("o", "", "filename to output to, defaults to stdout")
	inFile = flag.String("i", "", "filename to read from, defaults to stdin")
)

func scanWhatever(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, io.EOF
	}
	return len(data), data, nil
}

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
func addPan(t string) (string, error) {
	d := xml.NewDecoder(strings.NewReader(t))
	tok, err := d.RawToken()
	if err != nil {
		return "", fmt.Errorf("%s decoding %q", err, tok)
	}
	s, ok := tok.(xml.StartElement)
	if !ok  {
		return "", fmt.Errorf("expected %#v to be StartElement", tok)
	}
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
		return "", fmt.Errorf("%s encoding %#v", tok)
	}
	enc.Flush()

	script := fmt.Sprintf(`<script xlink:href="%s"/>`, *script)
	return script + "\n" + buf.String(), nil
}

func filterPan(src io.Reader, dst io.Writer) error {
	s := bufio.NewScanner(src)
	s.Split(scanXMLToken)
	seeking := true	// ... for first g element
	for s.Scan() {
		tok := s.Text()
		if seeking && strings.HasPrefix(tok, "<g") {
			seeking = false
			var err error
			tok, err = addPan(tok)
			if err != nil {
				return err
			}
			s.Split(scanWhatever)
		}
		if _, err := io.WriteString(dst, tok); err != nil {
			return err
		}
	}
	return s.Err()
}

func main() {
	flag.Parse()
	inf := os.Stdin
	var err error
	if *inFile != "" {
		inf, err = os.Open(*inFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	outf := os.Stdout
	if *outFile != "" {
		outf, err = os.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := filterPan(inf, outf); err != nil {
		log.Fatal(err)
	}
}
