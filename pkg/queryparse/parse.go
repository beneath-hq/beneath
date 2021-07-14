package queryparse

import "strings"

// StringToQuery attempts to parse str as either a JSON or Where query
func StringToQuery(str string) (Query, error) {
	str = strings.TrimSpace(str)
	if len(str) != 0 && str[0] == '{' {
		return JSONStringToQuery(str)
	}
	return WhereStringToQuery(str)
}
