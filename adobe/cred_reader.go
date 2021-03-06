package adobe

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CredReader parses a credential dump into records.
type CredReader struct {
	br     *bufio.Reader
	record int
	line   int
}

// NewCredReader constructs a CredReader.
func NewCredReader(r io.Reader) *CredReader {
	return &CredReader{bufio.NewReader(r), 0, 0}
}

// Read reads one record from r.
func (r *CredReader) Read() ([]string, error) {
	var line string
	for line == "" {
		r.line++
		l, err := r.br.ReadString('\n')
		if err != nil {
			return nil, &ParseError{r.record, r.line, err}
		}
		line = l[:len(l)-1]
	}

	if !strings.HasSuffix(line, "|--") {
		// Exit if row count encountered
		if strings.HasSuffix(line, " rows selected.") {
			records, err := strconv.Atoi(line[:len(line)-len(" rows selected.")])
			if err != nil {
				return nil, &ParseError{-1, r.line, err}
			}
			if records != r.record {
				return nil, &ParseError{-1, r.line, fmt.Errorf("%d records expected, but %d parsed", records, r.record)}
			}
			return nil, io.EOF
		}

		// Join with the next line to make a complete row
		r.line++
		next, err := r.br.ReadString('\n')
		if err != nil {
			return nil, &ParseError{r.record, r.line, err}
		}
		line += next[:len(next)-1]
		if !strings.HasSuffix(line, "|--") {
			return nil, &ParseError{r.record, r.line, errors.New("unterminated row")}
		}
	}

	line = line[:len(line)-len("|--")]
	record := strings.SplitN(line, "-|-", 5)
	if len(record) < 5 {
		return nil, &ParseError{r.record, r.line, fmt.Errorf("only %d columns", len(record))}
	}
	r.record++
	return record, nil
}

// ParseError is an error returned during parsing.
type ParseError struct {
	Record int
	Line   int
	Err    error
}

func (err *ParseError) Error() string {
	if err.Record == -1 {
		return fmt.Sprintf("record total on line %d: %v", err.Line, err.Err)
	}
	return fmt.Sprintf("record %d on line %d: %v", err.Record, err.Line, err.Err)
}
