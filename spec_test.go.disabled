package specta_test

import (
	"testing"

	"github.com/james-w/specta"
)

func TestLit(t *testing.T) {
	p := specta.New()
	provider := specta.Lit(42)

	result := provider(p)
	if result != 42 {
		t.Errorf("Expected Lit to return 42, got %d", result)
	}

	// Lit should return the same value regardless of primitives state
	p.Next()
	result2 := provider(p)
	if result2 != 42 {
		t.Errorf("Expected Lit to consistently return 42, got %d", result2)
	}
}

func TestProvider(t *testing.T) {
	p := specta.New()

	// Provider that uses primitives
	provider := func(s specta.Source) int {
		return int(p.Next())
	}

	first := provider(p)
	second := provider(p)

	if first != 1 || second != 2 {
		t.Errorf("Expected Provider to use primitives, got %d, %d", first, second)
	}
}

func TestMaybe_Some(t *testing.T) {
	p := specta.New()
	m := specta.Some(specta.Lit(100))

	if !m.IsSet() {
		t.Error("Expected Some to create a set Maybe")
	}

	value := m.Value(p)
	if value != 100 {
		t.Errorf("Expected Value to return 100, got %d", value)
	}
}

func TestMaybe_Get_WithSet(t *testing.T) {
	p := specta.New()
	m := specta.Some(specta.Lit(100))
	defaultProvider := specta.Lit(200)

	result := m.Get(p, defaultProvider)
	if result != 100 {
		t.Errorf("Expected Get to return set value 100, got %d", result)
	}
}

func TestMaybe_Get_WithUnset(t *testing.T) {
	p := specta.New()
	var m specta.Maybe[int] // unset
	defaultProvider := specta.Lit(200)

	result := m.Get(p, defaultProvider)
	if result != 200 {
		t.Errorf("Expected Get to return default value 200, got %d", result)
	}
}

func TestMaybe_IsSet(t *testing.T) {
	var unset specta.Maybe[int]
	set := specta.Some(specta.Lit(42))

	if unset.IsSet() {
		t.Error("Expected zero Maybe to not be set")
	}
	if !set.IsSet() {
		t.Error("Expected Some Maybe to be set")
	}
}

func TestMaybe_Value_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Value to panic on unset Maybe")
		}
	}()

	p := specta.New()
	var m specta.Maybe[int]
	_ = m.Value(p) // Should panic
}

func TestSpecFactory_Make(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	type PersonSpec struct {
		Name specta.Maybe[string]
		Age  specta.Maybe[int]
	}

	newSpec := func() PersonSpec {
		return PersonSpec{}
	}

	build := func(src specta.Source, spec PersonSpec) Person {
		return Person{
			Name: spec.Name.Get(src, func(s specta.Source) string { return specta.String().Draw(src, "") }),
			Age:  spec.Age.Get(src, func(s specta.Source) int { return int(specta.Int().Draw(src, "")) }),
		}
	}

	p := specta.New()
	factory := specta.NewSpecFactory(p, newSpec, build)

	// Test with defaults
	person1 := factory.Make()
	if person1.Name != "1" || person1.Age != 2 {
		t.Errorf("Expected default values, got Name='%s', Age=%d", person1.Name, person1.Age)
	}

	// Test with custom option
	setName := func(s *PersonSpec) {
		s.Name = specta.Some(specta.Lit("Alice"))
	}
	person2 := factory.Make(setName)
	if person2.Name != "Alice" || person2.Age != 3 {
		t.Errorf("Expected custom Name, got Name='%s', Age=%d", person2.Name, person2.Age)
	}
}

func TestSpecFactory_Many(t *testing.T) {
	type Item struct {
		ID int
	}

	type ItemSpec struct {
		ID specta.Maybe[int]
	}

	newSpec := func() ItemSpec {
		return ItemSpec{}
	}

	build := func(src specta.Source, spec ItemSpec) Item {
		return Item{
			ID: spec.ID.Get(src, func(s specta.Source) int { return int(specta.Int().Draw(src, "")) }),
		}
	}

	p := specta.New()
	factory := specta.NewSpecFactory(p, newSpec, build)

	// Test with positive n
	items := factory.Many(3)
	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}
	if items[0].ID != 1 || items[1].ID != 2 || items[2].ID != 3 {
		t.Errorf("Expected sequential IDs 1,2,3, got %d,%d,%d", items[0].ID, items[1].ID, items[2].ID)
	}

	// Test with zero n
	empty := factory.Many(0)
	if empty != nil {
		t.Errorf("Expected nil for Many(0), got %v", empty)
	}

	// Test with negative n
	negative := factory.Many(-5)
	if negative != nil {
		t.Errorf("Expected nil for Many(-5), got %v", negative)
	}
}

func TestSetLit(t *testing.T) {
	type PersonSpec struct {
		Name specta.Maybe[string]
	}

	assign := func(s *PersonSpec, m specta.Maybe[string]) {
		s.Name = m
	}

	opt := specta.SetLit(assign, "Bob")
	var spec PersonSpec
	opt(&spec)

	if !spec.Name.IsSet() {
		t.Error("Expected SetLit to set the Maybe")
	}

	p := specta.New()
	if spec.Name.Value(p) != "Bob" {
		t.Errorf("Expected SetLit to set value to 'Bob', got '%s'", spec.Name.Value(p))
	}
}

func TestSetWith(t *testing.T) {
	type PersonSpec struct {
		Age specta.Maybe[int]
	}

	assign := func(s *PersonSpec, m specta.Maybe[int]) {
		s.Age = m
	}

	provider := func(s specta.Source) int {
		return int(s.DrawBits(64) * 10)
	}

	opt := specta.SetWith(assign, provider)
	var spec PersonSpec
	opt(&spec)

	if !spec.Age.IsSet() {
		t.Error("Expected SetWith to set the Maybe")
	}

	p := specta.New()
	if spec.Age.Value(p) != 10 {
		t.Errorf("Expected SetWith to use provider (1*10=10), got %d", spec.Age.Value(p))
	}
}

