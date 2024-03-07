package helpers

import (
	"testing"
)

func TestFunc(t *testing.T) {
	if "hello world" != "hello world" {
		t.Error("test fail")
	}
}
