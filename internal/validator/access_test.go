package validator

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestACLValidatior(t *testing.T) {
	type TestCase struct {
		value string
		expected bool
	}

	cases := map[string]TestCase{
		"ok_rwx": {
			value: "rwx",
			expected: true,
		},
		"ok_rw-": {
			value: "rw-",
			expected: true,
		},
		"ok_r-x": {
			value: "r-x",
			expected: true,
		},
		"ok_r--": {
			value: "r--",
			expected: true,
		},
		"!ok_-wx": {
			value: "-wx",
			expected: false,
		},
		"!ok_rnd": {
			value: "rnd",
			expected: false,
		},
		"!ok_long": {
			value: "rwxxxxxxxxxx",
			expected: false,
		},
		"!ok_bad_long": {
			value: "safgafeqw123",
			expected: false,
		},
	}

	validator := NewACL()

	for tcname, tcase := range cases {
		t.Logf("Case %s\n", tcname)
		assert.Equal(t, tcase.expected, validator.Validate(tcase.value))
	}
}