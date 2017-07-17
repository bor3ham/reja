package schema

type Server interface {
	GetDatabase() Database

	GetDefaultDirectPageSize() int
	GetMaximumDirectPageSize() int
	GetIndirectPageSize() int

	GetModel(string) *Model
	GetRoute(string) string

	Whitespace() bool
}
