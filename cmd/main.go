package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v3"
)

type MatcherField struct {
	Name   string `yaml:"name"`
	Getter string `yaml:"getter"`
}

type DefaultProvider struct {
	Pattern string `yaml:"pattern"` // "email", "url", "uuid", or ""
	Custom  string `yaml:"custom"`  // Custom func code (if pattern is empty)
}

type TypeConfig struct {
	Name        string                     `yaml:"name"`
	Constructor string                     `yaml:"constructor"`
	Defaults    map[string]DefaultProvider `yaml:"defaults"`
	Matchers    []MatcherField             `yaml:"matchers"`
}

type Config struct {
	Version int          `yaml:"version"`
	Types   []TypeConfig `yaml:"types"`
	Targets []struct {
		Package string `yaml:"package"`
		Types   struct {
			Include []string `yaml:"include"`
		} `yaml:"types"`
		FileSuffix string `yaml:"file_suffix"`
	} `yaml:"targets"`
}

var cfgPath = flag.String("config", "testgen.yaml", "path to config (JSON for this skeleton)")

func main() {
	flag.Parse()
	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range cfg.Targets {
		if err := processTarget(cfg, t); err != nil {
			log.Fatalf("target %s: %v", t.Package, err)
		}
	}
	log.Println("ok")
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	// NOTE: For brevity this sample expects JSON; swap to yaml.v3 in your repo.
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if err := validateConfig(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func validateConfig(c *Config) error {
	// Validate global types configuration
	for i, typeCfg := range c.Types {
		if typeCfg.Name == "" {
			return fmt.Errorf("types[%d]: name is required", i)
		}

		// Validate matchers
		for j, matcher := range typeCfg.Matchers {
			if matcher.Name == "" {
				return fmt.Errorf("types[%d].matchers[%d]: name is required (for type %s)", i, j, typeCfg.Name)
			}
			if matcher.Getter == "" {
				return fmt.Errorf("types[%d].matchers[%d]: getter is required (for type %s)", i, j, typeCfg.Name)
			}
		}

		// Validate defaults
		for paramName, provider := range typeCfg.Defaults {
			if provider.Pattern == "" && provider.Custom == "" {
				return fmt.Errorf("types[%d].defaults[%s]: must specify either pattern or custom (for type %s)", i, paramName, typeCfg.Name)
			}
			if provider.Pattern != "" && provider.Custom != "" {
				return fmt.Errorf("types[%d].defaults[%s]: cannot specify both pattern and custom (for type %s)", i, paramName, typeCfg.Name)
			}
			if provider.Pattern != "" {
				validPatterns := []string{"email", "url", "uuid"}
				valid := false
				for _, p := range validPatterns {
					if provider.Pattern == p {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("types[%d].defaults[%s]: unknown pattern %q (valid: %v, for type %s)",
						i, paramName, provider.Pattern, validPatterns, typeCfg.Name)
				}
			}
		}
	}

	// Validate each target
	for i, target := range c.Targets {
		if target.Package == "" {
			return fmt.Errorf("targets[%d]: package is required", i)
		}
		if target.FileSuffix == "" {
			return fmt.Errorf("targets[%d]: file_suffix is required", i)
		}

		// Check that types configured in global types section are in the include list
		includeSet := make(map[string]bool)
		for _, typeName := range target.Types.Include {
			includeSet[typeName] = true
		}

		// Note: If a type is configured with constructor or matchers, it should be in the include list for at least one target.
		// We could validate this per-target to ensure configured types are actually used.
		// For now, we skip this check as types might be used in other targets.
	}
	return nil
}

type field struct {
	Name                string
	TypeExpr            string // qualified type expression (e.g., "showcase.User")
	UnqualifiedTypeName string // unqualified type name (e.g., "User") for function/type references
	IsCustomType        bool   // true if this is a local struct type (not primitive)
	RecipeName          string // e.g., "UserRecipe" if IsCustomType
}

type ConstructorParam struct {
	Name                string
	TypeExpr            string // qualified type expression
	UnqualifiedTypeName string // unqualified type name for function/type references
	IsCustomType        bool
}

type GetterInfo struct {
	Name       string // "Name" - the matcher method name
	Getter     string // "GetName" - the getter method name
	ReturnType string // "string" - the return type
}

type data struct {
	Package       string // "factory"
	ParentPackage string // e.g., "example"
	ParentImport  string // import path to parent package
	TypeName      string
	SpecName      string
	RecipeName    string
	Fields        []field
	ImportTime    bool
	ImportTestgen string
	ConfigPath    string // path to the config file for go:generate directive

	// Constructor-based generation
	HasConstructor     bool
	ConstructorName    string
	ConstructorParams  []ConstructorParam
	ConstructorReturns []string // return types from constructor, e.g., ["Email", "error"]

	// Matcher-based generation for getters
	GetterMatchers []GetterInfo

	// Custom defaults for constructor parameters
	CustomDefaults map[string]string
}

func findTypeConfig(cfg *Config, typeName string) *TypeConfig {
	for i := range cfg.Types {
		if cfg.Types[i].Name == typeName {
			return &cfg.Types[i]
		}
	}
	return nil
}

func resolveDefaultProvider(paramName, paramType string, provider DefaultProvider) (string, error) {
	if provider.Pattern != "" {
		// Validate pattern matches type
		switch provider.Pattern {
		case "email":
			if paramType != "string" {
				return "", fmt.Errorf("pattern 'email' requires string type, got %s", paramType)
			}
			return `func(p testgen.Primitives) string { return p.StringWith("user") + "@example.com" }`, nil
		case "url":
			if paramType != "string" {
				return "", fmt.Errorf("pattern 'url' requires string type, got %s", paramType)
			}
			return `func(p testgen.Primitives) string { return "https://example.com/" + p.StringWith("path") }`, nil
		case "uuid":
			if paramType != "string" {
				return "", fmt.Errorf("pattern 'uuid' requires string type, got %s", paramType)
			}
			return `func(p testgen.Primitives) string { return p.UUID().String() }`, nil
		default:
			return "", fmt.Errorf("unknown pattern: %s", provider.Pattern)
		}
	}
	// Use custom function
	if provider.Custom != "" {
		return provider.Custom, nil
	}
	return "", fmt.Errorf("provider must specify either pattern or custom")
}

func findConstructor(pkg *packages.Package, constructorName string) *ast.FuncDecl {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Name == constructorName {
				return fn
			}
		}
	}
	return nil
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func analyzeConstructorParams(fn *ast.FuncDecl, pkg *packages.Package, typeNames map[string]bool) []ConstructorParam {
	if fn.Type == nil || fn.Type.Params == nil {
		return nil
	}

	var params []ConstructorParam
	for _, param := range fn.Type.Params.List {
		if len(param.Names) == 0 {
			continue // skip unnamed params for now
		}

		// Use AST-based type qualification
		typeExpr := qualifyTypeExpr(pkg, param.Type, pkg.Name)
		if typeExpr == "" {
			// Fallback if qualification fails
			var buf bytes.Buffer
			if err := printer.Fprint(&buf, pkg.Fset, param.Type); err != nil {
				continue
			}
			typeExpr = buf.String()
		}

		// Extract unqualified type name for function/type references
		unqualifiedTypeName := extractUnqualifiedTypeName(typeExpr)

		// Check if custom type
		isCustomType := typeNames[unqualifiedTypeName]

		for _, name := range param.Names {
			params = append(params, ConstructorParam{
				Name:                capitalizeFirst(name.Name),
				TypeExpr:            typeExpr,
				UnqualifiedTypeName: unqualifiedTypeName,
				IsCustomType:        isCustomType,
			})
		}
	}

	return params
}

func analyzeConstructorReturns(fn *ast.FuncDecl, pkg *packages.Package) []string {
	if fn.Type == nil || fn.Type.Results == nil {
		return nil
	}

	var returns []string
	for _, result := range fn.Type.Results.List {
		// Render the type expression
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, pkg.Fset, result.Type); err != nil {
			continue
		}
		typeExpr := buf.String()

		// If there are multiple names (e.g., a, b int), add one entry per name
		// But constructor returns typically don't have names, so len(result.Names) == 0
		if len(result.Names) == 0 {
			returns = append(returns, typeExpr)
		} else {
			for range result.Names {
				returns = append(returns, typeExpr)
			}
		}
	}

	return returns
}

// findGetterMethod finds a method on the given type
func findGetterMethod(pkg *packages.Package, typeName string, methodName string) *ast.FuncDecl {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}

			// Check if this method is on our type
			recvType := fn.Recv.List[0].Type
			var recvTypeName string
			switch t := recvType.(type) {
			case *ast.Ident:
				recvTypeName = t.Name
			case *ast.StarExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					recvTypeName = ident.Name
				}
			}

			if recvTypeName == typeName && fn.Name.Name == methodName {
				return fn
			}
		}
	}
	return nil
}

