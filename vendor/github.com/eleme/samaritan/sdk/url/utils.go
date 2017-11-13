package url

import "strings"

// Join concatenates the elements of a to create a single string, and separate with '/'.
func join(a ...string) string {
	return strings.Join(a, "/")
}
