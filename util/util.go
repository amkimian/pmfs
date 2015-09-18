package util

import "strings"

// Path management.
// Given a current working directory (e.g. /alan)
// and a relative path (..) (.) /fred fred ../fred
// return / /alan /fred /alan/fred /fred
func ResolvePath(cwd string, relative string) string {
	if len(relative) == 0 {
		return cwd
	}
	if relative[0] == '/' {
		return relative
	}
	parts := strings.Split(relative, "/")
	dir := strings.Split(cwd, "/")
	for p := range parts {
		switch parts[p] {
		case ".":
		case "..":
			if len(dir) > 0 {
				dir = dir[:len(dir)-1]
			}
		default:
			dir = append(dir, parts[p])
		}
	}
	return strings.Join(dir, "/")
}
