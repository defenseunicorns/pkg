package other

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	if "hello world" != "hello world" {
		t.Error("test fail")
	}
	require.Empty(t, "")
}
