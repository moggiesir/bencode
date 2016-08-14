package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type ParseError struct {
	Index int
	Error error
}

func (e *ParseError) String() string {
	return fmt.Sprintf("bencode: parse error at %d: %v", e.Index, e.Error)
}

type scanner struct {
	internal *bufio.Reader
	position int
	results  []interface{}
}

func (s *scanner) readByte() byte {
	b, err := s.internal.ReadByte()
	if err != nil {
		panic(&ParseError{s.position, err})
	}
	s.position++
	return b
}

func (s *scanner) peek(n int) []byte {
	bytes, err := s.internal.Peek(n)
	if err != nil {
		panic(&ParseError{s.position, err})
	}
	return bytes
}

func (s *scanner) readBytes(delim byte) []byte {
	bytes, err := s.internal.ReadBytes(delim)
	if err != nil {
		panic(&ParseError{s.position, err})
	}
	s.position += len(bytes)
	return bytes
}

func (s *scanner) read(bytes []byte) int {
	count, err := s.internal.Read(bytes)
	if err != nil {
		panic(&ParseError{s.position, err})
	}
	s.position += count
	return count
}

func (s *scanner) int() int {
	s.readByte()
	numbers := s.readBytes('e')
	result, err := strconv.Atoi(string(numbers[:len(numbers)-1]))
	if err != nil {
		panic(&ParseError{s.position, err})
	}
	s.results = append(s.results, result)
	return result
}

func (s *scanner) string() []byte {
	position := s.position
	numbers := s.readBytes(':')
	length, err := strconv.Atoi(string(numbers[:len(numbers)-1]))
	if err != nil {
		panic(&ParseError{s.position, err})
	}
	if length < 1 {
		panic(&ParseError{position, fmt.Errorf("bad string length %d", length)})
	}
	result := make([]byte, length)
	total := 0
	for total < length {
		total += s.read(result);
		if total < length {
			result = result[total:]
		}
	}
	s.results = append(s.results, result)
	return result
}

func (s *scanner) list() []interface{} {
	result := make([]interface{}, 0)
	s.results = append(s.results, result)
	s.readByte()
	c := s.peek(1)[0]
	for c != 'e' {
		result = append(result, s.value())
		c = s.peek(1)[0]
	}
	s.readByte()
	return result
}

func (s *scanner) dict() map[string]interface{} {
	result := make(map[string]interface{})
	s.results = append(s.results, result)
	s.readByte()
	c := s.peek(1)[0]
	for c != 'e' {
		key := string(s.string())
		value := s.value()
		result[key] = value
		c = s.peek(1)[0]
	}
	s.readByte()
	return result
}

func (s *scanner) value() interface{} {
	position := s.position
	c := s.peek(1)[0]
	switch {
	case c == 'i':
		return s.int()
	case c > '0' && c < '9':
		return s.string()
	case c == 'l':
		return s.list()
	case c == 'd':
		return s.dict()
	}
	panic(&ParseError{position, fmt.Errorf("bad start: %c", c)})
}

func Parse(r io.Reader) (result interface{}, err error) {
	s := &scanner{bufio.NewReader(r), 0, make([]interface{}, 0)}
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v - results so far: %v", r, s.results)
			}
		}
	}()

	result = s.value()
	if result == nil {
		err = fmt.Errorf("bencode: bad starting character")
	}
	return
}

