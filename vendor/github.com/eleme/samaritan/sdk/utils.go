package sdk

import "strings"

// trimSpaces trims two strings' space characters by calling strings.TrimSpace.
func trimSpaces(strA, strB string) (string, string) {
	strA = strings.TrimSpace(strA)
	strB = strings.TrimSpace(strB)
	return strA, strB
}