// analyzeGetterMethod extracts the return type from a getter method
// Returns both the AST expression and a string representation (for backwards compatibility)
func analyzeGetterMethod(fn *ast.FuncDecl, pkg *packages.Package) (string, ast.Expr) {
	if fn.Type == nil || fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return "", nil
	}

	// Get first return type
	result := fn.Type.Results.List[0]
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, pkg.Fset, result.Type); err != nil {
		return "", nil
	}
	return buf.String(), result.Type
}

func processTarget(cfg *Config, tgt struct {
	Package string `yaml:"package"`
	Types   struct {
		Include []string `yaml:"include"`
	} `yaml:"types"`
	FileSuffix string `yaml:"file_suffix"`
}) error {
	if tgt.FileSuffix == "" {
		tgt.FileSuffix = "_testgen_gen.go"
	}

	pkg, err := loadPackage(tgt.Package)
	if err != nil {
		return err
	}

	typeNames := collectTypeNames(tgt.Types.Include)

	for _, typeName := range tgt.Types.Include {
		if err := processType(cfg, pkg, typeName, typeNames); err != nil {
			return err
		}
	}
	return nil
}

func loadPackage(packagePath string) (*packages.Package, error) {
	pkgCfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		Dir:  ".",
	}
	pkgs, err := packages.Load(pkgCfg, packagePath)
	if err != nil {
		return nil, fmt.Errorf("load error: %v", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages loaded")
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package has errors")
	}
	return pkgs[0], nil
}

func collectTypeNames(include []string) map[string]bool {
	typeNames := make(map[string]bool)
	for _, tn := range include {
		typeNames[tn] = true
	}
	return typeNames
}

func processType(cfg *Config, pkg *packages.Package, typeName string, typeNames map[string]bool) error {
	st := findStruct(pkg.Syntax, typeName)
	if st == nil {
		return fmt.Errorf("type %s not found in package %s (configured in types.include)", typeName, pkg.Name)
	}

	dirs, err := setupOutputDirectories(pkg)
	if err != nil {
		return err
	}

	d := buildData(cfg, pkg, typeName)
	typeCfg := findTypeConfig(cfg, typeName)

	if err := analyzeTypeStructure(cfg, pkg, &d, typeName, typeCfg, st, typeNames); err != nil {
		return err
	}

	if err := analyzeGetterMatchers(&d, pkg, typeName, typeCfg); err != nil {
		return err
	}

	return generateFiles(d, dirs, typeName)
}

func setupOutputDirectories(pkg *packages.Package) (struct{ factory, spec string }, error) {
	factoryDir := filepath.Join(filepath.Dir(pkg.GoFiles[0]), "factory")
	specDir := filepath.Join(factoryDir, "spec")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return struct{ factory, spec string }{}, fmt.Errorf("mkdir factory/spec: %w", err)
	}
	return struct{ factory, spec string }{factoryDir, specDir}, nil
}

func buildData(cfg *Config, pkg *packages.Package, typeName string) data {
	return data{
		Package:       "factory",
		ParentPackage: pkg.Name,
		ParentImport:  pkg.PkgPath,
		TypeName:      typeName,
		SpecName:      typeName + "Spec",
		RecipeName:    typeName + "Recipe",
		ImportTestgen: "github.com/james-w/specta",
		ConfigPath:    *cfgPath,
	}
}

func analyzeTypeStructure(cfg *Config, pkg *packages.Package, d *data, typeName string, typeCfg *TypeConfig, st *ast.StructType, typeNames map[string]bool) error {
	if typeCfg != nil && typeCfg.Constructor != "" {
		return analyzeConstructor(cfg, pkg, d, typeName, typeCfg, typeNames)
	}
	// Field-based generation
	fields := collectFields(pkg, st, typeNames)
	d.Fields = fields
	d.ImportTime = anyHas(fields, "time.Time") || anyHas(fields, "time.Duration")
	return nil
}

func analyzeConstructor(cfg *Config, pkg *packages.Package, d *data, typeName string, typeCfg *TypeConfig, typeNames map[string]bool) error {
	ctor := findConstructor(pkg, typeCfg.Constructor)
	if ctor == nil {
		return fmt.Errorf("type %s: constructor %s not found in package %s (configured in types.%s.constructor)",
			typeName, typeCfg.Constructor, pkg.Name, typeName)
	}

	d.HasConstructor = true
	d.ConstructorName = typeCfg.Constructor
	d.ConstructorParams = analyzeConstructorParams(ctor, pkg, typeNames)
	d.ConstructorReturns = analyzeConstructorReturns(ctor, pkg)

	// Validate constructor returns the expected type
	if len(d.ConstructorReturns) == 0 {
		return fmt.Errorf("type %s: constructor %s must return at least one value", typeName, typeCfg.Constructor)
	}
	if d.ConstructorReturns[0] != typeName {
		return fmt.Errorf("type %s: constructor %s returns %s, expected %s as first return value",
			typeName, typeCfg.Constructor, d.ConstructorReturns[0], typeName)
	}

	// Check if any params use time types
	for _, param := range d.ConstructorParams {
		if strings.Contains(param.TypeExpr, "time.") {
			d.ImportTime = true
			break
		}
	}

	// Resolve custom defaults
	if len(typeCfg.Defaults) > 0 {
		d.CustomDefaults = make(map[string]string)
		for _, param := range d.ConstructorParams {
			if provider, ok := typeCfg.Defaults[strings.ToLower(param.Name)]; ok {
				resolved, err := resolveDefaultProvider(param.Name, param.TypeExpr, provider)
				if err != nil {
					return fmt.Errorf("type %s: parameter %s: %w", typeName, param.Name, err)
				}
				d.CustomDefaults[param.Name] = resolved
			}
		}
	}

	return nil
}

func analyzeGetterMatchers(d *data, pkg *packages.Package, typeName string, typeCfg *TypeConfig) error {
	if typeCfg == nil || len(typeCfg.Matchers) == 0 {
		return nil
	}

	for _, mf := range typeCfg.Matchers {
		getter := findGetterMethod(pkg, typeName, mf.Getter)
		if getter == nil {
			return fmt.Errorf("type %s: getter method %s not found (configured in types.%s.matchers)",
				typeName, mf.Getter, typeName)
		}
		returnType, returnTypeExpr := analyzeGetterMethod(getter, pkg)
		if returnType == "" || returnTypeExpr == nil {
			return fmt.Errorf("type %s: getter method %s has invalid signature (expected exactly 1 return value)",
				typeName, mf.Getter)
		}

		// Qualify custom types with package name using go/types
		qualifiedReturnType := qualifyTypeExpr(pkg, returnTypeExpr, d.ParentPackage)

		d.GetterMatchers = append(d.GetterMatchers, GetterInfo{
			Name:       mf.Name,
			Getter:     mf.Getter,
			ReturnType: qualifiedReturnType,
		})
	}
	return nil
}

