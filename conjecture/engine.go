package conjecture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Engine orchestrates property-based test execution.
// It manages generation, test execution, shrinking, and example persistence.
type Engine struct {
	settings Settings
	database Database
}

// Settings configures the test engine.
type Settings struct {
	// MaxExamples is the number of test cases to generate.
	MaxExamples int

	// MaxShrinkCalls limits shrinking iterations.
	MaxShrinkCalls int

	// Timeout for the entire test run.
	Timeout time.Duration

	// Seed for reproducibility (0 = random).
	Seed uint64

	// Database for storing/loading examples.
	Database Database

	// Verbosity controls logging output.
	Verbosity int

	// Phases controls which phases run.
	Phases Phases
}

// Phases controls which test phases are enabled.
type Phases struct {
	Reuse    bool // Replay examples from database
	Generate bool // Generate new examples
	Shrink   bool // Shrink failing examples
}

// DefaultSettings returns sensible defaults.
func DefaultSettings() Settings {
	return Settings{
		MaxExamples:    100,
		MaxShrinkCalls: 5000,
		Timeout:        time.Minute,
		Seed:           0, // Will use time-based seed
		Verbosity:      0,
		Phases: Phases{
			Reuse:    true,
			Generate: true,
			Shrink:   true,
		},
	}
}

// SettingsOption configures settings.
type SettingsOption func(*Settings)

// WithMaxExamples sets the number of examples to generate.
func WithMaxExamples(n int) SettingsOption {
	return func(s *Settings) {
		s.MaxExamples = n
	}
}

// WithTimeout sets the test timeout.
func WithTimeout(d time.Duration) SettingsOption {
	return func(s *Settings) {
		s.Timeout = d
	}
}

// WithSeedValue sets the random seed for reproducibility.
func WithSeedValue(seed uint64) SettingsOption {
	return func(s *Settings) {
		s.Seed = seed
	}
}

// WithDatabase sets the example database.
func WithDatabase(db Database) SettingsOption {
	return func(s *Settings) {
		s.Database = db
	}
}

// NewEngine creates a test engine with the given settings.
func NewEngine(opts ...SettingsOption) *Engine {
	settings := DefaultSettings()
	for _, opt := range opts {
		opt(&settings)
	}

	if settings.Seed == 0 {
		settings.Seed = uint64(time.Now().UnixNano())
	}

	if settings.Database == nil {
		settings.Database = NewMemoryDatabase()
	}

	return &Engine{
		settings: settings,
		database: settings.Database,
	}
}

// TestResult contains the outcome of running a property test.
type TestResult struct {
	// Success is true if no failing example was found.
	Success bool

	// FailingExample is the minimal failing choice sequence (if any).
	FailingExample *ChoiceSequence

	// FailureReason describes why the test failed.
	FailureReason string

	// Statistics about the test run.
	Stats TestStats
}

// TestStats contains statistics about a test run.
type TestStats struct {
	ValidExamples   int
	InvalidExamples int
	TotalCalls      int
	ShrinkCalls     int
	Duration        time.Duration
}

// Property represents a testable property.
// The function should panic or return an error to indicate failure.
type Property func(d *ConjectureData) error

// Run executes a property test.
func (e *Engine) Run(ctx context.Context, name string, prop Property) TestResult {
	startTime := time.Now()

	// Create timeout context
	if e.settings.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.settings.Timeout)
		defer cancel()
	}

	result := TestResult{
		Success: true,
		Stats:   TestStats{},
	}

	// Wrap property to catch panics
	wrappedProp := func(d *ConjectureData) (interesting bool) {
		defer func() {
			if r := recover(); r != nil {
				d.MarkInteresting(fmt.Sprintf("panic: %v", r))
				interesting = true
			}
		}()

		err := prop(d)
		if err != nil {
			d.MarkInteresting(err.Error())
			return true
		}
		return false
	}

	// Hash the property name for database key
	key := hashKey(name)

	// Phase 1: Replay examples from database
	if e.settings.Phases.Reuse {
		for _, seq := range e.database.Load(key) {
			select {
			case <-ctx.Done():
				goto done
			default:
			}

			data := ForReplay(seq)
			result.Stats.TotalCalls++

			if wrappedProp(data) {
				result.Success = false
				result.FailingExample = data.Sequence()
				result.FailureReason = data.Tags()["failure_reason"].(string)
				goto shrink
			}
		}
	}

	// Phase 2: Generate new examples
	if e.settings.Phases.Generate {
		random := NewPCGRandom(e.settings.Seed)

		for i := 0; i < e.settings.MaxExamples; i++ {
			select {
			case <-ctx.Done():
				goto done
			default:
			}

			data := NewConjectureData(WithSeed(random.Uint64()))
			result.Stats.TotalCalls++

			interesting := wrappedProp(data)
			data.Freeze()

			switch data.Status() {
			case StatusValid:
				result.Stats.ValidExamples++
			case StatusInvalid:
				result.Stats.InvalidExamples++
			case StatusInteresting:
				result.Success = false
				result.FailingExample = data.Sequence()
				if reason, ok := data.Tags()["failure_reason"].(string); ok {
					result.FailureReason = reason
				}
				goto shrink
			}

			if interesting {
				result.Success = false
				result.FailingExample = data.Sequence()
				goto shrink
			}
		}
	}
	goto done

