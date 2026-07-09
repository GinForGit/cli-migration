package semver

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.0", 0},
		{"v1.2.0", "1.3.0", -1},
		{"1.10.0", "1.2.0", 1},
		{"1.0.0-beta", "1.0.0", -1},
		{"1.0.0", "1.0.0-beta", 1},
		{"1.0.0-rc1", "1.0.0-rc2", -1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0+build", "1.0.0", 0},
		{"1.2.3", "1.2.3.0", 0},
		{"1.2", "1.2.0", 0},
	}

	for _, c := range cases {
		got := Compare(c.a, c.b)
		if got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