// qualifyTypeExpr uses go/types to properly qualify a type expression.
// This handles all Go type patterns correctly: functions, arrays, maps, channels, etc.
func qualifyTypeExpr(pkg *packages.Package, expr ast.Expr, parentPackage string) string {
	// Get the type information from the type checker
	var typ types.Type
	if pkg.TypesInfo != nil {
		typ = pkg.TypesInfo.TypeOf(expr)
	}
	if typ == nil {
		// Fallback: use printer to get string representation
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, pkg.Fset, expr); err != nil {
			return ""
		}
		return buf.String()
	}

	// Create a qualifier function that adds package prefix for all types.
	// We use types.RelativeTo(nil) to get a qualifier that qualifies all packages
	// with their full import path.
	qualifier := types.RelativeTo(nil) // nil package means qualify everything

	result := types.TypeString(typ, qualifier)

	// Replace the full package path with the desired package name
	// For example, "github.com/james-w/specta/showcase.User" becomes "showcase.User"
	pkgPath := pkg.PkgPath
	if pkgPath != "" {
		result = strings.ReplaceAll(result, pkgPath+".", parentPackage+".")
	}

	return result
}

func isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		// Numeric types
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"uintptr": true,
		"byte":    true, "rune": true,
		"float32": true, "float64": true,
		"complex64": true, "complex128": true,
		// Other built-in types
		"bool":   true,
		"string": true,
		"error":  true,
		// Special types from standard library that don't need qualification
		"time.Time":     true,
		"time.Duration": true,
	}
	return primitives[typeName]
}

func generateFiles(d data, dirs struct{ factory, spec string }, typeName string) error {
	// Define output paths
	specOut := filepath.Join(dirs.spec, strings.ToLower(typeName)+"_gen.go")
	recipeOut := filepath.Join(dirs.factory, strings.ToLower(typeName)+"_gen.go")
	matcherOut := filepath.Join(dirs.factory, strings.ToLower(typeName)+"_matcher_gen.go")

	// Generate all three files in memory
	specSrc, err := generateSpec(d)
	if err != nil {
		return fmt.Errorf("generate spec: %w", err)
	}

	recipeSrc, err := generateRecipe(d)
	if err != nil {
		return fmt.Errorf("generate recipe: %w", err)
	}

	matcherSrc, err := generateMatcher(d)
	if err != nil {
		return fmt.Errorf("generate matcher: %w", err)
	}

	// Verify all three files together using overlays before writing
	if err := verifyFilesAsPackage(map[string][]byte{
		specOut:    specSrc,
		recipeOut:  recipeSrc,
		matcherOut: matcherSrc,
	}); err != nil {
		// Write broken files for debugging
		_ = os.WriteFile(specOut+".broken", specSrc, 0644)
		_ = os.WriteFile(recipeOut+".broken", recipeSrc, 0644)
		_ = os.WriteFile(matcherOut+".broken", matcherSrc, 0644)
		return fmt.Errorf("type-check failed: %v", err)
	}

	// Write all files
	if err := os.MkdirAll(filepath.Dir(specOut), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(specOut, specSrc, 0644); err != nil {
		return err
	}
	log.Printf("wrote %s", specOut)

	if err := os.WriteFile(recipeOut, recipeSrc, 0644); err != nil {
		return err
	}
	log.Printf("wrote %s", recipeOut)

	if err := os.WriteFile(matcherOut, matcherSrc, 0644); err != nil {
		return err
	}
	log.Printf("wrote %s", matcherOut)

	return nil
}

func findStruct(files []*ast.File, typeName string) *ast.StructType {
	var out *ast.StructType
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || ts.Name.Name != typeName {
				return true
			}
			if st, ok := ts.Type.(*ast.StructType); ok {
				out = st
			}
			return false
		})
		if out != nil {
			break
		}
	}
	return out
}

// extractUnqualifiedTypeName extracts the unqualified type name from a qualified type expression
// E.g., "showcase.User" -> "User", "*showcase.User" -> "User", "[]showcase.Item" -> "Item"
func extractUnqualifiedTypeName(typeExpr string) string {
	// Strip pointer/slice prefixes
	baseType := typeExpr
	for strings.HasPrefix(baseType, "*") || strings.HasPrefix(baseType, "[]") {
		if strings.HasPrefix(baseType, "*") {
			baseType = strings.TrimPrefix(baseType, "*")
		}
		if strings.HasPrefix(baseType, "[]") {
			baseType = strings.TrimPrefix(baseType, "[]")
		}
	}
	// Remove package prefix
	if idx := strings.LastIndex(baseType, "."); idx >= 0 {
		return baseType[idx+1:]
	}
	return baseType
}

func collectFields(pkg *packages.Package, st *ast.StructType, typeNames map[string]bool) []field {
	var out []field
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			continue // skip embedded/anon for now
		}
		name := f.Names[0].Name

		// Use AST-based type qualification
		typ := qualifyTypeExpr(pkg, f.Type, pkg.Name)
		if typ == "" {
			// Fallback if qualification fails
			var buf bytes.Buffer
			if err := printer.Fprint(&buf, pkg.Fset, f.Type); err != nil {
				buf.WriteString("interface{}")
			}
			typ = buf.String()
		}

		// Extract unqualified type name for function/type references
		unqualifiedTypeName := extractUnqualifiedTypeName(typ)

		// Check if custom type
		isCustomType := typeNames[unqualifiedTypeName]
		recipeName := ""
		if isCustomType {
			recipeName = unqualifiedTypeName + "Recipe"
		}

		out = append(out, field{
			Name:                name,
			TypeExpr:            typ,
			UnqualifiedTypeName: unqualifiedTypeName,
			IsCustomType:        isCustomType,
			RecipeName:          recipeName,
		})
	}
	return out
}

func anyHas(fields []field, typ string) bool {
	for _, f := range fields {
		if f.TypeExpr == typ {
			return true
		}
	}
	return false
}

// generateSpec generates the spec file source code
func generateSpec(d data) ([]byte, error) {
	var buf bytes.Buffer
	if err := specTmpl.Execute(&buf, d); err != nil {
		return nil, err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("format: %w", err)
	}
	return src, nil
}

// generateRecipe generates the recipe file source code
func generateRecipe(d data) ([]byte, error) {
	var buf bytes.Buffer
	if err := recipeTmpl.Execute(&buf, d); err != nil {
		return nil, err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("format: %w", err)
	}
	return src, nil
}

// generateMatcher generates the matcher file source code
func generateMatcher(d data) ([]byte, error) {
	var buf bytes.Buffer
	if err := matcherTmpl.Execute(&buf, d); err != nil {
		return nil, err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("format: %w", err)
	}
	return src, nil
}

