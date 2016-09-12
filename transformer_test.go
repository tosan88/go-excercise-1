package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransformStringReversedAndSwappedCase(t *testing.T) {
	assert.Equal(t, "HeLLO", transformString("ollEh"))
}

func TestTransformStringUnicodeReversedAndSwappedCase(t *testing.T) {
	assert.Equal(t, "hé⌘ű日本語", transformString("語本日Ű⌘ÉH"))
}

func TestTransformMultipleStringsReversedAndSwappedCase(t *testing.T) {
	assert.Equal(t, "HeLLO hé⌘ű日本語", transformString("ollEh 語本日Ű⌘ÉH"))
}

func TestTransformIntPositiveNumber(t *testing.T) {
	assert.Equal(t, "444", transformInt("321"))
}

func TestTransformIntNegativeNumber(t *testing.T) {
	assert.Equal(t, "23", transformInt("-100"))
}

func TestTransformIntNotValidNumber(t *testing.T) {
	assert.Equal(t, "test", transformInt("test"))
}

func TestTransformMultipleIntValidAndNotValidNumber(t *testing.T) {
	assert.Equal(t, "test 444", transformInt("test 321"))
}
