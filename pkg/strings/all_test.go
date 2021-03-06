package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllEqual(t *testing.T) {
	assertion := assert.New(t)

	listOfAlpha := All{"a", "a"}
	listOfStuff := All{"a", "b"}

	assertion.Equal(listOfAlpha.Equal("a"), true)
	assertion.Equal(listOfAlpha.Equal("b"), false)
	assertion.Equal(listOfStuff.Equal("a"), false)
}
