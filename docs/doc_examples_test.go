package docs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

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
	// Check if code already has package declaration
	hasPackage := strings.Contains(code, "package ")
	hasImport := strings.Contains(code, "import")
	hasFunc := strings.Contains(code, "func Test") || strings.Contains(code, "func Example")

	var sb strings.Builder

	if !hasPackage {
		sb.WriteString("package doctest\n\n")
	}

	if !hasImport && (strings.Contains(code, "specta.") || strings.Contains(code, "AssertThat")) {
		sb.WriteString(`import (
	"testing"
	. "github.com/james-w/specta"
)

`)
	}

	// If it's not already a function, wrap it in a test
	if !hasFunc {
		sb.WriteString("func TestDocExample")
		sb.WriteString(strings.Trim(strings.Title(strings.ReplaceAll(filepath.Base(""), "-", " ")), " "))
		sb.WriteString("_")
		sb.WriteString(string(rune(index)))
		sb.WriteString("(t *testing.T) {\n")

		// Indent the code
		lines := strings.Split(code, "\n")
		for _, line := range lines {
			if line != "" {
				sb.WriteString("\t")
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}

		sb.WriteString("}\n")
	} else {
		sb.WriteString(code)
	}

	return sb.String()
}

func TestIntroductionExamples(t *testing.T) {
	content, err := os.ReadFile("content/docs/introduction/_index.md")
	if err != nil {
		t.Fatalf("failed to read introduction: %v", err)
	}

	blocks := extractGoCodeBlocks(string(content))
	if len(blocks) == 0 {
		t.Skip("no code blocks found in introduction")
	}

	t.Logf("Found %d Go code blocks in introduction", len(blocks))

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "doctest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

	// Test each code block
	for i, block := range blocks {
		block := block // capture
		i := i

		t.Run(string(rune('A'+i)), func(t *testing.T) {
			t.Logf("Testing code block %d:\n%s", i, block)

			// Wrap the code block
			wrapped := wrapCodeBlock(block, i)

			// Write to a test file
			testFile := filepath.Join(tmpDir, "example_test.go")
			if err := os.WriteFile(testFile, []byte(wrapped), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Try to compile it
			cmd := exec.Command("go", "build", "./...")
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Code block %d failed to compile:\n%s\n\nGenerated code:\n%s\n\nError:\n%s",
					i, block, wrapped, string(output))
			} else {
				t.Logf("Code block %d compiled successfully", i)
			}
		})
	}
}
