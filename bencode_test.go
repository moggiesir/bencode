package bencode

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	result, err := Parse(strings.NewReader("ld3:fooli1234eee3:bare"))
	if err != nil {
		t.Error(err)
	}
	expected := []interface{}{
		map[string]interface{}{
			"foo": []interface{}{1234},
		},
		[]byte{'b', 'a', 'r'},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Error("Expected %v, got ", expected, result)
	}
}
