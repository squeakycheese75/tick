package api_test

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/squeakycheese75/tick/internal/delivery/api"
)

// The OpenAPI document is the contract callers read, so it must describe exactly the routes the
// server serves: no route missing from it, and none in it that the server does not have.
func TestOpenAPIDocumentMatchesTheRoutes(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}

	documented := documentedOperations(string(raw))

	served := api.Routes()
	sort.Strings(served)

	var inSpec []string
	for op := range documented {
		inSpec = append(inSpec, op)
	}
	sort.Strings(inSpec)

	if strings.Join(served, "\n") != strings.Join(inSpec, "\n") {
		t.Fatalf("the spec and the server disagree.\nserved:\n  %s\ndocumented:\n  %s",
			strings.Join(served, "\n  "), strings.Join(inSpec, "\n  "))
	}
}

var (
	pathLine   = regexp.MustCompile(`^  (/[^\s:]*):\s*$`)
	methodLine = regexp.MustCompile(`^    (get|post|put|patch|delete):\s*$`)
)

// documentedOperations lists "METHOD /path" for each operation under the spec's paths: section.
func documentedOperations(spec string) map[string]bool {
	ops := map[string]bool{}
	inPaths := false
	current := ""

	for _, line := range strings.Split(spec, "\n") {
		switch {
		case strings.HasPrefix(line, "paths:"):
			inPaths = true
		case inPaths && line != "" && !strings.HasPrefix(line, " "):
			inPaths = false
		case !inPaths:
		default:
			current = noteOperation(ops, current, line)
		}
	}

	return ops
}

// noteOperation reads one line of the paths section: a path starts a new current path, and a
// method line under it records an operation.
func noteOperation(ops map[string]bool, current, line string) string {
	if m := pathLine.FindStringSubmatch(line); m != nil {
		return m[1]
	}

	if m := methodLine.FindStringSubmatch(line); m != nil && current != "" {
		ops[strings.ToUpper(m[1])+" "+current] = true
	}

	return current
}
