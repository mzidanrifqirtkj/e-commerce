// Package conv holds small value conversions shared across layers.
package conv

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	slugDashEdge = regexp.MustCompile(`^-+|-+$`)
)

// Slugify turns "Elektronik & Gadget" into "elektronik-gadget".
//
// Non-ASCII runes are dropped, so a name made only of them slugifies to "".
// Callers that write the result to a UNIQUE column must reject that.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	return slugDashEdge.ReplaceAllString(s, "")
}

// StringToInt64 parses a base-10 string into an int64.
//
// The third argument is bitSize, not a range hint: anything below 64 clips the
// result (10 would cap it at -512..511), so it must stay 64.
func StringToInt64(s string) (int64, error) {
	newData, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return newData, nil
}

// StatusToBool maps the API status word to the DB boolean. ok is false for any
// word other than "published" or "unpublished".
func StatusToBool(s string) (v bool, ok bool) {
	switch s {
	case "published":
		return true, true
	case "unpublished":
		return false, true
	default:
		return false, false
	}
}

// BoolToStatus is the inverse of StatusToBool.
func BoolToStatus(v bool) string {
	if v {
		return "published"
	}
	return "unpublished"
}
