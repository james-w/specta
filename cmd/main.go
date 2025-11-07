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

	"gopkg.in/yaml.v3"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	Version int `json:"version"`
	Targets []struct {
		Package    string `json:"package"`
		Types      struct{ Include []string `json:"include"` } `json:"types"`
		FileSuffix string `json:"file_suffix"`
	} `json:"targets"`
}

var cfgPath = flag.String("config", "testgen.yaml", "path to config (JSON for this skeleton)")

func main() {
	flag.Parse()
	cfg, err := loadConfig(*cfgPath)
	if err != nil { log.Fatal(err) }

	for _, t := range cfg.Targets {
		if err := processTarget(t); err != nil {
			log.Fatalf("target %s: %v", t.Package, err)
		}
	}
	log.Println("ok")
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil { return nil, err }
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
	IsCustomType bool  // true if this is a local struct type (not primitive)
	RecipeName   string // e.g., "UserRecipe" if IsCustomType
}

type data struct {
	Package        string // "factory"
	ParentPackage  string // e.g., "example"
	ParentImport   string // import path to parent package
	TypeName       string
	SpecName       string
	RecipeName     string
	Fields         []field
	ImportTime     bool
	ImportTestgen  string
}

func processTarget(tgt struct {
	Package    string `json:"package"`
	Types      struct{ Include []string `json:"include"` } `json:"types"`
	FileSuffix string `json:"file_suffix"`
}) error {
	if tgt.FileSuffix == "" { tgt.FileSuffix = "_testgen_gen.go" }
	cfg := &packages.Config{Mode: packages.NeedName|packages.NeedSyntax|packages.NeedTypes|packages.NeedFiles, Dir: "."}
	pkgs, err := packages.Load(cfg, tgt.Package)
	if err != nil || packages.PrintErrors(pkgs) > 0 { return fmt.Errorf("load: %v", err) }
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
		fields := collectFields(pkg, st, typeNames)

		// Output to factory/ and factory/spec/ subdirectories
		factoryDir := filepath.Join(filepath.Dir(pkg.GoFiles[0]), "factory")
		specDir := filepath.Join(factoryDir, "spec")
		if err := os.MkdirAll(specDir, 0755); err != nil {
			return fmt.Errorf("mkdir factory/spec: %w", err)
		}

		d := data{
			Package:        "factory",
			ParentPackage:  pkg.Name,
			ParentImport:   pkg.PkgPath,
			TypeName:       typeName,
			SpecName:       typeName + "Spec",
			RecipeName:     typeName + "Recipe",
			Fields:         fields,
			ImportTime:     anyHas(fields, "time.Time") || anyHas(fields, "time.Duration"),
			ImportTestgen:  "github.com/james-w/gomatchers",
		}

		// Generate spec file (low-level API)
		specOut := filepath.Join(specDir, strings.ToLower(typeName)+"_gen.go")
		if err := renderSpec(specOut, d); err != nil { return err }
		log.Printf("wrote %s", specOut)

		// Generate recipe file (high-level API)
		recipeOut := filepath.Join(factoryDir, strings.ToLower(typeName)+"_gen.go")
		if err := renderRecipe(recipeOut, d); err != nil { return err }
		log.Printf("wrote %s", recipeOut)
	}
	return nil
}

func findStruct(files []*ast.File, typeName string) *ast.StructType {
	var out *ast.StructType
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || ts.Name.Name != typeName { return true }
			if st, ok := ts.Type.(*ast.StructType); ok { out = st }
			return false
		})
		if out != nil { break }
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
	for _, f := range fields { if f.TypeExpr == typ { return true } }
	return false
}

func renderSpec(out string, d data) error {
	os.MkdirAll(filepath.Dir(out), 0o755)
	var buf bytes.Buffer
	if err := specTmpl.Execute(&buf, d); err != nil { return err }
	src, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(out+".broken", buf.Bytes(), 0644)
		return fmt.Errorf("format: %v (wrote %s.broken)", err, out)
	}
	return os.WriteFile(out, src, 0644)
}

