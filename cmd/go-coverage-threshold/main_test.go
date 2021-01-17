package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoPath(t *testing.T) {
	t.Parallel()

	got, _, err := goPath()
	assert.NotEmpty(t, got)
	assert.Nil(t, err)
	// assert.NotEmpty(t, module)
}
