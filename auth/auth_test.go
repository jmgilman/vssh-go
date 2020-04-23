package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAuthNames(t *testing.T) {
	expectedLength := len(Types)
	gotLength := len(GetAuthNames())

	assert.Equal(t, expectedLength, gotLength)
}
