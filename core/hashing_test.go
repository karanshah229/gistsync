package core

import (
	"testing"
)

func TestComputeHash(t *testing.T) {
	data := []byte("hello world")
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	actual := ComputeHash(data)
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