func TestCompose(t *testing.T) {
	type PersonSpec struct {
		Name specta.Maybe[string]
		Age  specta.Maybe[int]
	}

	assignName := func(s *PersonSpec, m specta.Maybe[string]) {
		s.Name = m
	}
	assignAge := func(s *PersonSpec, m specta.Maybe[int]) {
		s.Age = m
	}

	opt1 := specta.SetLit(assignName, "Charlie")
	opt2 := specta.SetLit(assignAge, 25)
	composed := specta.Compose(opt1, opt2)

	var spec PersonSpec
	composed(&spec)

	p := specta.New()
	if spec.Name.Value(p) != "Charlie" || spec.Age.Value(p) != 25 {
		t.Errorf("Expected Compose to apply all opts, got Name='%s', Age=%d",
			spec.Name.Value(p), spec.Age.Value(p))
	}
}

func TestFromSpec(t *testing.T) {
	type Address struct {
		Street string
	}

	type AddressSpec struct {
		Street specta.Maybe[string]
	}

	newSpec := func() AddressSpec {
		return AddressSpec{}
	}

	build := func(src specta.Source, spec AddressSpec) Address {
		return Address{
			Street: spec.Street.Get(src, func(s specta.Source) string { return specta.String().Draw(src, "") }),
		}
	}

	assignStreet := func(s *AddressSpec, m specta.Maybe[string]) {
		s.Street = m
	}

	provider := specta.FromSpec(build, newSpec, specta.SetLit(assignStreet, "Main St"))

	p := specta.New()
	addr := provider(p)

	if addr.Street != "Main St" {
		t.Errorf("Expected FromSpec to build with opts, got Street='%s'", addr.Street)
	}
}

func TestFromSpecN(t *testing.T) {
	type Item struct {
		ID int
	}

	type ItemSpec struct {
		ID specta.Maybe[int]
	}

	newSpec := func() ItemSpec {
		return ItemSpec{}
	}

	build := func(src specta.Source, spec ItemSpec) Item {
		return Item{
			ID: spec.ID.Get(src, func(s specta.Source) int { return int(specta.Int().Draw(src, "")) }),
		}
	}

	provider := specta.FromSpecN(3, build, newSpec)

	p := specta.New()
	items := provider(p)

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}
	if items[0].ID != 1 || items[1].ID != 2 || items[2].ID != 3 {
		t.Errorf("Expected IDs 1,2,3, got %d,%d,%d", items[0].ID, items[1].ID, items[2].ID)
	}

	// Test with zero/negative
	emptyProvider := specta.FromSpecN(0, build, newSpec)
	empty := emptyProvider(p)
	if empty != nil {
		t.Errorf("Expected nil for FromSpecN(0), got %v", empty)
	}
}

func TestFromSpecEach(t *testing.T) {
	type Item struct {
		ID    int
		Index int
	}

	type ItemSpec struct {
		ID    specta.Maybe[int]
		Index specta.Maybe[int]
	}

	newSpec := func() ItemSpec {
		return ItemSpec{}
	}

	build := func(src specta.Source, spec ItemSpec) Item {
		return Item{
			ID:    spec.ID.Get(src, func(s specta.Source) int { return int(specta.Int().Draw(src, "")) }),
			Index: spec.Index.Get(src, func(s specta.Source) int { return -1 }),
		}
	}

	edit := func(i int, s *ItemSpec) {
		s.Index = specta.Some(specta.Lit(i * 10))
	}

	provider := specta.FromSpecEach(3, build, newSpec, edit)

	p := specta.New()
	items := provider(p)

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}
	if items[0].Index != 0 || items[1].Index != 10 || items[2].Index != 20 {
		t.Errorf("Expected Index 0,10,20, got %d,%d,%d", items[0].Index, items[1].Index, items[2].Index)
	}

	// Test with nil edit
	providerNoEdit := specta.FromSpecEach(2, build, newSpec, nil)
	itemsNoEdit := providerNoEdit(p)
	if len(itemsNoEdit) != 2 {
		t.Errorf("Expected 2 items with nil edit, got %d", len(itemsNoEdit))
	}
}

func TestPtrOf(t *testing.T) {
	provider := specta.Lit(42)
	ptrProvider := specta.PtrOf(provider)

	p := specta.New()
	ptr := ptrProvider(p)

	if ptr == nil {
		t.Fatal("Expected PtrOf to return non-nil pointer")
	}
	if *ptr != 42 {
		t.Errorf("Expected *ptr to be 42, got %d", *ptr)
	}
}

func TestSliceOf(t *testing.T) {
	provider := specta.SliceOf(
		specta.Lit(1),
		specta.Lit(2),
		specta.Lit(3),
	)

	p := specta.New()
	slice := provider(p)

	if len(slice) != 3 {
		t.Fatalf("Expected slice of length 3, got %d", len(slice))
	}
	if slice[0] != 1 || slice[1] != 2 || slice[2] != 3 {
		t.Errorf("Expected [1, 2, 3], got %v", slice)
	}

	// Test with dynamic providers
	dynProvider := specta.SliceOf(
		func(s specta.Source) int { return int(specta.Int().Draw(s, "")) },
		func(s specta.Source) int { return int(specta.Int().Draw(s, "")) },
	)
	p2 := specta.New()
	dynSlice := dynProvider(p2)
	if dynSlice[0] != 1 || dynSlice[1] != 2 {
		t.Errorf("Expected [1, 2] from dynamic providers, got %v", dynSlice)
	}
}