// verifyFilesAsPackage verifies that a set of files compile together as a package
func verifyFilesAsPackage(files map[string][]byte) error {
	if len(files) == 0 {
		return nil
	}

	// Get the directory from the first file
	var dir string
	for path := range files {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("abs path: %w", err)
		}
		dir = filepath.Dir(absPath)
		break
	}

	// Create overlay with absolute paths
	overlay := make(map[string][]byte)
	for path, content := range files {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("abs path: %w", err)
		}
		overlay[absPath] = content
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		Overlay: overlay,
	}

	// Load the package containing these files
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return fmt.Errorf("load for type-check: %w", err)
	}

	// Check for type errors
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			var errs []string
			for _, e := range pkg.Errors {
				errs = append(errs, e.Error())
			}
			return fmt.Errorf("type errors in generated code:\n%s", strings.Join(errs, "\n"))
		}
	}

	return nil
}

func verifyCompiles(src []byte, filename string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}

	// Use packages.Load with an overlay to type-check without writing
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		Dir:  filepath.Dir(absPath),
		Overlay: map[string][]byte{
			absPath: src,
		},
	}

	// Load the package containing this file
	pkgs, err := packages.Load(cfg, filepath.Dir(absPath))
	if err != nil {
		return fmt.Errorf("load for type-check: %w", err)
	}

	// Check for type errors
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			var errs []string
			for _, e := range pkg.Errors {
				errs = append(errs, e.Error())
			}
			return fmt.Errorf("type errors in generated code:\n%s", strings.Join(errs, "\n"))
		}
	}

	return nil
}

var specTmpl = template.Must(template.New("spec").Funcs(template.FuncMap{
	"lower": func(s string) string {
		if s == "" {
			return s
		}
		r := []rune(s)
		r[0] = []rune(strings.ToLower(string(r[0])))[0]
		return string(r)
	},
	"isPrimitiveType": isPrimitiveType,
	"defaultProvider": func(pkg string, f field) string { return defaultProvider(pkg, f) },
	"buildReturnSignature": func(pkg string, typeName string, returns []string) string {
		if len(returns) == 0 {
			// No explicit returns means single return of the type
			return pkg + "." + typeName
		}
		if len(returns) == 1 {
			// Single return - could be just the type, or could be qualified
			// First return is always the constructed type
			return pkg + "." + typeName
		}
		// Multiple returns - (Type, error) or (Type, bool, error) etc
		// First is the type, rest are passed through
		parts := []string{pkg + "." + typeName}
		parts = append(parts, returns[1:]...)
		return "(" + strings.Join(parts, ", ") + ")"
	},
	"qualifiedType": func(pkg string, f field) string {
		// TypeExpr is already qualified by collectFields using qualifyTypeExpr
		return f.TypeExpr
	},
	"qualifiedTypeParam": func(pkg string, p ConstructorParam) string {
		// TypeExpr is already qualified by analyzeConstructorParams using qualifyTypeExpr
		return p.TypeExpr
	},
	"defaultProviderParam": func(pkg string, p ConstructorParam) string {
		// Similar logic to defaultProvider but for ConstructorParam
		if strings.HasPrefix(p.TypeExpr, "[]") {
			return "func(p testgen.Primitives) " + p.TypeExpr + " { return nil }"
		}
		switch p.TypeExpr {
		case "string":
			return "func(p testgen.Primitives) string { return p.StringWith(\"" + strings.ToLower(p.Name) + "_\") }"
		case "int", "int64", "uint64":
			return "func(p testgen.Primitives) " + p.TypeExpr + " { return " + p.TypeExpr + "(p.Int()) }"
		case "bool":
			return "func(p testgen.Primitives) bool { return p.Bool() }"
		case "float64":
			return "func(p testgen.Primitives) float64 { return p.Float64() }"
		case "time.Time":
			return "func(p testgen.Primitives) time.Time { return p.Time() }"
		case "time.Duration":
			return "func(p testgen.Primitives) time.Duration { return p.Duration() }"
		default:
			if p.IsCustomType {
				// Use unqualified name for function references
				return "testgen.FromSpec(Build" + p.UnqualifiedTypeName + ", New" + p.UnqualifiedTypeName + "Spec)"
			}
			return "func(p testgen.Primitives) " + p.TypeExpr + " { return " + p.TypeExpr + "{} }"
		}
	},
}).Parse(`// Code generated by testgen-gen; DO NOT EDIT.
//go:build !ignore_testgen
// +build !ignore_testgen

// Package spec provides low-level specifications for building test data.
//
// Most users should use the high-level Recipe API from the parent factory package instead.
//
// Example usage:
//
//	// Create a spec and set fields explicitly
//	spec := New{{.TypeName}}Spec()
//	opts := []testgen.Opt[{{.TypeName}}Spec]{
//		{{- if .HasConstructor}}
//		{{- with index .ConstructorParams 0}}
//		With{{$.TypeName}}{{.Name}}({{if eq .TypeExpr "string"}}"example"{{else if eq .TypeExpr "int"}}42{{else}}value{{end}}),
//		{{- end}}
//		{{- else}}
//		{{- with index .Fields 0}}
//		With{{$.TypeName}}{{.Name}}({{if eq .TypeExpr "string"}}"example"{{else if eq .TypeExpr "int"}}42{{else}}value{{end}}),
//		{{- end}}
//		{{- end}}
//	}
//	for _, opt := range opts {
//		opt(&spec)
//	}
//	result := Build{{.TypeName}}(testgen.New(), spec)
//
//	// Or use a factory for convenience
//	factory := New{{.TypeName}}Factory(testgen.New())
//	{{- if .HasConstructor}}
//	{{- with index .ConstructorParams 0}}
//	result := factory.Make(With{{$.TypeName}}{{.Name}}({{if eq .TypeExpr "string"}}"example"{{else if eq .TypeExpr "int"}}42{{else}}value{{end}}))
//	{{- end}}
//	{{- else}}
//	{{- with index .Fields 0}}
//	result := factory.Make(With{{$.TypeName}}{{.Name}}({{if eq .TypeExpr "string"}}"example"{{else if eq .TypeExpr "int"}}42{{else}}value{{end}}))
//	{{- end}}
//	{{- end}}
package spec

import (
{{- if .ImportTime }}
	"time"
{{end}}
	testgen "{{.ImportTestgen}}"
	"{{.ParentImport}}"
)

// {{.SpecName}} is the low-level specification for building {{.TypeName}} instances.
// Most users should use {{.TypeName}}Recipe from the parent factory package instead.
type {{.SpecName}} struct {
	{{- if .HasConstructor}}
	{{- range .ConstructorParams}}
	{{.Name}} testgen.Maybe[{{qualifiedTypeParam $.ParentPackage .}}]
	{{- end}}
	{{- else}}
	{{- range .Fields}}
	{{.Name}} testgen.Maybe[{{qualifiedType $.ParentPackage .}}]
	{{- end}}
	{{- end}}
}

// New{{.SpecName}} creates a new {{.SpecName}} with all fields unset.
func New{{.SpecName}}() {{.SpecName}} { return {{.SpecName}}{} }

// New{{.TypeName}}Factory creates a new SpecFactory for {{.TypeName}}.
func New{{.TypeName}}Factory(p testgen.Primitives) *testgen.SpecFactory[{{$.ParentPackage}}.{{.TypeName}}, {{.SpecName}}] {
	{{- if and .HasConstructor (gt (len .ConstructorReturns) 1)}}
	// Wrap error-returning constructor - panic on error for test factories
	wrappedBuild := func(p testgen.Primitives, s {{.SpecName}}) {{$.ParentPackage}}.{{.TypeName}} {
		result, err := Build{{.TypeName}}(p, s)
		if err != nil {
			panic("Build{{.TypeName}} failed: " + err.Error())
		}
		return result
	}
	return testgen.NewSpecFactory(p, New{{.SpecName}}, wrappedBuild)
	{{- else}}
	return testgen.NewSpecFactory(p, New{{.SpecName}}, Build{{.TypeName}})
	{{- end}}
}

{{- if .HasConstructor}}
// Default parameter providers.
var (
	{{- range .ConstructorParams}}
	{{- if $customDefault := index $.CustomDefaults .Name}}
	{{$.TypeName}}Default{{.Name}} = {{ $customDefault }}
	{{- else}}
	{{$.TypeName}}Default{{.Name}} = {{ defaultProviderParam $.ParentPackage . }}
	{{- end}}
	{{- end}}
)

// Build{{.TypeName}} constructs a {{.TypeName}} from a {{.SpecName}}.
func Build{{.TypeName}}(p testgen.Primitives, s {{.SpecName}}) {{buildReturnSignature $.ParentPackage $.TypeName $.ConstructorReturns}} {
	{{- range .ConstructorParams}}
	{{lower .Name}} := s.{{.Name}}.Get(p, {{$.TypeName}}Default{{.Name}})
	{{- end}}
	return {{$.ParentPackage}}.{{$.ConstructorName}}({{range $i, $param := .ConstructorParams}}{{if $i}}, {{end}}{{lower $param.Name}}{{end}})
}

{{range .ConstructorParams}}
// With{{$.TypeName}}{{.Name}} sets the {{.Name}} parameter to a literal value.
func With{{$.TypeName}}{{.Name}}(v {{qualifiedTypeParam $.ParentPackage .}}) testgen.Opt[{{$.SpecName}}] {
	return testgen.SetLit(func(s *{{$.SpecName}}, m testgen.Maybe[{{qualifiedTypeParam $.ParentPackage .}}]) { s.{{.Name}} = m }, v)
}
{{if .IsCustomType}}
// With{{$.TypeName}}{{.Name}}FromProvider sets the {{.Name}} parameter using a Provider (evaluated lazily).
func With{{$.TypeName}}{{.Name}}FromProvider(prov testgen.Provider[{{qualifiedTypeParam $.ParentPackage .}}]) testgen.Opt[{{$.SpecName}}] {
	return testgen.SetWith(
		func(s *{{$.SpecName}}, m testgen.Maybe[{{qualifiedTypeParam $.ParentPackage .}}]) { s.{{.Name}} = m },
		prov,
	)
}
{{end}}
{{end}}
{{- else}}
// Default field providers.
var (
	{{- range .Fields}}
	{{$.TypeName}}Default{{.Name}} = {{ defaultProvider $.ParentPackage . }}
	{{- end}}
)

// Build{{.TypeName}} constructs a {{.TypeName}} from a {{.SpecName}}.
func Build{{.TypeName}}(p testgen.Primitives, s {{.SpecName}}) {{$.ParentPackage}}.{{.TypeName}} {
	{{- range .Fields}}
	{{lower .Name}} := s.{{.Name}}.Get(p, {{$.TypeName}}Default{{.Name}})
	{{- end}}
	return {{$.ParentPackage}}.{{.TypeName}}{
		{{- range .Fields}}
		{{.Name}}: {{lower .Name}},
		{{- end}}
	}
}

{{range .Fields}}
// With{{$.TypeName}}{{.Name}} sets the {{.Name}} field to a literal value.
func With{{$.TypeName}}{{.Name}}(v {{qualifiedType $.ParentPackage .}}) testgen.Opt[{{$.SpecName}}] {
	return testgen.SetLit(func(s *{{$.SpecName}}, m testgen.Maybe[{{qualifiedType $.ParentPackage .}}]) { s.{{.Name}} = m }, v)
}
{{if .IsCustomType}}
// With{{$.TypeName}}{{.Name}}FromProvider sets the {{.Name}} field using a Provider (evaluated lazily).
func With{{$.TypeName}}{{.Name}}FromProvider(prov testgen.Provider[{{qualifiedType $.ParentPackage .}}]) testgen.Opt[{{$.SpecName}}] {
	return testgen.SetWith(
		func(s *{{$.SpecName}}, m testgen.Maybe[{{qualifiedType $.ParentPackage .}}]) { s.{{.Name}} = m },
		prov,
	)
}
{{end}}
{{end}}
{{- end}}
`))

