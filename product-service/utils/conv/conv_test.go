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

// bitSize must be 64: with 10 the parser would reject anything above 511.
func TestStringToInt64(t *testing.T) {
	cases := map[string]int64{
		"1024":   1024,
		"0":      0,
		"-7":     -7,
		"999999": 999999,
	}

	for in, want := range cases {
		got, err := StringToInt64(in)
		if err != nil {
			t.Errorf("StringToInt64(%q) returned error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("StringToInt64(%q) = %d, want %d", in, got, want)
		}
	}

	if _, err := StringToInt64("abc"); err == nil {
		t.Error(`StringToInt64("abc") returned no error`)
	}
}
