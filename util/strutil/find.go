package strutil

// SliceContainsAny check if source slice contains any string in the search slice
func SliceContainsAny(source, search []string) bool {
	lookup := map[string]bool{}
	for _, value := range search {
		lookup[value] = true
	}

	for _, s := range source {
		if lookup[s] {
			return true
		}
	}

	return false
}
