package conjecture

import (
	"testing"
)

func TestChoiceSequence_AppendAndGet(t *testing.T) {
	seq := NewChoiceSequence()

	c1 := Choice{Type: ChoiceInteger, Value: int64(42)}
	c2 := Choice{Type: ChoiceString, Value: "hello"}

	seq.Append(c1)
	seq.Append(c2)

	if seq.Len() != 2 {
		t.Errorf("expected length 2, got %d", seq.Len())
	}

	if got := seq.Get(0); got.Value != int64(42) {
		t.Errorf("expected first choice value 42, got %v", got.Value)
	}

	if got := seq.Get(1); got.Value != "hello" {
		t.Errorf("expected second choice value 'hello', got %v", got.Value)
	}
}

func TestChoiceSequence_Clone(t *testing.T) {
	original := NewChoiceSequence()
	original.Append(Choice{Type: ChoiceInteger, Value: int64(10)})

	clone := original.Clone()

	// Modify clone
	clone.Append(Choice{Type: ChoiceInteger, Value: int64(20)})

	if original.Len() != 1 {
		t.Errorf("original should have length 1, got %d", original.Len())
	}

	if clone.Len() != 2 {
		t.Errorf("clone should have length 2, got %d", clone.Len())
	}
}

func TestChoiceSequence_Compare(t *testing.T) {
	tests := []struct {
		name string
		seq1 *ChoiceSequence
		seq2 *ChoiceSequence
		want int // -1, 0, or 1
	}{
		{
			name: "shorter is less",
			seq1: seqWith(Choice{Type: ChoiceInteger, Value: int64(1)}),
			seq2: seqWith(Choice{Type: ChoiceInteger, Value: int64(1)}, Choice{Type: ChoiceInteger, Value: int64(2)}),
			want: -1,
		},
		{
			name: "equal sequences",
			seq1: seqWith(Choice{Type: ChoiceInteger, Value: int64(1)}),
			seq2: seqWith(Choice{Type: ChoiceInteger, Value: int64(1)}),
			want: 0,
		},
		{
			name: "lex order - smaller value first",
			seq1: seqWith(Choice{Type: ChoiceInteger, Value: int64(5)}),
			seq2: seqWith(Choice{Type: ChoiceInteger, Value: int64(10)}),
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.seq1.Compare(tt.seq2)
			if got != tt.want {
				t.Errorf("Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

// Helper to create a sequence with choices
func seqWith(choices ...Choice) *ChoiceSequence {
	seq := NewChoiceSequence()
	for _, c := range choices {
		seq.Append(c)
	}
	return seq
}
