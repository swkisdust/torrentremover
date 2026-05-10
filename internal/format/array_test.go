package format

import (
	"testing"

	"github.com/goccy/go-yaml"
)

func TestArrayUnmarshalYAML_Null(t *testing.T) {
	var arr Array[string]
	if err := yaml.Unmarshal([]byte("null\n"), &arr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 || arr[0] != "" {
		t.Fatalf("expected [\"\"], got %#v", arr)
	}
}

func TestArrayUnmarshalYAML_SingleElement(t *testing.T) {
	var arr Array[string]
	if err := yaml.Unmarshal([]byte("foo\n"), &arr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 || arr[0] != "foo" {
		t.Fatalf("expected [foo], got %#v", arr)
	}
}

func TestArrayUnmarshalYAML_Sequence(t *testing.T) {
	var arr Array[string]
	if err := yaml.Unmarshal([]byte("- foo\n- bar\n"), &arr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 2 || arr[0] != "foo" || arr[1] != "bar" {
		t.Fatalf("expected [foo bar], got %#v", arr)
	}
}
