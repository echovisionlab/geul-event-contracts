package contracttest_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readSourceProto(t *testing.T, relativePath string) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source contract test path")
	}
	contents, err := os.ReadFile(filepath.Join(filepath.Dir(sourceFile), "..", "proto", filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatalf("read source proto %s: %v", relativePath, err)
	}
	return string(contents)
}

func sourceDeclaration(t *testing.T, source, kind, name string) string {
	t.Helper()
	prefix := kind + " " + name + " {"
	start := strings.Index(source, prefix)
	if start < 0 {
		t.Fatalf("source declaration %s %s is missing", kind, name)
	}
	depth := 0
	for index := start + len(prefix) - 1; index < len(source); index++ {
		switch source[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[start : index+1]
			}
		}
	}
	t.Fatalf("source declaration %s %s is not closed", kind, name)
	return ""
}

func requireSourceContains(t *testing.T, source string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Errorf("source contract is missing %q", fragment)
		}
	}
}

func requireSourceExcludes(t *testing.T, source string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(source, fragment) {
			t.Errorf("source contract must not contain %q", fragment)
		}
	}
}