shrink:
	// Phase 3: Shrink the failing example
	if e.settings.Phases.Shrink && result.FailingExample != nil {
		shrinkTest := func(d *ConjectureData) bool {
			wrappedProp(d)
			return d.Status() == StatusInteresting
		}

		shrinkResult := Shrink(ctx, result.FailingExample, shrinkTest,
			WithMaxShrinkCalls(e.settings.MaxShrinkCalls))

		result.FailingExample = shrinkResult.Sequence
		result.Stats.ShrinkCalls = shrinkResult.Calls

		// Update failure reason from shrunk example
		data := ForReplay(result.FailingExample)
		wrappedProp(data)
		if reason, ok := data.Tags()["failure_reason"].(string); ok {
			result.FailureReason = reason
		}

		// Save to database
		e.database.Save(key, result.FailingExample)
	}

done:
	result.Stats.Duration = time.Since(startTime)
	return result
}

// RunProperty is a convenience function for running a single property.
func RunProperty(ctx context.Context, name string, prop Property, opts ...SettingsOption) TestResult {
	engine := NewEngine(opts...)
	return engine.Run(ctx, name, prop)
}

// hashKey creates a database key from a property name.
func hashKey(name string) string {
	h := sha256.Sum256([]byte(name))
	return hex.EncodeToString(h[:16])
}

// =============================================================================
// Database Interface
// =============================================================================

// Database stores and retrieves failing examples for replay.
type Database interface {
	// Save stores a failing example.
	Save(key string, seq *ChoiceSequence)

	// Load retrieves all stored examples for a key.
	Load(key string) []*ChoiceSequence

	// Delete removes all examples for a key.
	Delete(key string)
}

// MemoryDatabase is an in-memory database (doesn't persist).
type MemoryDatabase struct {
	mu       sync.RWMutex
	examples map[string][]*ChoiceSequence
}

// NewMemoryDatabase creates a new in-memory database.
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		examples: make(map[string][]*ChoiceSequence),
	}
}

func (db *MemoryDatabase) Save(key string, seq *ChoiceSequence) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Avoid duplicates
	for _, existing := range db.examples[key] {
		if existing.Compare(seq) == 0 {
			return
		}
	}

	db.examples[key] = append(db.examples[key], seq.Clone())
}

func (db *MemoryDatabase) Load(key string) []*ChoiceSequence {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make([]*ChoiceSequence, len(db.examples[key]))
	for i, seq := range db.examples[key] {
		result[i] = seq.Clone()
	}
	return result
}

func (db *MemoryDatabase) Delete(key string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.examples, key)
}

// =============================================================================
// Assertion Helpers
// =============================================================================

// Checker is a helper for common assertion patterns.
type Checker struct {
	data *ConjectureData
}

// NewChecker creates a checker for the given data context.
func NewChecker(d *ConjectureData) Checker {
	return Checker{data: d}
}

// That checks a condition and fails if false.
func (c Checker) That(cond bool, msg string, args ...any) error {
	if !cond {
		return fmt.Errorf(msg, args...)
	}
	return nil
}

// Equal checks that two values are equal.
func (c Checker) Equal(got, want any, name string) error {
	if got != want {
		return fmt.Errorf("%s: got %v, want %v", name, got, want)
	}
	return nil
}

// NoError checks that an error is nil.
func (c Checker) NoError(err error, msg string) error {
	if err != nil {
		return fmt.Errorf("%s: unexpected error: %w", msg, err)
	}
	return nil
}

// =============================================================================
// Fluent Property Builder
// =============================================================================

// ForAll creates a property builder for the given generator.
func ForAll[T any](g Gen[T]) *PropertyBuilder[T] {
	return &PropertyBuilder[T]{gen: g}
}

// PropertyBuilder allows fluent property construction.
type PropertyBuilder[T any] struct {
	gen      Gen[T]
	setup    func(T) error
	property func(T) error
	label    string
}

// Setup adds a setup/precondition check.
func (pb *PropertyBuilder[T]) Setup(f func(T) error) *PropertyBuilder[T] {
	pb.setup = f
	return pb
}

// Check sets the property to verify.
func (pb *PropertyBuilder[T]) Check(f func(T) error) *PropertyBuilder[T] {
	pb.property = f
	return pb
}

// WithLabel sets a label for the property.
func (pb *PropertyBuilder[T]) WithLabel(label string) *PropertyBuilder[T] {
	pb.label = label
	return pb
}

// Build creates the property function.
func (pb *PropertyBuilder[T]) Build() Property {
	return func(d *ConjectureData) error {
		value, err := pb.gen.Draw(d)
		if err != nil {
			// Drawing error (overrun, etc.) - not a property failure
			return nil
		}

		// Run setup if provided
		if pb.setup != nil {
			if err := pb.setup(value); err != nil {
				// Setup failure - filter this example
				return d.Assume(false)
			}
		}

		// Run the actual property
		if pb.property != nil {
			return pb.property(value)
		}

		return nil
	}
}

// Run executes the property with the given options.
func (pb *PropertyBuilder[T]) Run(ctx context.Context, opts ...SettingsOption) TestResult {
	name := pb.label
	if name == "" {
		name = pb.gen.String()
	}
	return RunProperty(ctx, name, pb.Build(), opts...)
}

// =============================================================================
// Error Types
// =============================================================================

// PropertyError wraps a property failure with context.
type PropertyError struct {
	Property string
	Message  string
	Sequence *ChoiceSequence
}

func (e *PropertyError) Error() string {
	return fmt.Sprintf("property %q failed: %s", e.Property, e.Message)
}

// IsPropertyError checks if an error is a property failure.
func IsPropertyError(err error) bool {
	var pe *PropertyError
	return errors.As(err, &pe)
}

// Reproduce replays a failing example to get the generated values.
func Reproduce[T any](g Gen[T], seq *ChoiceSequence) (T, error) {
	data := ForReplay(seq)
	return g.Draw(data)
}