func renderRecipe(out string, d data) error {
	os.MkdirAll(filepath.Dir(out), 0o755)
	var buf bytes.Buffer
	if err := recipeTmpl.Execute(&buf, d); err != nil { return err }
	src, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(out+".broken", buf.Bytes(), 0644)
		return fmt.Errorf("format: %v (wrote %s.broken)", err, out)
	}
	return os.WriteFile(out, src, 0644)
}

var specTmpl = template.Must(template.New("spec").Funcs(template.FuncMap{
	"lower": func(s string) string { if s=="" {return s}; r:=[]rune(s); r[0] = []rune(strings.ToLower(string(r[0])))[0]; return string(r) },
	"defaultProvider": func(pkg string, f field) string { return defaultProvider(pkg, f) },
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
	{{- range .Fields}}
	{{.Name}} testgen.Maybe[{{qualifiedType $.ParentPackage .}}]
	{{- end}}
}

// New{{.SpecName}} creates a new {{.SpecName}} with all fields unset.
func New{{.SpecName}}() {{.SpecName}} { return {{.SpecName}}{} }

// New{{.TypeName}}Factory creates a new SpecFactory for {{.TypeName}}.
func New{{.TypeName}}Factory(p testgen.Primitives) *testgen.SpecFactory[{{$.ParentPackage}}.{{.TypeName}}, {{.SpecName}}] {
	return testgen.NewSpecFactory(p, New{{.SpecName}}, Build{{.TypeName}})
}

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
`))

var recipeTmpl = template.Must(template.New("recipe").Funcs(template.FuncMap{
	"lower": func(s string) string { if s=="" {return s}; r:=[]rune(s); r[0] = []rune(strings.ToLower(string(r[0])))[0]; return string(r) },
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
	"{{.ParentImport}}/factory/spec"
)

// {{.RecipeName}} provides a fluent API for building {{.TypeName}} instances.
type {{.RecipeName}} struct{ opts []testgen.Opt[spec.{{.SpecName}}] }

// {{.TypeName}} creates a new {{.RecipeName}} for building {{.TypeName}} instances.
func {{.TypeName}}() {{.RecipeName}} { return {{.RecipeName}}{} }

{{range .Fields}}
// {{.Name}} sets the {{.Name}} field.
func (r {{$.RecipeName}}) {{.Name}}(v {{qualifiedType $.ParentPackage .}}) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}(v))
	return r
}
{{if .IsCustomType}}
// {{.Name}}FromRecipe sets the {{.Name}} field using another Recipe (creates unique instances).
func (r {{$.RecipeName}}) {{.Name}}FromRecipe(v {{.RecipeName}}) {{$.RecipeName}} {
	r.opts = append(r.opts, spec.With{{$.TypeName}}{{.Name}}FromProvider(v.Provider()))
	return r
}
{{end}}
{{end}}

// Provider returns a Provider for lazy evaluation in parent factories.
func (r {{.RecipeName}}) Provider() testgen.Provider[{{.ParentPackage}}.{{.TypeName}}] {
	return testgen.FromSpec(spec.Build{{.TypeName}}, spec.New{{.SpecName}}, r.opts...)
}

// Build creates a single {{.TypeName}} instance.
func (r {{.RecipeName}}) Build(p testgen.Primitives) {{.ParentPackage}}.{{.TypeName}} {
	return spec.New{{.TypeName}}Factory(p).Make(r.opts...)
}

// Many creates multiple {{.TypeName}} instances with unique generated values.
func (r {{.RecipeName}}) Many(n int, p testgen.Primitives) []{{.ParentPackage}}.{{.TypeName}} {
	return spec.New{{.TypeName}}Factory(p).Many(n, r.opts...)
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
		if strings.HasSuffix(ln, "id") || ln == "id" { return "func(p testgen.Primitives) string { return p.ID() }" }
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

