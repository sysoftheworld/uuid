package uuid

import (
	"testing"
)

const ()

func TestInsertTimestamp(t *testing.T) {

	b := make([]byte, 8)
	full := uint64(0xFFFFFFFFFFFFFFFF)
	insertTimestamp(b, uint64(full))

	for _, v := range b {
		if v != 0xFF {
			t.Error("Not all bytes were set in Insert Timestamp", v)
		}
	}
}

func TestV1(t *testing.T) {
	uuid := New(1)

	println(uuid.String())
}
