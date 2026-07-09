// Package semver provides a small semantic-version comparison helper.
package semver

import (
	"strconv"
	"strings"
)

// Compare returns -1 if a < b, 0 if equal, 1 if a > b.
// It understands "v" prefixes and numeric pre-release parts (e.g. -beta, -rc1).
func Compare(a, b string) int {
	a = normalize(a)
	b = normalize(b)
	if a == b {
		return 0
	}

	aCore, aPre := splitPreRelease(a)
	bCore, bPre := splitPreRelease(b)

	aParts := splitVersion(aCore)
	bParts := splitVersion(bCore)

	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		av := 0
		if i < len(aParts) {
			av = aParts[i]
		}
		bv := 0
		if i < len(bParts) {
			bv = bParts[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}

	// 1.0.0 > 1.0.0-alpha
	if aPre == "" && bPre != "" {
		return 1
	}
	if aPre != "" && bPre == "" {
		return -1
	}
	if aPre != "" && bPre != "" {
		return comparePreRelease(aPre, bPre)
	}
	return 0
}

func normalize(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

func splitPreRelease(v string) (core, pre string) {
	// Build metadata after '+' is ignored for comparison.
	if idx := strings.Index(v, "+"); idx >= 0 {
		v = v[:idx]
	}
	idx := strings.Index(v, "-")
	if idx < 0 {
		return v, ""
	}
	return v[:idx], v[idx+1:]
}

func splitVersion(v string) []int {
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			n = 0
		}
		out = append(out, n)
	}
	return out
}

func comparePreRelease(a, b string) int {
	aParts := splitPreReleaseParts(a)
	bParts := splitPreReleaseParts(b)
	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		var av, bv preReleasePart
		if i < len(aParts) {
			av = aParts[i]
		}
		if i < len(bParts) {
			bv = bParts[i]
		}
		if av.numeric && !bv.numeric {
			return -1
		}
		if !av.numeric && bv.numeric {
			return 1
		}
		if av.numeric && bv.numeric {
			if av.value < bv.value {
				return -1
			}
			if av.value > bv.value {
				return 1
			}
		} else {
			cmp := strings.Compare(av.text, bv.text)
			if cmp != 0 {
				return cmp
			}
		}
	}
	return 0
}

type preReleasePart struct {
	text    string
	value   int
	numeric bool
}

func splitPreReleaseParts(pre string) []preReleasePart {
	parts := strings.Split(pre, ".")
	out := make([]preReleasePart, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err == nil {
			out = append(out, preReleasePart{value: n, numeric: true})
		} else {
			out = append(out, preReleasePart{text: p})
		}
	}
	return out
}
