package etrace

import (
	"fmt"
	"strings"
)

// BuildSQL builds a full SQL string from query string and args.
func BuildSQL(query string, args ...interface{}) string {
	qfmt := strings.Replace(query, "?", "%v", -1)
	return fmt.Sprintf(qfmt, args...)
}

// parsesSimpleSQL returns table name and operation name, a copy from zeus-core..
func parsesSimpleSQL(sql string) string {
	parts := strings.Split(sql, "*/")
	code := strings.ToLower(parts[len(parts)-1])
	elements := strings.Split(strings.Trim(code, " "), " ")
	nelements := len(elements)
	if nelements == 0 {
		return "unknown.sql"
	}

	op := elements[0]
	table := "unknown"
	switch op {
	case "insert", "delete":
		if nelements > 2 {
			table = elements[2]
		}
	case "select":
		for idx, elem := range elements {
			if elem == "from" {
				if target := idx + 1; target < nelements {
					table = elements[target]
				}
				break
			}
		}
	case "update":
		if nelements > 1 {
			table = elements[1]
		}
	}
	return fmt.Sprintf("%s.%s", table, op)
}
