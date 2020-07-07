package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStringEnv(t *testing.T) {
	os.Unsetenv("FOOBAR")
	assert.Equal(t, "default", GetStringEnv("FOOBAR", "default"))
	os.Setenv("FOOBAR", "specified")
	assert.Equal(t, "specified", GetStringEnv("FOOBAR", "default"))
}

func TestGetBoolEnv(t *testing.T) {
	os.Unsetenv("FOOBAR")
	assert.Equal(t, true, GetBoolEnv("FOOBAR", true))
	os.Setenv("FOOBAR", "false")
	assert.Equal(t, false, GetBoolEnv("FOOBAR", true))
	os.Setenv("FOOBAR", "1")
	assert.Equal(t, true, GetBoolEnv("FOOBAR", false))
	os.Setenv("FOOBAR", "ABCD")
	assert.Equal(t, false, GetBoolEnv("FOOBAR", false))
}
