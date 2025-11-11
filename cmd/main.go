package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
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

type TypeConfig struct {
	Name        string         `yaml:"name"`
	Constructor string         `yaml:"constructor"`
	Matchers    []MatcherField `yaml:"matchers"`
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
	return &c, nil
}

type field struct {
	Name         string
	TypeExpr     string
	IsCustomType bool   // true if this is a local struct type (not primitive)
	RecipeName   string // e.g., "UserRecipe" if IsCustomType
}

type ConstructorParam struct {
	Name         string
	TypeExpr     string
	IsCustomType bool
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
}

func findTypeConfig(cfg *Config, typeName string) *TypeConfig {
	for i := range cfg.Types {
		if cfg.Types[i].Name == typeName {
			return &cfg.Types[i]
		}
	}
	return nil
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

		// Render the type expression
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, pkg.Fset, param.Type); err != nil {
			continue
		}
		typeExpr := buf.String()

		// Check if custom type
		isCustomType := typeNames[typeExpr]

		for _, name := range param.Names {
			params = append(params, ConstructorParam{
				Name:         capitalizeFirst(name.Name),
				TypeExpr:     typeExpr,
				IsCustomType: isCustomType,
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
func analyzeGetterMethod(fn *ast.FuncDecl, pkg *packages.Package) string {
	if fn.Type == nil || fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return ""
	}

	// Get first return type
	result := fn.Type.Results.List[0]
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, pkg.Fset, result.Type); err != nil {
		return ""
	}
	return buf.String()
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
	pkgCfg := &packages.Config{Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedFiles, Dir: "."}
	pkgs, err := packages.Load(pkgCfg, tgt.Package)
	if err != nil || packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("load: %v", err)
	}
	pkg := pkgs[0]

	// Collect all type names for detecting custom types
	typeNames := make(map[string]bool)
	for _, tn := range tgt.Types.Include {
		typeNames[tn] = true
	}

	for _, typeName := range tgt.Types.Include {
		st := findStruct(pkg.Syntax, typeName)
		if st == nil {
			log.Printf("skip %s: not found", typeName)
			continue
		}

		// Output to factory/ and factory/spec/ subdirectories
		factoryDir := filepath.Join(filepath.Dir(pkg.GoFiles[0]), "factory")
		specDir := filepath.Join(factoryDir, "spec")
		if err := os.MkdirAll(specDir, 0755); err != nil {
			return fmt.Errorf("mkdir factory/spec: %w", err)
		}

		d := data{
			Package:       "factory",
			ParentPackage: pkg.Name,
			ParentImport:  pkg.PkgPath,
			TypeName:      typeName,
			SpecName:      typeName + "Spec",
			RecipeName:    typeName + "Recipe",
			ImportTestgen: "github.com/james-w/specta",
			ConfigPath:    *cfgPath,
		}

		// Check if this type has a constructor configured
		typeCfg := findTypeConfig(cfg, typeName)
		if typeCfg != nil && typeCfg.Constructor != "" {
			// Use constructor-based generation
			ctor := findConstructor(pkg, typeCfg.Constructor)
			if ctor != nil {
				d.HasConstructor = true
				d.ConstructorName = typeCfg.Constructor
				d.ConstructorParams = analyzeConstructorParams(ctor, pkg, typeNames)
				d.ConstructorReturns = analyzeConstructorReturns(ctor, pkg)

				// Check if any params use time types
				for _, param := range d.ConstructorParams {
					if strings.Contains(param.TypeExpr, "time.") {
						d.ImportTime = true
						break
					}
				}
			} else {
				log.Printf("warning: constructor %s not found for %s, using field-based generation", typeCfg.Constructor, typeName)
			}
		}

		// If no constructor, use field-based generation
		if !d.HasConstructor {
			fields := collectFields(pkg, st, typeNames)
			d.Fields = fields
			d.ImportTime = anyHas(fields, "time.Time") || anyHas(fields, "time.Duration")
		}

		// Analyze configured getter matchers
		if typeCfg != nil && len(typeCfg.Matchers) > 0 {
			for _, mf := range typeCfg.Matchers {
				getter := findGetterMethod(pkg, typeName, mf.Getter)
				if getter == nil {
					log.Printf("warning: getter method %s not found for %s", mf.Getter, typeName)
					continue
				}
				returnType := analyzeGetterMethod(getter, pkg)
				if returnType == "" {
					log.Printf("warning: could not determine return type for %s.%s", typeName, mf.Getter)
					continue
				}
				d.GetterMatchers = append(d.GetterMatchers, GetterInfo{
					Name:       mf.Name,
					Getter:     mf.Getter,
					ReturnType: returnType,
				})
			}
		}

		// Generate spec file (low-level API)
		specOut := filepath.Join(specDir, strings.ToLower(typeName)+"_gen.go")
		if err := renderSpec(specOut, d); err != nil {
			return err
		}
		log.Printf("wrote %s", specOut)

		// Generate recipe file (high-level API)
		recipeOut := filepath.Join(factoryDir, strings.ToLower(typeName)+"_gen.go")
		if err := renderRecipe(recipeOut, d); err != nil {
			return err
		}
		log.Printf("wrote %s", recipeOut)

		// Generate matcher file
		matcherOut := filepath.Join(factoryDir, strings.ToLower(typeName)+"_matcher_gen.go")
		if err := renderMatcher(matcherOut, d); err != nil {
			return err
		}
		log.Printf("wrote %s", matcherOut)
	}
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

func collectFields(pkg *packages.Package, st *ast.StructType, typeNames map[string]bool) []field {
	var out []field
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			continue // skip embedded/anon for now
		}
		name := f.Names[0].Name

		// Render ONLY the type node. This never includes tags.
		var buf bytes.Buffer
		// Either printer.Fprint or format.Node works; printer is fine here.
		if err := printer.Fprint(&buf, pkg.Fset, f.Type); err != nil {
			// fallback: extremely conservative
			buf.WriteString("interface{}")
		}
		typ := buf.String()

		// Just in case: if weird spacing left a tag tail, drop anything after a backtick.
		if i := strings.IndexByte(typ, '`'); i >= 0 {
			typ = strings.TrimSpace(typ[:i])
		}

		// Detect if this is a custom struct type (not slices)
		isCustomType := typeNames[typ]
		recipeName := ""
		if isCustomType {
			recipeName = typ + "Recipe"
		}

		out = append(out, field{
			Name:         name,
			TypeExpr:     typ,
			IsCustomType: isCustomType,
			RecipeName:   recipeName,
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

// verifyCompiles checks that the generated code type-checks without writing it to disk
func verifyCompiles(src []byte, filename string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}

	// Use packages.Load with an overlay to type-check without writing
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
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

func renderSpec(out string, d data) error {
	os.MkdirAll(filepath.Dir(out), 0o755)
	var buf bytes.Buffer
	if err := specTmpl.Execute(&buf, d); err != nil {
		return err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(out+".broken", buf.Bytes(), 0644)
		return fmt.Errorf("format: %v (wrote %s.broken)", err, out)
	}

	// Verify it compiles before writing
	if err := verifyCompiles(src, out); err != nil {
		_ = os.WriteFile(out+".broken", src, 0644)
		return fmt.Errorf("type-check failed: %v (wrote %s.broken)", err, out)
	}

	return os.WriteFile(out, src, 0644)
}

func renderRecipe(out string, d data) error {
	os.MkdirAll(filepath.Dir(out), 0o755)
	var buf bytes.Buffer
	if err := recipeTmpl.Execute(&buf, d); err != nil {
		return err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(out+".broken", buf.Bytes(), 0644)
		return fmt.Errorf("format: %v (wrote %s.broken)", err, out)
	}

	// Verify it compiles before writing
	if err := verifyCompiles(src, out); err != nil {
		_ = os.WriteFile(out+".broken", src, 0644)
		return fmt.Errorf("type-check failed: %v (wrote %s.broken)", err, out)
	}

	return os.WriteFile(out, src, 0644)
}

func renderMatcher(out string, d data) error {
	os.MkdirAll(filepath.Dir(out), 0o755)
	var buf bytes.Buffer
	if err := matcherTmpl.Execute(&buf, d); err != nil {
		return err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(out+".broken", buf.Bytes(), 0644)
		return fmt.Errorf("format: %v (wrote %s.broken)", err, out)
	}

	// Verify it compiles before writing
	if err := verifyCompiles(src, out); err != nil {
		_ = os.WriteFile(out+".broken", src, 0644)
		return fmt.Errorf("type-check failed: %v (wrote %s.broken)", err, out)
	}

	return os.WriteFile(out, src, 0644)
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
		// Handle slice of custom type
		if strings.HasPrefix(f.TypeExpr, "[]") {
			elemType := strings.TrimPrefix(f.TypeExpr, "[]")
			// Check if this is a known primitive type
			isPrimitive := elemType == "string" || elemType == "int" || elemType == "int64" ||
				elemType == "uint64" || elemType == "bool" || elemType == "float64" ||
				elemType == "time.Time" || elemType == "time.Duration"
			if !isPrimitive {
				return "[]" + pkg + "." + elemType
			}
		}
		// Handle custom type
		if f.IsCustomType && !strings.HasPrefix(f.TypeExpr, "[]") {
			return pkg + "." + f.TypeExpr
		}
		return f.TypeExpr
	},
	"qualifiedTypeParam": func(pkg string, p ConstructorParam) string {
		// Handle slice of custom type
		if strings.HasPrefix(p.TypeExpr, "[]") {
			elemType := strings.TrimPrefix(p.TypeExpr, "[]")
			isPrimitive := elemType == "string" || elemType == "int" || elemType == "int64" ||
				elemType == "uint64" || elemType == "bool" || elemType == "float64" ||
				elemType == "time.Time" || elemType == "time.Duration"
			if !isPrimitive {
				return "[]" + pkg + "." + elemType
			}
		}
		// Handle custom type
		if p.IsCustomType && !strings.HasPrefix(p.TypeExpr, "[]") {
			return pkg + "." + p.TypeExpr
		}
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
				return "testgen.FromSpec(Build" + p.TypeExpr + ", New" + p.TypeExpr + "Spec)"
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
	{{$.TypeName}}Default{{.Name}} = {{ defaultProviderParam $.ParentPackage . }}
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
	"hasPrefix": strings.HasPrefix,
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
		// Handle slice of custom type
		if strings.HasPrefix(f.TypeExpr, "[]") {
			elemType := strings.TrimPrefix(f.TypeExpr, "[]")
			// Check if this is a known primitive type
			isPrimitive := elemType == "string" || elemType == "int" || elemType == "int64" ||
				elemType == "uint64" || elemType == "bool" || elemType == "float64" ||
				elemType == "time.Time" || elemType == "time.Duration"
			if !isPrimitive {
				return "[]" + pkg + "." + elemType
			}
		}
		// Handle custom type
		if f.IsCustomType && !strings.HasPrefix(f.TypeExpr, "[]") {
			return pkg + "." + f.TypeExpr
		}
		return f.TypeExpr
	},
	"qualifiedTypeParam": func(pkg string, p ConstructorParam) string {
		// Handle slice of custom type
		if strings.HasPrefix(p.TypeExpr, "[]") {
			elemType := strings.TrimPrefix(p.TypeExpr, "[]")
			isPrimitive := elemType == "string" || elemType == "int" || elemType == "int64" ||
				elemType == "uint64" || elemType == "bool" || elemType == "float64" ||
				elemType == "time.Time" || elemType == "time.Duration"
			if !isPrimitive {
				return "[]" + pkg + "." + elemType
			}
		}
		// Handle custom type
		if p.IsCustomType && !strings.HasPrefix(p.TypeExpr, "[]") {
			return pkg + "." + p.TypeExpr
		}
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
	{{- if not .HasConstructor}}
	{{- range .Fields}}
	{{- if .IsCustomType}}
	{{- if not (hasPrefix .TypeExpr "[]")}}
	{{lower .Name}}Recipe *{{.TypeExpr}}Recipe
	{{- end}}
	{{- end}}
	{{- end}}
	{{- end}}
}

// {{.TypeName}} creates a new {{.RecipeName}} for building {{.TypeName}} instances.
func {{.TypeName}}() {{.RecipeName}} { return {{.RecipeName}}{} }

{{- if .HasConstructor}}
{{range .ConstructorParams}}
// {{.Name}} sets the {{.Name}} parameter.
func (r {{$.RecipeName}}) {{.Name}}(v {{qualifiedTypeParam $.ParentPackage .}}) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}(v))
	return r
}
{{if .IsCustomType}}
{{- if not (hasPrefix .TypeExpr "[]")}}
// {{.Name}}FromRecipe sets the {{.Name}} parameter using another Recipe (creates unique instances).
func (r {{$.RecipeName}}) {{.Name}}FromRecipe(v {{.TypeExpr}}Recipe) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(v.Provider()))
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
func (r {{$.RecipeName}}) {{.Name}}FromRecipe(v {{.TypeExpr}}Recipe) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(v.Provider()))
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
	{{- if .HasConstructor}}
	// Constructor-based types: matchers are not yet fully supported, using deep equality
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
			m = m.{{.Name}}(r.{{lower .Name}}Recipe.AsEqualMatcher())
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
	"lower": func(s string) string {
		if s == "" {
			return s
		}
		r := []rune(s)
		r[0] = []rune(strings.ToLower(string(r[0])))[0]
		return string(r)
	},
	"qualifiedType": func(pkg string, f field) string {
		// Handle slice of custom type
		if strings.HasPrefix(f.TypeExpr, "[]") {
			elemType := strings.TrimPrefix(f.TypeExpr, "[]")
			// Check if this is a known primitive type
			isPrimitive := elemType == "string" || elemType == "int" || elemType == "int64" ||
				elemType == "uint64" || elemType == "bool" || elemType == "float64" ||
				elemType == "time.Time" || elemType == "time.Duration"
			if !isPrimitive {
				return "[]" + pkg + "." + elemType
			}
		}
		// Handle custom type
		if f.IsCustomType && !strings.HasPrefix(f.TypeExpr, "[]") {
			return pkg + "." + f.TypeExpr
		}
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
	{{lower .Name}}Matcher testgen.Matcher[{{.ReturnType}}]
	{{- end}}
}

