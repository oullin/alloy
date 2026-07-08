package compiler

// SourceRoute is the minimum surface a compilable route must expose.
type SourceRoute interface {
	Path() string
	Host() string
	Requirements() map[string]string
	HasDefault(name string) bool
}
