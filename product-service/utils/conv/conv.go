// Package conv holds small value conversions shared across layers.
package conv

import (
	"regexp"
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
