# Conjecture: A Go Port of Hypothesis's Core Engine

## Overview

This package implements property-based testing for Go using the same architectural approach as Python's Hypothesis library. The key insight from Hypothesis is that **generation and shrinking operate on a unified choice sequence**, not on the generated values directly.

## Core Architecture

### The Choice Sequence Model

All generated values are produced by interpreting a sequence of typed **choices**. Each choice is one of five primitive types:

- `Boolean` - true/false with probability parameter
- `Integer` - bounded integers with shrink-towards target  
- `Float` - IEEE 754 floats with constraints
- `String` - unicode strings with codepoint intervals
- `Bytes` - raw byte sequences

**Key principle**: Smaller choice values always produce simpler output values. This is what makes shrinking work—we shrink the choice sequence, and simpler values emerge automatically.

### Simplicity Ordering (Shortlex)

Choice sequences are ordered by **shortlex**: shorter sequences are simpler; among equal-length sequences, lexicographically smaller is simpler (treating each choice as comparable within its type).

### Three-Layer Design

```
┌─────────────────────────────────────┐
│         User Test Code              │
│   Property functions, assertions    │
├─────────────────────────────────────┤
│         Generator Layer             │
│  Composable generators (Gen[T])     │
│  Build complex types from primitives│
├─────────────────────────────────────┤
│         Conjecture Engine           │
│  Choice sequence, shrinking passes  │
│  Test execution, database           │
└─────────────────────────────────────┘
```

## Package Structure

```
conjecture/
├── choice.go          # Choice types and sequence
├── data.go            # ConjectureData - generation context
├── provider.go        # PrimitiveProvider interface
├── gen.go             # Gen[T] type and combinators
├── primitives.go      # Primitive generators
├── shrink.go          # Shrinking passes
├── engine.go          # Test runner/orchestrator
├── database.go        # Example database
├── settings.go        # Configuration
└── testing.go         # testing.T integration
```

## Design Decisions for Go

### 1. Generics for Type Safety
Go 1.18+ generics allow us to have `Gen[T]` instead of returning `interface{}`.

### 2. Explicit Error Handling
Unlike Python's exceptions, we use explicit error returns and a `Result[T]` pattern for generation outcomes.

### 3. Context for Cancellation
Long-running shrinking can be cancelled via `context.Context`.

### 4. Functional Options for Configuration
Settings use the functional options pattern for clean API.

### 5. Interface-Based Extension
Custom generators and providers implement interfaces rather than inheriting.

## Shrinking Strategy

Shrinking operates on the choice sequence through multiple passes:

1. **Deletion passes** - Remove chunks of the sequence
2. **Zeroing passes** - Replace regions with zero-values
3. **Lexicographic reduction** - Binary search values toward targets
4. **Block operations** - Minimize intervals, redistribute values

The shrinker runs passes until no progress is made, always checking that the test still fails after each transformation.
