package docs_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// findDocFiles finds all documentation markdown files
func findDocFiles() ([]string, error) {
	pattern := "content/docs/**/_index.md"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Also check single-level directories
	singleLevel, err := filepath.Glob("content/docs/*/_index.md")
	if err != nil {
		return nil, err
	}

	// Combine and deduplicate
	seen := make(map[string]bool)
	var result []string
	for _, m := range append(matches, singleLevel...) {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}

	return result, nil
}

// testNameFromPath converts a file path to a test name
// e.g., "content/docs/core-matchers/_index.md" -> "CoreMatchers"
func testNameFromPath(path string) string {
	// Remove content/docs/ prefix and _index.md suffix
	name := strings.TrimPrefix(path, "content/docs/")
	name = strings.TrimSuffix(name, "/_index.md")

	// Convert kebab-case to PascalCase
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// extractGoCodeBlocks extracts all ```go code blocks from a markdown file
func extractGoCodeBlocks(content string) []string {
	// Match ```go ... ``` blocks
	re := regexp.MustCompile("(?s)```go\n(.*?)```")
	matches := re.FindAllStringSubmatch(content, -1)

	var blocks []string
	for _, match := range matches {
		if len(match) > 1 {
			blocks = append(blocks, match[1])
		}
	}
	return blocks
}

// wrapCodeBlock wraps a code snippet in a minimal test file
func wrapCodeBlock(code string, index int) string {
	// Strip any existing package declaration - we'll use doctest
	packageRe := regexp.MustCompile(`(?m)^package\s+\w+\s*$`)
	code = packageRe.ReplaceAllString(code, "")

	// Check what the code contains
	hasImport := strings.Contains(code, "import")
	hasFunc := strings.Contains(code, "func Test") || strings.Contains(code, "func Example")

	var sb strings.Builder

	// Always use doctest package to match fixtures
	sb.WriteString("package doctest\n\n")

	// Strip existing import if present (we'll add our own)
	if hasImport {
		importRe := regexp.MustCompile(`(?s)import\s*\([^)]+\)`)
		code = importRe.ReplaceAllString(code, "")
		importRe2 := regexp.MustCompile(`(?m)^import\s+"[^"]+"\s*$`)
		code = importRe2.ReplaceAllString(code, "")
	}

	// Add standard imports
	sb.WriteString(`import (
	"testing"
	"github.com/james-w/specta"
)

`)

	// If it's not already a function, wrap it in a test
	if !hasFunc {
		sb.WriteString("func TestDocExample_")
		sb.WriteString(fmt.Sprintf("%d", index))
		sb.WriteString("(t *testing.T) {\n")

		// Indent the code
		lines := strings.Split(strings.TrimSpace(code), "\n")
		for _, line := range lines {
			if line != "" {
				sb.WriteString("\t")
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}

		sb.WriteString("}\n")
	} else {
		sb.WriteString(strings.TrimSpace(code))
		sb.WriteString("\n")
	}

	return sb.String()
}

// getFixturesCode returns common fixtures needed for doc examples
func getFixturesCode() string {
	return `package doctest

import (
	"time"
	"github.com/james-w/specta"
)

// Common types used in examples
type User struct {
	ID        string
	Name      string
	Email     string
	Age       int
	Active    bool
	CreatedAt time.Time
}

// Mock matcher builder for User
type UserMatcher struct {
	nameMatcher  specta.Matcher[string]
	emailMatcher specta.Matcher[string]
}

func MatchUser() *UserMatcher {
	return &UserMatcher{}
}

func (m *UserMatcher) WithName(matcher specta.Matcher[string]) *UserMatcher {
	m.nameMatcher = matcher
	return m
}

func (m *UserMatcher) WithEmail(matcher specta.Matcher[string]) *UserMatcher {
	m.emailMatcher = matcher
	return m
}

// Implement specta.Matcher[User] interface
func (m *UserMatcher) Matches(u User) specta.MatchResult {
	if m.nameMatcher != nil {
		result := m.nameMatcher.Matches(u.Name)
		if !result.Matched {
			return specta.MatchResult{Matched: false, Message: "name did not match"}
		}
	}
	if m.emailMatcher != nil {
		result := m.emailMatcher.Matches(u.Email)
		if !result.Matched {
			return specta.MatchResult{Matched: false, Message: "email did not match"}
		}
	}
	return specta.MatchResult{Matched: true}
}

// Common test variables
var (
	value         = 42
	expectedUser  = User{ID: "user-123", Name: "Alice", Email: "alice@example.com", Age: 30}
	expected      = 42  // Changed to int to match 'value' type
	list          = []string{"a", "b", "c", "apple", "banana", "cherry"}
	emptyList     = []string{}
	name          = "Alice"
	age           = 30
	score         = 85
	count         = 5
	message       = "error occurred"
	filename      = "test_file.go"
	email         = "user@example.com"
	status        = "active"
	numbers       = []int{2, 4, 6, 8}
	user          = User{
		ID:     "user-123",
		Name:   "Alice",
		Email:  "alice@example.com",
		Age:    30,
		Active: true,
	}
	ptr *int // nil pointer for testing
)
`
}

func TestDocumentationExamples(t *testing.T) {
	docFiles, err := findDocFiles()
	if err != nil {
		t.Fatalf("failed to find doc files: %v", err)
	}

	if len(docFiles) == 0 {
		t.Skip("no documentation files found")
	}

	for _, docPath := range docFiles {
		docPath := docPath // capture
		testName := testNameFromPath(docPath)
		t.Run(testName, func(t *testing.T) {
			testDocFile(t, docPath)
		})
	}
}

func testDocFile(t *testing.T, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", filePath, err)
	}

	blocks := extractGoCodeBlocks(string(content))
	if len(blocks) == 0 {
		t.Skipf("no code blocks found in %s", filePath)
	}

	t.Logf("Found %d Go code blocks in %s", len(blocks), filePath)

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Initialize a test module
	modFile := filepath.Join(tmpDir, "go.mod")
	modContent := `module doctest

go 1.22

require github.com/james-w/specta v0.0.0
`
	if err := os.WriteFile(modFile, []byte(modContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Point to the parent directory for specta
	workFile := filepath.Join(tmpDir, "go.work")
	spectaPath, _ := filepath.Abs("..")
	workContent := `go 1.22

use .
replace github.com/james-w/specta => ` + spectaPath + `
`
	if err := os.WriteFile(workFile, []byte(workContent), 0644); err != nil {
		t.Fatalf("failed to write go.work: %v", err)
	}

	// Write fixtures file
	fixturesFile := filepath.Join(tmpDir, "fixtures.go")
	if err := os.WriteFile(fixturesFile, []byte(getFixturesCode()), 0644); err != nil {
		t.Fatalf("failed to write fixtures: %v", err)
	}

	// Test each code block
	for i, block := range blocks {
		block := block // capture
		i := i

		t.Run(fmt.Sprintf("block_%d", i), func(t *testing.T) {
			t.Logf("Testing %s block %d:\n%s", filePath, i, block)

			// Wrap the code block
			wrapped := wrapCodeBlock(block, i)

			// Write to a test file
			testFile := filepath.Join(tmpDir, "example_test.go")
			if err := os.WriteFile(testFile, []byte(wrapped), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Run the tests (not just compile)
			cmd := exec.Command("go", "test", "-v", "./...")
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("FAILED: %s block %d\n\nOriginal code:\n%s\n\nGenerated code:\n%s\n\nError:\n%s",
					filePath, i, block, wrapped, string(output))
			} else {
				t.Logf("PASSED: %s block %d", filePath, i)
			}
		})
	}
}