// {{.TypeName}}Matches creates a new {{.TypeName}}Matcher for matching {{.TypeName}} instances.
func {{.TypeName}}Matches() {{.TypeName}}Matcher {
	return {{.TypeName}}Matcher{}
}

{{range .Fields}}
// {{.Name}} adds a matcher for the {{.Name}} field.
func (m {{$.TypeName}}Matcher) {{.Name}}(matcher testgen.Matcher[{{qualifiedType $.ParentPackage .}}]) {{$.TypeName}}Matcher {
	m.{{lower .Name}}Matcher = matcher
	return m
}
{{if .IsCustomType}}
// {{.Name}}Matches is a convenience method that accepts a {{.TypeExpr}}Matcher.
func (m {{$.TypeName}}Matcher) {{.Name}}Matches(matcher {{.TypeExpr}}Matcher) {{$.TypeName}}Matcher {
	m.{{lower .Name}}Matcher = matcher.Matcher()
	return m
}
{{end}}
{{end}}

{{range .GetterMatchers}}
// {{.Name}} adds a matcher for the {{.Name}} property (via {{.Getter}}).
func (m {{$.TypeName}}Matcher) {{.Name}}(matcher testgen.Matcher[{{.ReturnType}}]) {{$.TypeName}}Matcher {
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

	// Handle slices separately - need to qualify custom types
	if strings.HasPrefix(typ, "[]") {
		elemType := strings.TrimPrefix(typ, "[]")
		// Check if elem is primitive
		isPrimitive := elemType == "string" || elemType == "int" || elemType == "int64" ||
			elemType == "uint64" || elemType == "bool" || elemType == "float64" ||
			elemType == "time.Time" || elemType == "time.Duration"

		if isPrimitive {
			return fmt.Sprintf("func(p testgen.Primitives) %s { var zero %s; return zero }", typ, typ)
		}
		// Custom type slice - qualify it
		qualifiedType := "[]" + pkg + "." + elemType
		return fmt.Sprintf("func(p testgen.Primitives) %s { var zero %s; return zero }", qualifiedType, qualifiedType)
	}

	// If it's a custom type, use FromSpec
	if f.IsCustomType {
		return fmt.Sprintf("testgen.FromSpec(Build%s, New%sSpec)", typ, typ)
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
