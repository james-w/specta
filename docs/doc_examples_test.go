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

// extractSetupCode extracts setup code from <!-- setup ... --> comment in markdown
func extractSetupCode(content string) string {
	// Match <!-- setup ... --> blocks
	re := regexp.MustCompile(`(?s)<!--\s*setup\s*\n(.*?)\n\s*-->`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		return matches[1]
	}

	// Return minimal default if no setup found
	return `package doctest

import (
	"testing"
	"github.com/james-w/specta"
)
`
}

// CodeBlockType indicates how a code block should be validated
type CodeBlockType int

const (
	CodeBlockTest        CodeBlockType = iota // Full test execution (default)
	CodeBlockCompileOnly                      // Type checking only
	CodeBlockSkip                             // No validation
)

// CodeBlock represents an extracted code block with its validation type
type CodeBlock struct {
	Code string
	Type CodeBlockType
}

// extractGoCodeBlocks extracts all ```go code blocks from a markdown file
// Returns blocks with their validation type based on preceding markers:
// - <!-- skip-test --> : no validation
// - <!-- compile-only --> : type checking only
// - (no marker) : full test execution
func extractGoCodeBlocks(content string) []CodeBlock {
	// Match ```go ... ``` blocks with optional marker
	// Allow whitespace (including indentation) before the marker and code fence
	re := regexp.MustCompile("(?s)[ \\t]*(<!--\\s*(skip-test|compile-only)\\s*-->\\s*\n)?[ \\t]*```go\n(.*?)```")
	matches := re.FindAllStringSubmatch(content, -1)

	var blocks []CodeBlock
	for _, match := range matches {
		if len(match) > 3 {
			marker := match[2] // "skip-test", "compile-only", or ""
			code := match[3]   // The actual code

			var blockType CodeBlockType
			switch marker {
			case "skip-test":
				blockType = CodeBlockSkip
			case "compile-only":
				blockType = CodeBlockCompileOnly
			default:
				blockType = CodeBlockTest
			}

			if blockType != CodeBlockSkip {
				blocks = append(blocks, CodeBlock{Code: code, Type: blockType})
			}
		}
	}
	return blocks
}

// wrapCodeBlock wraps a code snippet in a minimal test file
// For compile-only blocks, it returns raw code without test wrapper
func wrapCodeBlock(code string, index int, blockType CodeBlockType) string {
	// Strip any existing package declaration - we'll use doctest
	packageRe := regexp.MustCompile(`(?m)^package\s+\w+\s*$`)
	code = packageRe.ReplaceAllString(code, "")

	// Check what the code contains
	hasImport := strings.Contains(code, "import")
	hasFunc := strings.Contains(code, "func Test") || strings.Contains(code, "func Example") || strings.Contains(code, "func ")

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

	// For compile-only, just emit the code as-is (already has functions or is top-level)
	// Don't add imports for compile-only - let the code use what it needs
	if blockType == CodeBlockCompileOnly {
		sb.WriteString(strings.TrimSpace(code))
		sb.WriteString("\n")
		return sb.String()
	}

	// Add standard imports (only for test blocks)
	sb.WriteString(`import (
	"testing"
	"github.com/james-w/specta"
)

`)

	// For test blocks: if it's not already a function, wrap it in a test
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

	contentStr := string(content)
	blocks := extractGoCodeBlocks(contentStr)
	if len(blocks) == 0 {
		t.Skipf("no code blocks found in %s", filePath)
	}

	t.Logf("Found %d Go code blocks in %s", len(blocks), filePath)

	// Extract setup code from markdown comment
	setupCode := extractSetupCode(contentStr)

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Initialize a test module
	modFile := filepath.Join(tmpDir, "go.mod")
	modContent := `module doctest

go 1.24.0

require github.com/james-w/specta v0.0.0
`
	if err := os.WriteFile(modFile, []byte(modContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Point to the parent directory for specta
	workFile := filepath.Join(tmpDir, "go.work")
	spectaPath, _ := filepath.Abs("..")
	workContent := `go 1.24.0

use .
replace github.com/james-w/specta => ` + spectaPath + `
`
	if err := os.WriteFile(workFile, []byte(workContent), 0644); err != nil {
		t.Fatalf("failed to write go.work: %v", err)
	}

	// Write setup file with code extracted from markdown
	setupFile := filepath.Join(tmpDir, "setup.go")
	if err := os.WriteFile(setupFile, []byte(setupCode), 0644); err != nil {
		t.Fatalf("failed to write setup: %v", err)
	}

	// Test each code block
	for i, block := range blocks {
		block := block // capture
		i := i

		var suffix string
		switch block.Type {
		case CodeBlockTest:
			suffix = "_test"
		case CodeBlockCompileOnly:
			suffix = "_compile"
		}

		t.Run(fmt.Sprintf("block_%d%s", i, suffix), func(t *testing.T) {
			t.Logf("Testing %s block %d (type: %v):\n%s", filePath, i, block.Type, block.Code)

			// Wrap the code block
			wrapped := wrapCodeBlock(block.Code, i, block.Type)

			// Write to a test file
			testFile := filepath.Join(tmpDir, "example_test.go")
			if err := os.WriteFile(testFile, []byte(wrapped), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			var cmd *exec.Cmd
			var cmdDesc string
			if block.Type == CodeBlockCompileOnly {
				// For compile-only, just build (don't run tests)
				cmd = exec.Command("go", "build", "./...")
				cmdDesc = "compile"
			} else {
				// For test blocks, run the tests
				cmd = exec.Command("go", "test", "-v", "./...")
				cmdDesc = "test"
			}
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("FAILED (%s): %s block %d\n\nOriginal code:\n%s\n\nGenerated code:\n%s\n\nError:\n%s",
					cmdDesc, filePath, i, block.Code, wrapped, string(output))
			} else {
				t.Logf("PASSED (%s): %s block %d", cmdDesc, filePath, i)
			}
		})
	}
}
