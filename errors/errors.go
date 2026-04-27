package errors

// Kind categorizes an error for rendering and exit code mapping.
type Kind int

const (
	KindUser Kind = iota + 1
	KindInternal
	KindConfig
)

// Kinder is implemented by errors that carry a Kind.
type Kinder interface {
	ErrorKind() Kind
}

// Hinter is implemented by errors that carry a user-facing hint.
type Hinter interface {
	Hint() string
}
