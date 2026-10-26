package conv

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Elektronik & Gadget": "elektronik-gadget",
		"  Sayur  ":           "sayur",
		"Minuman---Dingin":    "minuman-dingin",
		"":                    "",
		"日本語":                 "",
	}

	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