var recipeTmpl = template.Must(template.New("recipe").Funcs(template.FuncMap{
	"lower": func(s string) string {
		if s == "" {
			return s
		}
		r := []rune(s)
		r[0] = []rune(strings.ToLower(string(r[0])))[0]
		return string(r)
	},
	"isPrimitiveType": isPrimitiveType,
	"hasPrefix":       strings.HasPrefix,
	"buildReturnSignature": func(pkg string, typeName string, returns []string) string {
		if len(returns) == 0 {
			return pkg + "." + typeName
		}
		if len(returns) == 1 {
			return pkg + "." + typeName
		}
		parts := []string{pkg + "." + typeName}
		parts = append(parts, returns[1:]...)
		return "(" + strings.Join(parts, ", ") + ")"
	},
	"qualifiedType": func(pkg string, f field) string {
		// TypeExpr is already qualified by collectFields using qualifyTypeExpr
		return f.TypeExpr
	},
	"qualifiedTypeParam": func(pkg string, p ConstructorParam) string {
		// TypeExpr is already qualified by analyzeConstructorParams using qualifyTypeExpr
		return p.TypeExpr
	},
}).Parse(`// Code generated by testgen-gen; DO NOT EDIT.
//go:build !ignore_testgen
// +build !ignore_testgen

// Package factory provides a fluent Recipe API for building test data and matchers.
//
// Example usage:
//
//	// Build a single instance with custom fields
//	{{- if .HasConstructor}}
//	{{- with index .ConstructorParams 0}}
//	result := {{$.TypeName}}().{{.Name}}({{if eq .TypeExpr "string"}}"example"{{else if eq .TypeExpr "int"}}42{{else}}value{{end}}).Build(testgen.New())
//	{{- end}}
//	{{- else}}
//	{{- with index .Fields 0}}
//	result := {{$.TypeName}}().{{.Name}}({{if eq .TypeExpr "string"}}"example"{{else if eq .TypeExpr "int"}}42{{else}}value{{end}}).Build(testgen.New())
//	{{- end}}
//	{{- end}}
//
//	// Build many instances with unique values
//	results := {{$.TypeName}}().Many(5, testgen.New())
//
//	// Build a matcher to verify specific fields
//	{{- if .HasConstructor}}
//	{{- with index .ConstructorParams 0}}
//	matcher := {{$.TypeName}}Matches().{{.Name}}(testgen.DeepEqual({{if eq .TypeExpr "string"}}"expected"{{else if eq .TypeExpr "int"}}42{{else}}expectedValue{{end}}))
//	{{- end}}
//	{{- else}}
//	{{- with index .Fields 0}}
//	matcher := {{$.TypeName}}Matches().{{.Name}}(testgen.DeepEqual({{if eq .TypeExpr "string"}}"expected"{{else if eq .TypeExpr "int"}}42{{else}}expectedValue{{end}}))
//	{{- end}}
//	{{- end}}
//	testgen.AssertThat(t, actual, matcher.Matcher())
//
//	// Convert a recipe to a matcher for partial matching
//	{{- if .HasConstructor}}
//	{{- with index .ConstructorParams 0}}
//	partialMatcher := {{$.TypeName}}().{{.Name}}({{if eq .TypeExpr "string"}}"expected"{{else if eq .TypeExpr "int"}}99{{else}}expected{{end}}).AsEqualMatcher()
//	{{- end}}
//	{{- else}}
//	{{- with index .Fields 0}}
//	partialMatcher := {{$.TypeName}}().{{.Name}}({{if eq .TypeExpr "string"}}"expected"{{else if eq .TypeExpr "int"}}99{{else}}expected{{end}}).AsEqualMatcher()
//	{{- end}}
//	{{- end}}
//	testgen.AssertThat(t, actual, partialMatcher)
package factory

import (
{{- if and .HasConstructor (gt (len .ConstructorReturns) 1) }}
	"fmt"
{{end}}
{{- if .ImportTime }}
	"time"
{{end}}
	testgen "{{.ImportTestgen}}"
	"{{.ParentImport}}"
	"{{.ParentImport}}/factory/spec"
)

// {{.RecipeName}} provides a fluent API for building {{.TypeName}} instances.
type {{.RecipeName}} struct{
	opts []testgen.Opt[spec.{{.SpecName}}]
	{{- if .HasConstructor}}
	{{- range .ConstructorParams}}
	{{- if .IsCustomType}}
	{{- if not (hasPrefix .TypeExpr "[]")}}
	{{- $baseType := .TypeExpr}}
	{{- if hasPrefix .TypeExpr "*"}}
	{{- $baseType = (slice .TypeExpr 1)}}
	{{- end}}
	{{lower .Name}}Recipe *{{.UnqualifiedTypeName}}Recipe
	{{- end}}
	{{- end}}
	{{- end}}
	{{- else}}
	{{- range .Fields}}
	{{- if .IsCustomType}}
	{{- if not (hasPrefix .TypeExpr "[]")}}
	{{lower .Name}}Recipe *{{.RecipeName}}
	{{- end}}
	{{- end}}
	{{- end}}
	{{- end}}
}

// {{.TypeName}} creates a new {{.RecipeName}} for building {{.TypeName}} instances.
{{- if .HasConstructor}}
//
// The underlying constructor is {{.ConstructorName}}({{range $i, $p := .ConstructorParams}}{{if $i}}, {{end}}{{lower $p.Name}} {{qualifiedTypeParam $.ParentPackage $p}}{{end}}).
{{- else}}
//
// This is a struct-based type with the following fields:
{{- range .Fields}}
//   - {{.Name}} ({{qualifiedType $.ParentPackage .}})
{{- end}}
{{- end}}
//
// Example:
//
{{- if .HasConstructor}}
//	p := testgen.New()
//	{{lower .TypeName}} := factory.{{.TypeName}}().
{{- with index .ConstructorParams 0}}
//	    {{.Name}}({{if eq .TypeExpr "string"}}"custom_value"{{else if eq .TypeExpr "int"}}100{{else if eq .TypeExpr "bool"}}true{{else}}value{{end}}).
{{- end}}
{{- if gt (len .ConstructorParams) 1}}
{{- with index .ConstructorParams 1}}
//	    {{.Name}}({{if eq .TypeExpr "string"}}"another"{{else if eq .TypeExpr "int"}}200{{else if eq .TypeExpr "bool"}}false{{else}}value{{end}}).
{{- end}}
{{- end}}
//	    Build(p)
{{- else}}
//	p := testgen.New()
//	{{lower .TypeName}} := factory.{{.TypeName}}().
{{- with index .Fields 0}}
//	    {{.Name}}({{if eq .TypeExpr "string"}}"custom_value"{{else if eq .TypeExpr "int"}}100{{else if eq .TypeExpr "bool"}}true{{else}}value{{end}}).
{{- end}}
{{- if gt (len .Fields) 1}}
{{- with index .Fields 1}}
//	    {{.Name}}({{if eq .TypeExpr "string"}}"another"{{else if eq .TypeExpr "int"}}200{{else if eq .TypeExpr "bool"}}false{{else}}value{{end}}).
{{- end}}
{{- end}}
//	    Build(p)
{{- end}}
func {{.TypeName}}() {{.RecipeName}} { return {{.RecipeName}}{} }

{{- if .HasConstructor}}
{{range .ConstructorParams}}
// {{.Name}} sets the {{lower .Name}} parameter of {{$.ConstructorName}}.
func (r {{$.RecipeName}}) {{.Name}}(v {{qualifiedTypeParam $.ParentPackage .}}) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}(v))
	{{- if .IsCustomType}}
	{{- if not (hasPrefix .TypeExpr "[]")}}
	r.{{lower .Name}}Recipe = nil
	{{- end}}
	{{- end}}
	return r
}
{{if .IsCustomType}}
{{- if not (hasPrefix .TypeExpr "[]")}}
// {{.Name}}FromRecipe sets the {{.Name}} parameter using another Recipe (creates unique instances).
// The recipe is captured at call time (value semantics) - subsequent changes to v won't affect this recipe.
// The nested recipe is used for partial matching in AsEqualMatcher.
func (r {{$.RecipeName}}) {{.Name}}FromRecipe(v {{.UnqualifiedTypeName}}Recipe) {{$.RecipeName}} {
	{{- if hasPrefix .TypeExpr "*"}}
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(testgen.PtrOf(v.Provider())))
	{{- else}}
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(v.Provider()))
	{{- end}}
	r.{{lower .Name}}Recipe = &v
	return r
}
{{- end}}
{{end}}
{{end}}
{{- else}}
{{range .Fields}}
// {{.Name}} sets the {{.Name}} field.
func (r {{$.RecipeName}}) {{.Name}}(v {{qualifiedType $.ParentPackage .}}) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}(v))
	{{- if .IsCustomType}}
	{{- if not (hasPrefix .TypeExpr "[]")}}
	r.{{lower .Name}}Recipe = nil
	{{- end}}
	{{- end}}
	return r
}
{{if .IsCustomType}}
{{- if not (hasPrefix .TypeExpr "[]")}}
// {{.Name}}FromRecipe sets the {{.Name}} field using another Recipe (creates unique instances).
// The recipe is captured at call time (value semantics) - subsequent changes to v won't affect this recipe.
// The nested recipe is used for partial matching in AsEqualMatcher.
func (r {{$.RecipeName}}) {{.Name}}FromRecipe(v {{.RecipeName}}) {{$.RecipeName}} {
	{{- if hasPrefix .TypeExpr "*"}}
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(testgen.PtrOf(v.Provider())))
	{{- else}}
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(v.Provider()))
	{{- end}}
	r.{{lower .Name}}Recipe = &v
	return r
}
{{- end}}
{{end}}
{{end}}
{{- end}}

// Provider returns a Provider for lazy evaluation in parent factories.
func (r {{.RecipeName}}) Provider() testgen.Provider[{{.ParentPackage}}.{{.TypeName}}] {
	{{- if and .HasConstructor (gt (len .ConstructorReturns) 1)}}
	// Wrap error-returning constructor - panic on error for test factories
	return func(p testgen.Primitives) {{.ParentPackage}}.{{.TypeName}} {
		s := spec.New{{.SpecName}}()
		for _, opt := range r.opts {
			opt(&s)
		}
		result, err := spec.Build{{.TypeName}}(p, s)
		if err != nil {
			panic("Provider failed: " + err.Error())
		}
		return result
	}
	{{- else}}
	return testgen.FromSpec(spec.Build{{.TypeName}}, spec.New{{.SpecName}}, r.opts...)
	{{- end}}
}

// Build creates a single {{.TypeName}} instance.
func (r {{.RecipeName}}) Build(p testgen.Primitives) {{buildReturnSignature .ParentPackage .TypeName .ConstructorReturns}} {
	{{- if and .HasConstructor (gt (len .ConstructorReturns) 1)}}
	// Constructor returns multiple values - apply opts and call Build directly
	s := spec.New{{.SpecName}}()
	for _, opt := range r.opts {
		opt(&s)
	}
	return spec.Build{{.TypeName}}(p, s)
	{{- else}}
	return spec.New{{.TypeName}}Factory(p).Make(r.opts...)
	{{- end}}
}

// Many creates multiple {{.TypeName}} instances with unique generated values.
func (r {{.RecipeName}}) Many(n int, p testgen.Primitives) []{{.ParentPackage}}.{{.TypeName}} {
	{{- if and .HasConstructor (gt (len .ConstructorReturns) 1)}}
	// Wrap error-returning constructor - panic on first error
	var results []{{.ParentPackage}}.{{.TypeName}}
	for i := range n {
		item, err := r.Build(p)
		if err != nil {
			panic("Many() failed on item " + fmt.Sprint(i) + ": " + err.Error())
		}
		results = append(results, item)
	}
	return results
	{{- else}}
	return spec.New{{.TypeName}}Factory(p).Many(n, r.opts...)
	{{- end}}
}

// AsEqualMatcher converts this Recipe into a Matcher that checks for equality on all set fields.
// Only fields that were explicitly set in the Recipe will be checked - unset fields are ignored.
// This enables partial matching where you only verify specific fields.
// For nested types set via FromRecipe, partial matching is applied recursively.
func (r {{.RecipeName}}) AsEqualMatcher() testgen.Matcher[{{.ParentPackage}}.{{.TypeName}}] {
	{{- if and .HasConstructor (gt (len .GetterMatchers) 0)}}
	// Constructor-based type with matchers configured: use matcher builder for partial matching
	s := spec.New{{.SpecName}}()
	for _, opt := range r.opts {
		opt(&s)
	}
	p := testgen.New()

	m := {{.TypeName}}Matches()
	{{- range .GetterMatchers}}
	if s.{{.Name}}.IsSet() {
		{{- $matcherName := .Name -}}
		{{- $isCustomType := false -}}
		{{- range $.ConstructorParams -}}
			{{- if eq .Name $matcherName -}}
				{{- if and .IsCustomType (not (hasPrefix .TypeExpr "[]")) -}}
					{{- $isCustomType = true -}}
				{{- end -}}
			{{- end -}}
		{{- end -}}
		{{- if $isCustomType}}
		// Use nested recipe for partial matching if available
		if r.{{lower .Name}}Recipe != nil {
			{{- range $.ConstructorParams -}}
			{{- if eq .Name $matcherName -}}
			{{- if hasPrefix .TypeExpr "*"}}
			m = m.{{.Name}}(testgen.PointsTo(r.{{lower .Name}}Recipe.AsEqualMatcher()))
			{{- else}}
			m = m.{{.Name}}(r.{{lower .Name}}Recipe.AsEqualMatcher())
			{{- end -}}
			{{- end -}}
			{{- end}}
		} else {
			m = m.{{.Name}}(testgen.Equal(s.{{.Name}}.Value(p)))
		}
		{{- else}}
		m = m.{{.Name}}(testgen.Equal(s.{{.Name}}.Value(p)))
		{{- end}}
	}
	{{- end}}
	return m.Matcher()
	{{- else if .HasConstructor}}
	// Constructor-based types without matchers: use deep equality
	s := spec.New{{.SpecName}}()
	for _, opt := range r.opts {
		opt(&s)
	}
	p := testgen.New()
	{{- if gt (len .ConstructorReturns) 1}}
	expected, err := spec.Build{{.TypeName}}(p, s)
	if err != nil {
		panic("AsEqualMatcher: failed to build expected value: " + err.Error())
	}
	{{- else}}
	expected := spec.Build{{.TypeName}}(p, s)
	{{- end}}
	return testgen.DeepEqual(expected)
	{{- else}}
	// Apply opts to a spec to see what was set
	s := spec.New{{.SpecName}}()
	for _, opt := range r.opts {
		opt(&s)
	}

	// Use a dummy Primitives to evaluate literal values
	// This works for SetLit values; SetWith/Provider values will be evaluated too
	p := testgen.New()

	// Build matcher only for set fields
	m := {{.TypeName}}Matches()
	{{- range .Fields}}
	if s.{{.Name}}.IsSet() {
		{{- if .IsCustomType}}
		{{- if not (hasPrefix .TypeExpr "[]")}}
		// Check if we have a nested recipe for partial matching
		if r.{{lower .Name}}Recipe != nil {
			{{- if hasPrefix .TypeExpr "*"}}
			m = m.{{.Name}}(testgen.PointsTo(r.{{lower .Name}}Recipe.AsEqualMatcher()))
			{{- else}}
			m = m.{{.Name}}(r.{{lower .Name}}Recipe.AsEqualMatcher())
			{{- end}}
		} else {
			m = m.{{.Name}}(testgen.DeepEqual(s.{{.Name}}.Value(p)))
		}
		{{- else}}
		m = m.{{.Name}}(testgen.DeepEqual(s.{{.Name}}.Value(p)))
		{{- end}}
		{{- else}}
		m = m.{{.Name}}(testgen.DeepEqual(s.{{.Name}}.Value(p)))
		{{- end}}
	}
	{{- end}}

	return m.Matcher()
	{{- end}}
}
`))

