package utils

import (
	"strings"
)

func WhereIn(column string, values []interface{}, inRange bool) (string) {
	if column == "" {
		return ""
	}

	if len(values) == 0 {
		return ""
	}

	var stmt string

	if len(values) > 1 {
		stmt = "(?" + strings.Repeat(", ?", len(values) - 1) + ")"

		if inRange == true {
			stmt = "IN " + stmt
		} else {
			stmt = "NOT IN " + stmt
		}

		stmt = column + " " + stmt
	} else {
		stmt = column + " = ?"
	}

	return stmt
}