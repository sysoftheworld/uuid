package uuid

import (
	"testing"
)

// This will test FromString and FromByte
// but only with cases that should work
func TestNamespaceInit(t *testing.T) {
	if err := initNamespace(); err != nil {
		t.Fatal(err)
	}
}
