package strutil

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringerFunc(t *testing.T) {
	fruits := []string{
		"apple", "banana", "melon", "papaya", "avocado", "lemon", "grape",
	}

	printFruitFn := func() string {
		sb := strings.Builder{}

		max := len(fruits) - 1
		for i, fruit := range fruits {
			sb.WriteString(strconv.Itoa(i + 1))
			sb.WriteString(". ")
			sb.WriteString(fruit)

			if i < max {
				sb.WriteString(", ")
			}
		}

		return sb.String()
	}

	assert.Equal(t, "1. apple, 2. banana, 3. melon, 4. papaya, 5. avocado, 6. lemon, 7. grape", fmt.Sprintf("%v", StringerFunc(printFruitFn)))
}
