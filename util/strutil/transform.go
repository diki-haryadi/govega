package strutil

import "strings"

// ToCommaSeparatedList shorthand of join string with comma as separator
// Which also perform trimspace for each elements in string
func ToCommaSeparatedList(elems []string) string {
	sep := ","

	// modified version of strings.Join, which also perform trimspace for each elements
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(strings.TrimSpace(elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(strings.TrimSpace(s))
	}
	return b.String()
}

// FromCommaSeparatedList turn comma separated list into slice of string
// additionally it will remove any leading and trailing space for each item
func FromCommaSeparatedList(s string) []string {
	sep := ","
	sepSave := 0

	// modified version of strings.Split, which also perform trimspace for each elements
	n := strings.Count(s, sep) + 1

	a := make([]string, n)
	n--
	i := 0
	for i < n {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}
		a[i] = strings.TrimSpace(s[:m+sepSave])
		s = s[m+len(sep):]
		i++
	}
	a[i] = strings.TrimSpace(s)
	return a[:i+1]
}
