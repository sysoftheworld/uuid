package uuid

import (
	"fmt"
	"regexp"
	"testing"
)

const (
	testSize = 100000
)

var (
	uuidRegex = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
)

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

func TestVersion(t *testing.T) {
	uuid := UUID{}

	for i := 0; i < 6; i++ {
		uuid.version(uint8(i))
		if uuid[6] != uint8(i<<4) {
			t.Error("Version is not correct:", uuid[6], "should be:", uint8(i<<4))
		}
	}

	fmt.Println("")
}

func TestVariant(t *testing.T) {
	uuid := UUID{}

	for i := uint8(0); i < 0xFF; i++ {
		uuid[8] = i
		uuid.variant(rfc4122)

		if uuid[8] < 0x0F || uuid[8] > 0xBF {
			t.Error("Varient is not correct", uuid[8], "at", i)
		}
	}

}

func TestRegexV1(t *testing.T) {

	for i := 0; i < testSize; i++ {
		uuid := New(1)

		if !uuidRegex.MatchString(uuid.String()) {
			t.Error("V1 does not pass regex test", uuid.String())
		}
	}
}

func TestMutexV1(t *testing.T) {

	for i := 0; i < testSize/10; i++ {
		go func() {
			New(1)
		}()
	}
}

func TestCollisionV1(t *testing.T) {
	uuids := make(map[UUID]uint8)

	for i := 0; i < testSize; i++ {
		uuid := New(1)

		_, ok := uuids[uuid]

		if ok == true { //collision
			t.Error("Collision V1:", uuid.String())
		} else {
			uuids[uuid] = 0
		}
	}
}

func TestRegexV2(t *testing.T) {

	for i := 0; i < testSize; i++ {
		uuid := New(2)
		if !uuidRegex.MatchString(uuid.String()) {
			t.Error("V2 does not pass regex test", uuid.String())
		}
	}
}

func TestMutexV2(t *testing.T) {

	for i := 0; i < testSize/10; i++ {
		go func() {
			New(2)
		}()
	}
}

func TestCollisionV2(t *testing.T) {
	uuids := make(map[UUID]uint8)

	for i := 0; i < testSize; i++ {
		uuid := New(2)

		_, ok := uuids[uuid]

		if ok == true { //collision
			t.Error("Collision V2:", uuid.String())
		} else {
			uuids[uuid] = 0
		}
	}
}

func TestRegexV4(t *testing.T) {

	for i := 0; i < testSize; i++ {
		uuid := New(4)

		if !uuidRegex.MatchString(uuid.String()) {
			t.Error("V4 does not pass regex test", uuid.String())
		}
	}
}

func TestMutexV4(t *testing.T) {

	for i := 0; i < testSize/10; i++ {
		go func() {
			New(4)
		}()
	}
}

func TestCollisionV4(t *testing.T) {
	uuids := make(map[UUID]uint8)

	for i := 0; i < testSize; i++ {
		uuid := New(4)

		_, ok := uuids[uuid]

		if ok == true { //collision
			t.Error("Collision V4:", uuid.String())
		} else {
			uuids[uuid] = 0
		}
	}
}

func TestClockSeqInit(t *testing.T) {
	var cs uint16
	var dup int

	for i := 0; i < testSize; i++ {
		temp := clockSeqInit()

		if cs == temp {
			dup++
		}

		cs = temp
	}
	if dup > 10 {
		t.Error("Clock Sequence is not random")
	}
}
