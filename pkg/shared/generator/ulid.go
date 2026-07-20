package generator

import "github.com/oklog/ulid/v2"

/*
NewULID generates a new ULID (Universally Unique Lexicographically Sortable
Identifier) as a string.
*/
func NewULID() string {
	return ulid.Make().String()
}