var matcherTmpl = template.Must(template.New("matcher").Funcs(template.FuncMap{
	"isPrimitiveType": isPrimitiveType,
	"lower": func(s string) string {
		if s == "" {
			return s
		}
		r := []rune(s)
		r[0] = []rune(strings.ToLower(string(r[0])))[0]
		return string(r)
	},
	"qualifiedReturnType": func(pkg string, returnType string) string {
		// If already qualified (contains .) or is a primitive, return as-is
		if strings.Contains(returnType, ".") {
			return returnType
		}
		// Check if it's a known primitive type
		if isPrimitiveType(returnType) {
			return returnType
		}
		// Handle slices
		if strings.HasPrefix(returnType, "[]") {
			elemType := strings.TrimPrefix(returnType, "[]")
			if strings.Contains(elemType, ".") {
				return returnType
			}
			if isPrimitiveType(elemType) {
				return returnType
			}
			return "[]" + pkg + "." + elemType
		}
		// Otherwise, qualify with package
		return pkg + "." + returnType
	},
	"hasPrefix": strings.HasPrefix,
	"qualifiedType": func(pkg string, f field) string {
		// TypeExpr is already qualified by collectFields using qualifyTypeExpr
		return f.TypeExpr
	},
}).Parse(`// Code generated by testgen-gen; DO NOT EDIT.
//go:build !ignore_testgen
// +build !ignore_testgen

package factory

import (
{{- if .ImportTime }}
	"time"
{{end}}
	testgen "{{.ImportTestgen}}"
	"{{.ParentImport}}"
)

// {{.TypeName}}Matcher provides a fluent API for matching {{.TypeName}} instances.
type {{.TypeName}}Matcher struct {
	{{- range .Fields}}
	{{lower .Name}}Matcher testgen.Matcher[{{qualifiedType $.ParentPackage .}}]
	{{- end}}
	{{- range .GetterMatchers}}
	{{lower .Name}}Matcher testgen.Matcher[{{qualifiedReturnType $.ParentPackage .ReturnType}}]
	{{- end}}
}

// {{.TypeName}}Matches creates a new {{.TypeName}}Matcher for matching {{.TypeName}} instances.
{{- if .GetterMatchers}}
//
// This matcher provides methods for properties accessed via getters:
{{- range .GetterMatchers}}
//   - {{.Name}}: matches {{.Getter}}() → {{.ReturnType}}
{{- end}}
{{- end}}
{{- if .Fields}}
//
// This matcher provides methods for the following fields:
{{- range .Fields}}
//   - {{.Name}} ({{qualifiedType $.ParentPackage .}})
{{- end}}
{{- end}}
//
// Example:
//
//	matcher := factory.{{.TypeName}}Matches().
{{- if .GetterMatchers}}
{{- with index .GetterMatchers 0}}
//	    {{.Name}}(testgen.Equal({{if eq .ReturnType "string"}}"expected_value"{{else if eq .ReturnType "int"}}100{{else if eq .ReturnType "bool"}}true{{else}}expectedValue{{end}})).
{{- end}}
{{- if gt (len .GetterMatchers) 1}}
{{- with index .GetterMatchers 1}}
//	    {{.Name}}(testgen.{{if eq .ReturnType "string"}}Contains("substring"){{else if eq .ReturnType "int"}}GreaterThan(0){{else}}Equal(expectedValue){{end}}).
{{- end}}
{{- end}}
{{- else if .Fields}}
{{- with index .Fields 0}}
//	    {{.Name}}(testgen.Equal({{if eq .TypeExpr "string"}}"expected_value"{{else if eq .TypeExpr "int"}}100{{else if eq .TypeExpr "bool"}}true{{else}}expectedValue{{end}})).
{{- end}}
{{- if gt (len .Fields) 1}}
{{- with index .Fields 1}}
//	    {{.Name}}(testgen.{{if eq .TypeExpr "string"}}Contains("substring"){{else if eq .TypeExpr "int"}}GreaterThan(0){{else}}Equal(expectedValue){{end}}).
{{- end}}
{{- end}}
{{- end}}
//	    Matcher()
//
//	testgen.AssertThat(t, actual{{.TypeName}}, matcher)
func {{.TypeName}}Matches() {{.TypeName}}Matcher {
	return {{.TypeName}}Matcher{}
}

{{range .Fields}}
// {{.Name}} adds a matcher for the {{.Name}} field.
func (m {{$.TypeName}}Matcher) {{.Name}}(matcher testgen.Matcher[{{qualifiedType $.ParentPackage .}}]) {{$.TypeName}}Matcher {
	m.{{lower .Name}}Matcher = matcher
	return m
}
{{if and .IsCustomType (not (hasPrefix .TypeExpr "[]")) (not (hasPrefix .TypeExpr "*"))}}
// {{.Name}}Matches is a convenience method that accepts a {{.UnqualifiedTypeName}}Matcher.
func (m {{$.TypeName}}Matcher) {{.Name}}Matches(matcher {{.UnqualifiedTypeName}}Matcher) {{$.TypeName}}Matcher {
	m.{{lower .Name}}Matcher = matcher.Matcher()
	return m
}
{{end}}
{{end}}

{{range .GetterMatchers}}
// {{.Name}} adds a matcher for the {{.Name}} property.
// This property is accessed via the {{.Getter}}() method.
func (m {{$.TypeName}}Matcher) {{.Name}}(matcher testgen.Matcher[{{qualifiedReturnType $.ParentPackage .ReturnType}}]) {{$.TypeName}}Matcher {
	m.{{lower .Name}}Matcher = matcher
	return m
}
{{end}}

// Matcher returns the composed matcher for {{.TypeName}}.
func (m {{.TypeName}}Matcher) Matcher() testgen.Matcher[{{.ParentPackage}}.{{.TypeName}}] {
	return testgen.MatcherFunc[{{.ParentPackage}}.{{.TypeName}}](func(actual {{.ParentPackage}}.{{.TypeName}}) testgen.MatchResult {
		// Extract all field values upfront (call each getter exactly once)
		{{- range .Fields}}
		{{lower .Name}}Value := actual.{{.Name}}
		{{- end}}
		{{- range .GetterMatchers}}
		{{lower .Name}}Value := actual.{{.Getter}}()
		{{- end}}

		// Build fieldValues map for structured diff
		fieldValues := map[string]any{
			{{- range .Fields}}
			"{{.Name}}": {{lower .Name}}Value,
			{{- end}}
			{{- range .GetterMatchers}}
			"{{.Name}}": {{lower .Name}}Value,
			{{- end}}
		}

		// Check matchers using cached values and store results
		fieldResults := make(map[string]*testgen.MatchResult)
		hasFailures := false
		{{- range .Fields}}

		if m.{{lower .Name}}Matcher != nil {
			result := m.{{lower .Name}}Matcher.Matches({{lower .Name}}Value)
			fieldResults["{{.Name}}"] = &result
			if !result.Matched {
				hasFailures = true
			}
		}
		{{- end}}
		{{- range .GetterMatchers}}

		if m.{{lower .Name}}Matcher != nil {
			result := m.{{lower .Name}}Matcher.Matches({{lower .Name}}Value)
			fieldResults["{{.Name}}"] = &result
			if !result.Matched {
				hasFailures = true
			}
		}
		{{- end}}

		if hasFailures {
			// Use structured diff for struct types
			structDiff := testgen.BuildMatcherStructDiff("{{.TypeName}}", fieldValues, fieldResults)
			return testgen.MatchResult{
				Matched: false,
				Message: structDiff,
			}
		}
		return testgen.MatchResult{Matched: true}
	})
}
`))

