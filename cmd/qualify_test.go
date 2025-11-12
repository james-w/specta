package main

import (
	"testing"
)

func TestQualifyReturnType(t *testing.T) {
	tests := []struct {
		name          string
		returnType    string
		parentPackage string
		want          string
	}{
		// Primitive types - should not be qualified
		{
			name:          "string primitive",
			returnType:    "string",
			parentPackage: "showcase",
			want:          "string",
		},
		{
			name:          "int primitive",
			returnType:    "int",
			parentPackage: "showcase",
			want:          "int",
		},
		{
			name:          "int8 primitive",
			returnType:    "int8",
			parentPackage: "showcase",
			want:          "int8",
		},
		{
			name:          "int16 primitive",
			returnType:    "int16",
			parentPackage: "showcase",
			want:          "int16",
		},
		{
			name:          "int32 primitive",
			returnType:    "int32",
			parentPackage: "showcase",
			want:          "int32",
		},
		{
			name:          "int64 primitive",
			returnType:    "int64",
			parentPackage: "showcase",
			want:          "int64",
		},
		{
			name:          "uint primitive",
			returnType:    "uint",
			parentPackage: "showcase",
			want:          "uint",
		},
		{
			name:          "uint8 primitive",
			returnType:    "uint8",
			parentPackage: "showcase",
			want:          "uint8",
		},
		{
			name:          "uint16 primitive",
			returnType:    "uint16",
			parentPackage: "showcase",
			want:          "uint16",
		},
		{
			name:          "uint32 primitive",
			returnType:    "uint32",
			parentPackage: "showcase",
			want:          "uint32",
		},
		{
			name:          "uint64 primitive",
			returnType:    "uint64",
			parentPackage: "showcase",
			want:          "uint64",
		},
		{
			name:          "uintptr primitive",
			returnType:    "uintptr",
			parentPackage: "showcase",
			want:          "uintptr",
		},
		{
			name:          "byte primitive",
			returnType:    "byte",
			parentPackage: "showcase",
			want:          "byte",
		},
		{
			name:          "rune primitive",
			returnType:    "rune",
			parentPackage: "showcase",
			want:          "rune",
		},
		{
			name:          "float32 primitive",
			returnType:    "float32",
			parentPackage: "showcase",
			want:          "float32",
		},
		{
			name:          "float64 primitive",
			returnType:    "float64",
			parentPackage: "showcase",
			want:          "float64",
		},
		{
			name:          "complex64 primitive",
			returnType:    "complex64",
			parentPackage: "showcase",
			want:          "complex64",
		},
		{
			name:          "complex128 primitive",
			returnType:    "complex128",
			parentPackage: "showcase",
			want:          "complex128",
		},
		{
			name:          "bool primitive",
			returnType:    "bool",
			parentPackage: "showcase",
			want:          "bool",
		},
		{
			name:          "error primitive",
			returnType:    "error",
			parentPackage: "showcase",
			want:          "error",
		},
		{
			name:          "time.Time",
			returnType:    "time.Time",
			parentPackage: "showcase",
			want:          "time.Time",
		},
		{
			name:          "time.Duration",
			returnType:    "time.Duration",
			parentPackage: "showcase",
			want:          "time.Duration",
		},

		// Custom types - should be qualified
		{
			name:          "custom type",
			returnType:    "User",
			parentPackage: "showcase",
			want:          "showcase.User",
		},
		{
			name:          "already qualified custom type",
			returnType:    "showcase.User",
			parentPackage: "showcase",
			want:          "showcase.User",
		},
		{
			name:          "custom type from different package",
			returnType:    "other.Type",
			parentPackage: "showcase",
			want:          "other.Type",
		},

		// Pointer types
		{
			name:          "pointer to custom type",
			returnType:    "*User",
			parentPackage: "showcase",
			want:          "*showcase.User",
		},
		{
			name:          "pointer to primitive",
			returnType:    "*string",
			parentPackage: "showcase",
			want:          "*string",
		},
		{
			name:          "pointer to already qualified type",
			returnType:    "*showcase.User",
			parentPackage: "showcase",
			want:          "*showcase.User",
		},

		// Slice types
		{
			name:          "slice of custom type",
			returnType:    "[]User",
			parentPackage: "showcase",
			want:          "[]showcase.User",
		},
		{
			name:          "slice of primitive",
			returnType:    "[]string",
			parentPackage: "showcase",
			want:          "[]string",
		},
		{
			name:          "slice of pointers to custom type",
			returnType:    "[]*User",
			parentPackage: "showcase",
			want:          "[]*showcase.User",
		},

		// Map types
		{
			name:          "map with custom value type",
			returnType:    "map[string]User",
			parentPackage: "showcase",
			want:          "map[string]showcase.User",
		},
		{
			name:          "map with custom key type",
			returnType:    "map[User]string",
			parentPackage: "showcase",
			want:          "map[showcase.User]string",
		},
		{
			name:          "map with both custom types",
			returnType:    "map[User]Product",
			parentPackage: "showcase",
			want:          "map[showcase.User]showcase.Product",
		},
		{
			name:          "map with primitive types",
			returnType:    "map[string]int",
			parentPackage: "showcase",
			want:          "map[string]int",
		},
		{
			name:          "map with pointer value",
			returnType:    "map[string]*User",
			parentPackage: "showcase",
			want:          "map[string]*showcase.User",
		},
		{
			name:          "map with slice value",
			returnType:    "map[string][]User",
			parentPackage: "showcase",
			want:          "map[string][]showcase.User",
		},
		{
			name:          "nested map",
			returnType:    "map[string]map[int]User",
			parentPackage: "showcase",
			want:          "map[string]map[int]showcase.User",
		},

		// Channel types
		{
			name:          "bidirectional channel of custom type",
			returnType:    "chan User",
			parentPackage: "showcase",
			want:          "chan showcase.User",
		},
		{
			name:          "receive-only channel of custom type",
			returnType:    "<-chan User",
			parentPackage: "showcase",
			want:          "<-chan showcase.User",
		},
		{
			name:          "send-only channel of custom type",
			returnType:    "chan<- User",
			parentPackage: "showcase",
			want:          "chan<- showcase.User",
		},
		{
			name:          "channel of primitive",
			returnType:    "chan string",
			parentPackage: "showcase",
			want:          "chan string",
		},
		{
			name:          "channel of pointer to custom type",
			returnType:    "chan *User",
			parentPackage: "showcase",
			want:          "chan *showcase.User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := qualifyReturnType(tt.returnType, tt.parentPackage)
			if got != tt.want {
				t.Errorf("qualifyReturnType(%q, %q) = %q, want %q", tt.returnType, tt.parentPackage, got, tt.want)
			}
		})
	}
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
		{"*string", "*string", false},
		{"[]int", "[]int", false},
		{"map[string]int", "map[string]int", false},
		{"chan int", "chan int", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrimitiveType(tt.typeName)
			if got != tt.want {
				t.Errorf("isPrimitiveType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}
