package validator

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestFNameValidatior(t *testing.T) {
	fname := "hello2.goi"

	v := NewFName()

	assert.Equal(t, v.Validate(fname), true)

	fname2 := "LSPaz-2143./xzc%#$$"

	assert.Equal(t, v.Validate(fname2), false)
}