package errors

// Kind categorizes an error for rendering and exit code mapping.
type Kind int

const (
	KindUser Kind = iota + 1
	KindInternal
	KindConfig
)

// firstCustomKind is where application-defined kinds start, far above the
// built-in constants so they never collide.
const firstCustomKind Kind = 1000

// CustomKind returns an application-defined Kind for n >= 0. Distinct n give
// distinct kinds, and none equals a built-in kind. Declare them once as
// package-level values and map them to exit codes with klee.App.WithExitCodes.
func CustomKind(n int) Kind {
	if n < 0 {
		panic("errors: CustomKind requires n >= 0")
	}
	return firstCustomKind + Kind(n)
}

// Kinder is implemented by errors that carry a Kind.
type Kinder interface {
	ErrorKind() Kind
}

// Hinter is implemented by errors that carry a user-facing hint.
type Hinter interface {
	Hint() string
}
