package specta

import (
	"testing"
)

// BenchmarkIntGenerator measures Int() generation performance
func BenchmarkIntGenerator(b *testing.B) {
	b.Run("unconstrained", func(b *testing.B) {
		p := New()
		gen := Int()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("with_range", func(b *testing.B) {
		p := New()
		gen := Int().Range(0, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := Int().Range(0, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			gen.Draw(source, "value")
		}
	})
}

// BenchmarkStringGenerator measures String() generation performance
func BenchmarkStringGenerator(b *testing.B) {
	b.Run("unconstrained", func(b *testing.B) {
		p := New()
		gen := String()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("alphanumeric_short", func(b *testing.B) {
		p := New()
		gen := String().AlphaNum().MaxLen(20)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := String().AlphaNum().MaxLen(50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			gen.Draw(source, "value")
		}
	})
}

// BenchmarkSliceGenerator measures Slice() generation performance
func BenchmarkSliceGenerator(b *testing.B) {
	b.Run("int_slice_small", func(b *testing.B) {
		p := New()
		gen := Slice(Int()).MaxLen(10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("string_slice_small", func(b *testing.B) {
		p := New()
		gen := Slice(String().MaxLen(10)).MaxLen(5)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := Slice(Int()).MaxLen(20)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			gen.Draw(source, "value")
		}
	})
}

// BenchmarkMapGenerator measures Map() generation performance
func BenchmarkMapGenerator(b *testing.B) {
	b.Run("int_to_string_small", func(b *testing.B) {
		p := New()
		gen := Map(Int(), String().MaxLen(10)).MaxLen(10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen.Draw(p, "value")
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := Map(Int(), String().MaxLen(20)).MaxLen(15)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			gen.Draw(source, "value")
		}
	})
}

// BenchmarkPropertyTest measures full property test iteration overhead
func BenchmarkPropertyTest(b *testing.B) {
	b.Run("simple_int_property", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Run a single property test iteration
			source := NewSource(int64(i))
			gen := Int().Range(0, 100)
			value := gen.Draw(source, "x")
			// Simulate simple assertion
			_ = value >= 0 && value <= 100
		}
	})

	b.Run("complex_property", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			slice := Slice(Int().Range(0, 100)).MaxLen(20).Draw(source, "slice")
			str := String().AlphaNum().MaxLen(30).Draw(source, "str")
			// Simulate complex assertions
			_ = len(slice) <= 20
			_ = len(str) <= 30
		}
	})
}

// BenchmarkDrawBits measures the core DrawBits performance
func BenchmarkDrawBits(b *testing.B) {
	b.Run("deterministic", func(b *testing.B) {
		p := New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.DrawBits(32)
		}
	})

	b.Run("random", func(b *testing.B) {
		source := NewSource(12345)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source.DrawBits(32)
		}
	})
}

// BenchmarkEdgeCaseBiasing measures biasing overhead
func BenchmarkEdgeCaseBiasing(b *testing.B) {
	b.Run("int_with_biasing", func(b *testing.B) {
		gen := Int().Range(0, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			gen.Draw(source, "value")
		}
	})

	b.Run("slice_with_biasing", func(b *testing.B) {
		gen := Slice(Int()).MaxLen(50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			source := NewSource(int64(i))
			gen.Draw(source, "value")
		}
	})
}