func defaultProvider(pkg string, f field) string {
	name := f.Name
	typ := f.TypeExpr

	// Handle pointers separately
	if strings.HasPrefix(typ, "*") {
		elemType := strings.TrimPrefix(typ, "*")
		// Check if elem is primitive
		isPrimitive := isPrimitiveType(elemType)

		if isPrimitive {
			return fmt.Sprintf("func(p testgen.Primitives) %s { var zero %s; return zero }", typ, typ)
		}
		// Custom type pointer - use PtrOf with FromSpec, use unqualified name for function references
		return fmt.Sprintf("testgen.PtrOf(testgen.FromSpec(Build%s, New%sSpec))", f.UnqualifiedTypeName, f.UnqualifiedTypeName)
	}

	// Handle slices separately
	if strings.HasPrefix(typ, "[]") {
		elemType := strings.TrimPrefix(typ, "[]")
		if isPrimitiveType(elemType) {
			return fmt.Sprintf("func(p testgen.Primitives) %s { var zero %s; return zero }", typ, typ)
		}
		// Custom type slice - type is already qualified from collectFields
		return fmt.Sprintf("func(p testgen.Primitives) %s { var zero %s; return zero }", typ, typ)
	}

	// If it's a custom type, use FromSpec with unqualified name for function references
	if f.IsCustomType {
		return fmt.Sprintf("testgen.FromSpec(Build%s, New%sSpec)", f.UnqualifiedTypeName, f.UnqualifiedTypeName)
	}

	switch typ {
	case "string":
		ln := strings.ToLower(name)
		if strings.HasSuffix(ln, "id") || ln == "id" {
			return "func(p testgen.Primitives) string { return p.ID() }"
		}
		return fmt.Sprintf("func(p testgen.Primitives) string { return p.StringWith(%q) }", strings.ToLower(name)+"_")
	case "int":
		return "func(p testgen.Primitives) int { return p.Int() }"
	case "int64":
		return "func(p testgen.Primitives) int64 { return p.Int64() }"
	case "uint64":
		return "func(p testgen.Primitives) uint64 { return p.Uint64() }"
	case "bool":
		return "func(p testgen.Primitives) bool { return p.Bool() }"
	case "float64":
		return "func(p testgen.Primitives) float64 { return p.Float64() }"
	case "time.Time":
		return "func(p testgen.Primitives) time.Time { return p.Time() }"
	case "time.Duration":
		return "func(p testgen.Primitives) time.Duration { return p.Duration() }"
	default:
		// conservative: default to zero for unknown types in this skeleton
		return fmt.Sprintf("func(p testgen.Primitives) %s { var zero %s; return zero }", typ, typ)
	}
}
