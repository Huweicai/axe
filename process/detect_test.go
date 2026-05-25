package process

import "testing"

func TestParseProcessList(t *testing.T) {
	input := `12345 claude -r abc123 --cwd /Users/test/project1
67890 codex resume xyz --cwd /Users/test/project2
11111 codex --cwd /Users/test/project3`

	result := ParseProcessList(input)
	for _, path := range []string{"/Users/test/project1", "/Users/test/project2", "/Users/test/project3"} {
		if !result[path] {
			t.Errorf("expected %q in running set", path)
		}
	}
	if result["/nonexistent"] {
		t.Error("expected /nonexistent not in running set")
	}
}

func TestParseProcessList_Empty(t *testing.T) {
	result := ParseProcessList("")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}
