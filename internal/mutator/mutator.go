package mutator

// Mutator defines how a payload is generated for fuzzing.
// By using an interface, you can easily swap in format strings, SQLi, etc.
type Mutator interface {
	Generate() []byte
	Name() string
}
