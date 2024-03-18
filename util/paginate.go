package util

func GetOffset(page, limit int) int {
	return limit * (page - 1)
}
