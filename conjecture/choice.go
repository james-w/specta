// Package conjecture implements property-based testing using Hypothesis's
// choice sequence approach to generation and shrinking.
package conjecture

import (
	"cmp"
	"fmt"
	"math"
)

// ChoiceType identifies the type of a choice in the sequence.
type ChoiceType uint8

const (
	ChoiceBoolean ChoiceType = iota
	ChoiceInteger
	ChoiceFloat
	ChoiceString
	ChoiceBytes
)

func (t ChoiceType) String() string {
	switch t {
	case ChoiceBoolean:
		return "boolean"
	case ChoiceInteger:
		return "integer"
	case ChoiceFloat:
		return "float"
	case ChoiceString:
		return "string"
	case ChoiceBytes:
		return "bytes"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// Choice represents a single choice in the choice sequence.
// Each choice has a type, value, and constraints that were used during generation.
type Choice struct {
	Type        ChoiceType
	Value       any         // bool, int64, float64, string, or []byte
	Constraints Constraints // Type-specific generation constraints
	Forced      bool        // Whether this value was forced (not random)
}

// Constraints holds type-specific parameters for a choice.
type Constraints struct {
	// Boolean constraints
	Probability float64 // P(true), default 0.5

	// Integer constraints
	MinInt       int64
	MaxInt       int64
	ShrinkToward int64 // Value to shrink toward (usually 0)

	// Float constraints
	MinFloat       float64
	MaxFloat       float64
	AllowNaN       bool
	AllowInfinity  bool
	SmallestNormal float64 // Smallest non-subnormal

	// String constraints
	MinSize   int
	MaxSize   int // -1 for unbounded
	Intervals []CodepointInterval

	// Bytes constraints (uses MinSize/MaxSize)
}

// CodepointInterval represents a range of allowed Unicode codepoints.
type CodepointInterval struct {
	Low  rune
	High rune
}

// ChoiceSequence is an ordered sequence of choices that can be replayed
// or shrunk to produce different test inputs.
type ChoiceSequence struct {
	choices []Choice
	spans   []Span // Hierarchical structure for shrinking
}

// Span marks a semantically meaningful region in the choice sequence.
// Spans can be nested and are used for targeted shrinking.
type Span struct {
	Start   int    // Index of first choice in span
	End     int    // Index after last choice in span
	Label   string // Identifies what this span represents
	Discard bool   // Whether this span can be discarded (e.g., filtered draws)
}

// NewChoiceSequence creates an empty choice sequence.
func NewChoiceSequence() *ChoiceSequence {
	return &ChoiceSequence{
		choices: make([]Choice, 0, 64),
		spans:   make([]Span, 0, 16),
	}
}

// Len returns the number of choices in the sequence.
func (cs *ChoiceSequence) Len() int {
	return len(cs.choices)
}

// Get returns the choice at index i.
func (cs *ChoiceSequence) Get(i int) Choice {
	return cs.choices[i]
}

// Append adds a choice to the sequence.
func (cs *ChoiceSequence) Append(c Choice) {
	cs.choices = append(cs.choices, c)
}

// Choices returns a copy of all choices.
func (cs *ChoiceSequence) Choices() []Choice {
	result := make([]Choice, len(cs.choices))
	copy(result, cs.choices)
	return result
}

// StartSpan begins a new span at the current position.
func (cs *ChoiceSequence) StartSpan(label string) int {
	spanID := len(cs.spans)
	cs.spans = append(cs.spans, Span{
		Start: len(cs.choices),
		End:   -1, // Not yet closed
		Label: label,
	})
	return spanID
}

// EndSpan closes the span with the given ID.
func (cs *ChoiceSequence) EndSpan(spanID int, discard bool) {
	if spanID >= 0 && spanID < len(cs.spans) {
		cs.spans[spanID].End = len(cs.choices)
		cs.spans[spanID].Discard = discard
	}
}

// Spans returns a copy of all spans.
func (cs *ChoiceSequence) Spans() []Span {
	result := make([]Span, len(cs.spans))
	copy(result, cs.spans)
	return result
}

// Clone creates a deep copy of the choice sequence.
func (cs *ChoiceSequence) Clone() *ChoiceSequence {
	clone := &ChoiceSequence{
		choices: make([]Choice, len(cs.choices)),
		spans:   make([]Span, len(cs.spans)),
	}
	copy(clone.choices, cs.choices)
	copy(clone.spans, cs.spans)
	return clone
}

// Compare returns -1, 0, or 1 comparing two choice sequences using shortlex ordering.
// Shorter sequences are smaller; equal-length sequences are compared lexicographically.
func (cs *ChoiceSequence) Compare(other *ChoiceSequence) int {
	// Shortlex: shorter is smaller
	if len(cs.choices) != len(other.choices) {
		return cmp.Compare(len(cs.choices), len(other.choices))
	}

	// Same length: lexicographic comparison
	for i := range cs.choices {
		if c := compareChoices(cs.choices[i], other.choices[i]); c != 0 {
			return c
		}
	}
	return 0
}

// compareChoices compares two choices of the same type.
// Returns -1, 0, or 1.
func compareChoices(a, b Choice) int {
	if a.Type != b.Type {
		// Different types shouldn't happen in valid sequences at same position
		return cmp.Compare(a.Type, b.Type)
	}

	switch a.Type {
	case ChoiceBoolean:
		// false < true
		av, bv := a.Value.(bool), b.Value.(bool)
		if av == bv {
			return 0
		}
		if !av && bv {
			return -1
		}
		return 1

	case ChoiceInteger:
		av, bv := a.Value.(int64), b.Value.(int64)
		// Compare by distance from shrink target
		aDist := intDistance(av, a.Constraints.ShrinkToward)
		bDist := intDistance(bv, b.Constraints.ShrinkToward)
		if c := cmp.Compare(aDist, bDist); c != 0 {
			return c
		}
		// Same distance: prefer positive over negative
		return cmp.Compare(av, bv)

	case ChoiceFloat:
		av, bv := a.Value.(float64), b.Value.(float64)
		return compareFloats(av, bv)

	case ChoiceString:
		av, bv := a.Value.(string), b.Value.(string)
		// Shorter strings are simpler
		if c := cmp.Compare(len(av), len(bv)); c != 0 {
			return c
		}
		// Same length: lexicographic by codepoint
		return cmp.Compare(av, bv)

	case ChoiceBytes:
		av, bv := a.Value.([]byte), b.Value.([]byte)
		if c := cmp.Compare(len(av), len(bv)); c != 0 {
			return c
		}
		for i := range av {
			if c := cmp.Compare(av[i], bv[i]); c != 0 {
				return c
			}
		}
		return 0
	}

	return 0
}

// intDistance returns the absolute distance between two integers,
// handling overflow.
func intDistance(a, target int64) uint64 {
	if a >= target {
		return uint64(a - target)
	}
	return uint64(target - a)
}

// compareFloats compares floats for shrinking purposes.
// Order: 0 < positive < negative < larger magnitude < special values
func compareFloats(a, b float64) int {
	// Handle special cases
	aNaN, bNaN := math.IsNaN(a), math.IsNaN(b)
	aInf, bInf := math.IsInf(a, 0), math.IsInf(b, 0)

	// NaN is largest (least simple)
	if aNaN && !bNaN {
		return 1
	}
	if !aNaN && bNaN {
		return -1
	}
	if aNaN && bNaN {
		return 0
	}

	// Infinity is next largest
	if aInf && !bInf {
		return 1
	}
	if !aInf && bInf {
		return -1
	}
	if aInf && bInf {
		return cmp.Compare(a, b) // -Inf < +Inf
	}

	// For normal values: smaller absolute value is simpler
	aAbs, bAbs := math.Abs(a), math.Abs(b)
	if c := cmp.Compare(aAbs, bAbs); c != 0 {
		return c
	}

	// Same magnitude: positive is simpler than negative
	if a >= 0 && b < 0 {
		return -1
	}
	if a < 0 && b >= 0 {
		return 1
	}

	return 0
}

// DeleteRange returns a new sequence with choices [start, end) removed.
func (cs *ChoiceSequence) DeleteRange(start, end int) *ChoiceSequence {
	if start < 0 || end > len(cs.choices) || start >= end {
		return cs.Clone()
	}

	newChoices := make([]Choice, 0, len(cs.choices)-(end-start))
	newChoices = append(newChoices, cs.choices[:start]...)
	newChoices = append(newChoices, cs.choices[end:]...)

	// Adjust spans
	delta := end - start
	newSpans := make([]Span, 0, len(cs.spans))
	for _, s := range cs.spans {
		ns := s
		// Span entirely before deletion
		if s.End <= start {
			newSpans = append(newSpans, ns)
			continue
		}
		// Span entirely after deletion
		if s.Start >= end {
			ns.Start -= delta
			ns.End -= delta
			newSpans = append(newSpans, ns)
			continue
		}
		// Span overlaps deletion - adjust or skip
		if s.Start < start {
			ns.End = start
			if ns.End > ns.Start {
				newSpans = append(newSpans, ns)
			}
		}
		// Could also handle spans that straddle the deletion more carefully
	}

	return &ChoiceSequence{
		choices: newChoices,
		spans:   newSpans,
	}
}

// ReplaceChoice returns a new sequence with the choice at index i replaced.
func (cs *ChoiceSequence) ReplaceChoice(i int, c Choice) *ChoiceSequence {
	if i < 0 || i >= len(cs.choices) {
		return cs.Clone()
	}

	clone := cs.Clone()
	clone.choices[i] = c
	return clone
}
