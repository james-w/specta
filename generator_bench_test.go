package specta

import (
	"testing"

	"github.com/james-w/specta/conjecture"
)

// BenchmarkIntGenerator measures Int() generation performance
func BenchmarkIntGenerator(b *testing.B) {
	b.Run("unconstrained", func(b *testing.B) {
		gen := Int()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("with_range", func(b *testing.B) {
		gen := Int().Range(0, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := Int().Range(0, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			_, _ = gen.Draw(data)
		}
	})
}

// BenchmarkStringGenerator measures String() generation performance
func BenchmarkStringGenerator(b *testing.B) {
	b.Run("unconstrained", func(b *testing.B) {
		gen := String()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("alphanumeric_short", func(b *testing.B) {
		gen := String().AlphaNum().MaxLen(20)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := String().AlphaNum().MaxLen(50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			_, _ = gen.Draw(data)
		}
	})
}

// BenchmarkSliceGenerator measures Slice() generation performance
func BenchmarkSliceGenerator(b *testing.B) {
	b.Run("int_slice_small", func(b *testing.B) {
		gen := Slice(Int()).MaxLen(10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("string_slice_small", func(b *testing.B) {
		gen := Slice(String().MaxLen(10)).MaxLen(5)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := Slice(Int()).MaxLen(20)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			_, _ = gen.Draw(data)
		}
	})
}

// BenchmarkMapGenerator measures Map() generation performance
func BenchmarkMapGenerator(b *testing.B) {
	b.Run("int_to_string_small", func(b *testing.B) {
		gen := MapOf(Int(), String().MaxLen(10)).MaxLen(10)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("random_mode", func(b *testing.B) {
		gen := MapOf(Int(), String().MaxLen(20)).MaxLen(15)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			_, _ = gen.Draw(data)
		}
	})
}

// BenchmarkPropertyTest measures full property test iteration overhead
func BenchmarkPropertyTest(b *testing.B) {
	b.Run("simple_int_property", func(b *testing.B) {
		gen := Int().Range(0, 100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			value, _ := gen.Draw(data)
			// Simulate simple assertion
			_ = value >= 0 && value <= 100
		}
	})

	b.Run("complex_property", func(b *testing.B) {
		sliceGen := Slice(Int().Range(0, 100)).MaxLen(20)
		strGen := String().AlphaNum().MaxLen(30)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			slice, _ := sliceGen.Draw(data)
			str, _ := strGen.Draw(data)
			// Simulate complex assertions
			_ = len(slice) <= 20
			_ = len(str) <= 30
		}
	})
}

// BenchmarkDrawBits measures the core drawing performance
func BenchmarkDrawBits(b *testing.B) {
	b.Run("draw_integer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = data.DrawInteger(conjecture.IntegerParams{Min: 0, Max: 100, ShrinkToward: 0})
		}
	})

	b.Run("draw_bytes", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(0))
			_, _ = data.DrawBytes(conjecture.BytesParams{MinSize: 32, MaxSize: 32})
		}
	})
}

// BenchmarkEdgeCaseBiasing measures biasing overhead
func BenchmarkEdgeCaseBiasing(b *testing.B) {
	b.Run("int_with_biasing", func(b *testing.B) {
		gen := Int().Range(0, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			_, _ = gen.Draw(data)
		}
	})

	b.Run("slice_with_biasing", func(b *testing.B) {
		gen := Slice(Int()).MaxLen(50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := conjecture.NewConjectureData(conjecture.WithSeed(uint64(i)))
			_, _ = gen.Draw(data)
		}
	})
}
