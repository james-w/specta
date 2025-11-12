package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	testgen "github.com/james-w/specta"
	"golang.org/x/tools/go/packages"
)

// TestQualifyTypeExpr tests the qualifyTypeExpr function with real Go code
func TestQualifyTypeExpr(t *testing.T) {
	// Create a test package with various type patterns
	testCode := `
package main

import "time"

type User struct {
	Name string
}

type Product struct {
	ID int
}

type TestType struct {
	// Simple types
	SimpleCustom User
	SimplePrimitive string

	// Pointer types
	PointerCustom *User
	PointerPrimitive *string

	// Slice types
	SliceCustom []User
	SlicePrimitive []string
	SlicePointer []*User

	// Array types
	ArrayCustom [10]User
	ArrayPrimitive [5]string

	// Map types
	MapStringToCustom map[string]User
	MapCustomToString map[User]string
	MapCustomToCustom map[User]Product
	MapPrimitive map[string]int

	// Channel types
	ChanCustom chan User
	ChanReceive <-chan User
	ChanSend chan<- User
	ChanPrimitive chan string

	// Function types
	FuncSimple func(User) error
	FuncMultiParam func(User, Product) error
	FuncMultiReturn func(User) (Product, error)
	FuncComplex func([]User) map[string]Product

	// Standard library types
	TimeField time.Time
	DurationField time.Duration
}
`

	// Load the package with type information using a temporary directory
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.go"

	// Write test file to temp directory
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Write a minimal go.mod
	goMod := "module testpkg\n\ngo 1.21\n"
	if err := os.WriteFile(tmpDir+"/go.mod", []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		Dir:  tmpDir,
	}

	// Load from the temp directory
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("No packages loaded")
	}
	if len(pkgs[0].Errors) > 0 {
		t.Fatalf("Package has errors: %v", pkgs[0].Errors)
	}

	pkg := pkgs[0]

	// Use the AST from the loaded package (not from parser.ParseFile)
	if len(pkg.Syntax) == 0 {
		t.Fatal("No syntax files in package")
	}
	file := pkg.Syntax[0]

	// Find the TestType struct
	var testStruct *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == "TestType" {
			if st, ok := ts.Type.(*ast.StructType); ok {
				testStruct = st
				return false
			}
		}
		return true
	})

	if testStruct == nil {
		t.Fatal("Could not find TestType struct")
	}

	tests := []struct {
		fieldName string
		want      string
	}{
		// Simple types
		{"SimpleCustom", "testpkg.User"},
		{"SimplePrimitive", "string"},

		// Pointer types
		{"PointerCustom", "*testpkg.User"},
		{"PointerPrimitive", "*string"},

		// Slice types
		{"SliceCustom", "[]testpkg.User"},
		{"SlicePrimitive", "[]string"},
		{"SlicePointer", "[]*testpkg.User"},

		// Array types
		{"ArrayCustom", "[10]testpkg.User"},
		{"ArrayPrimitive", "[5]string"},

		// Map types
		{"MapStringToCustom", "map[string]testpkg.User"},
		{"MapCustomToString", "map[testpkg.User]string"},
		{"MapCustomToCustom", "map[testpkg.User]testpkg.Product"},
		{"MapPrimitive", "map[string]int"},

		// Channel types
		{"ChanCustom", "chan testpkg.User"},
		{"ChanReceive", "<-chan testpkg.User"},
		{"ChanSend", "chan<- testpkg.User"},
		{"ChanPrimitive", "chan string"},

		// Function types
		{"FuncSimple", "func(testpkg.User) error"},
		{"FuncMultiParam", "func(testpkg.User, testpkg.Product) error"},
		{"FuncMultiReturn", "func(testpkg.User) (testpkg.Product, error)"},
		{"FuncComplex", "func([]testpkg.User) map[string]testpkg.Product"},

		// Standard library types
		{"TimeField", "time.Time"},
		{"DurationField", "time.Duration"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			// Find the field in the struct
			var fieldType ast.Expr
			for _, field := range testStruct.Fields.List {
				if len(field.Names) > 0 && field.Names[0].Name == tt.fieldName {
					fieldType = field.Type
					break
				}
			}

			if fieldType == nil {
				t.Fatalf("Field %s not found", tt.fieldName)
			}

			// Test qualifyTypeExpr
			got := qualifyTypeExpr(pkg, fieldType, "testpkg")

			// Normalize whitespace for comparison
			gotNorm := strings.Join(strings.Fields(got), " ")
			wantNorm := strings.Join(strings.Fields(tt.want), " ")

			testgen.AssertThat(t, gotNorm, testgen.Equal(wantNorm))
		})
	}
}

func TestQualifyTypeExprWithNilTypeInfo(t *testing.T) {
	// Test fallback behavior when TypesInfo is nil
	testCode := `package test
type User struct { Name string }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, 0)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Create package without type info
	pkg := &packages.Package{
		Fset:   fset,
		Syntax: []*ast.File{file},
		// TypesInfo intentionally nil
	}

	// Find User type
	var userIdent *ast.Ident
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == "User" {
			userIdent = ts.Name
			return false
		}
		return true
	})

	// Should fall back to string representation
	result := qualifyTypeExpr(pkg, userIdent, "test")
	testgen.AssertThat(t, result, testgen.Not(testgen.Equal("")))
}

func TestIsPrimitiveType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		// All numeric types
		{"int", "int", true},
		{"int8", "int8", true},
		{"int16", "int16", true},
		{"int32", "int32", true},
		{"int64", "int64", true},
		{"uint", "uint", true},
		{"uint8", "uint8", true},
		{"uint16", "uint16", true},
		{"uint32", "uint32", true},
		{"uint64", "uint64", true},
		{"uintptr", "uintptr", true},
		{"byte", "byte", true},
		{"rune", "rune", true},
		{"float32", "float32", true},
		{"float64", "float64", true},
		{"complex64", "complex64", true},
		{"complex128", "complex128", true},

		// Other built-in types
		{"bool", "bool", true},
		{"string", "string", true},
		{"error", "error", true},

		// Standard library types that don't need qualification
		{"time.Time", "time.Time", true},
		{"time.Duration", "time.Duration", true},

		// Custom types (not primitive)
		{"User", "User", false},
		{"showcase.User", "showcase.User", false},
		{"pointer", "*string", false},
		{"slice", "[]int", false},
		{"map", "map[string]int", false},
		{"channel", "chan int", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrimitiveType(tt.typeName)
			testgen.AssertThat(t, got, testgen.Equal(tt.want))
		})
	}
}
