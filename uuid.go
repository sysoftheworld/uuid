package uuid

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"sync"
)

const (
	uuidSize = 16

	// https://tools.ietf.org/html/rfc4122#section-4.1.1
	NCS       = 0x00
	RFC4122   = 0x04
	Microsoft = 0x05
	Future    = 0x07
)

var (
	mu         = sync.Mutex{} // global mutex to prevent races on timeSource and clockSeq
	timeSource timestamp      // please see timestamp.go for info
	addr       [6]byte        // hardware address used for v1 and v2
	cs         clockSeq       // used for v1 and v2
)

func init() {
	addr = hardwareAddr()
	cs.Init()
}

type clockSeq struct {
	seq uint16
}

// Set the clock to random bytes
func (cs *clockSeq) Init() {
	var b [2]byte
	randomBytes(b[:])
	cs.seq = binary.BigEndian.Uint16(b[:])
}

type UUID [uuidSize]byte

// New and String are the only two Public functions

// Return a copy of a UUID
func New(ver int) UUID {

	var u UUID

	switch ver {
	default:
		u.v1()
	case 2:
		u.v2()
	case 4:
		u.v4()
	}

	return u
}

// Format in bytes 4-2-2-2-6
func (u *UUID) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(u[:4]), hex.EncodeToString(u[4:6]), hex.EncodeToString(u[6:8]), hex.EncodeToString(u[8:10]), hex.EncodeToString(u[10:16]))
}

func (u *UUID) v1() {

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidTime{}

	insertTimestamp(u[:], timeSource.timestamp())
	u.version(1)
	u.variant(RFC4122)

	cs.seq++

	binary.BigEndian.PutUint16(u[8:], cs.seq)
	copy(u[10:], addr[:])
}

func (u *UUID) v2() {

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidDCE{}
	insertTimestamp(u[:], timeSource.timestamp())
	u.version(2)

	cs.seq++

	binary.BigEndian.PutUint16(u[8:], cs.seq)
	copy(u[10:], addr[:])

}

func (u *UUID) v3() {

}

// See https://tools.ietf.org/html/rfc4122#section-4.4
func (u *UUID) v4() {

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidRand{}
	insertTimestamp(u[:], timeSource.timestamp())
	u.version(4)

	// From Doc: Set the two most significant bits (bits 6 and 7) of
	// the clock_seq_hi_and_reserved to zero and one, respectively.
	// I am assuming bits are 0 index meaning 7 not 8 is highest bit
	u[8] = u[8] | 0x64

	u.variant(RFC4122)
	// From Doc: Set all the other bits to randomly (or pseudo-randomly) chosen values
	randomBytes(u[9:])
}

func (u *UUID) v5() {

}

// https://tools.ietf.org/html/rfc4122 (Section: 4.1.3)
// The version number is in the most significant 4 bits of the time
// stamp (bits 4 through 7 of the time_hi_and_version field).
func (u *UUID) version(v byte) {
	u[6] = (v & 0x0F) | (v << 4)
}

// https://tools.ietf.org/html/rfc4122#section-4.1.1
func (u *UUID) variant(v byte) {
	u[8] = (v & 0xBF) | (v << 4)
}

// Timestamp layout and byte order https://tools.ietf.org/html/rfc4122#section-4.1.2
func insertTimestamp(b []byte, t uint64) {
	binary.BigEndian.PutUint32(b[0:], uint32(t))
	binary.BigEndian.PutUint16(b[4:], uint16(t>>32))
	binary.BigEndian.PutUint16(b[6:], uint16(t>>48))
}

// https://tools.ietf.org/html/rfc4122 (Section: 4.1.6)
// Address attempts to grab a hardware address that is 6 bytes or greater
// If there is more than one, first one found is ok
// If one cannot be found the byte array is randomized in accordanize with Section 4.1.6
func hardwareAddr() [6]byte {

	var addr [6]byte
	inter, err := net.Interfaces()

	// if there is an error with interfaces
	// don't panic just randomize
	if err != nil {
		randomBytes(addr[:])
		return addr
	}

	for _, i := range inter {
		if len(i.HardwareAddr) > 5 {
			copy(addr[:], i.HardwareAddr)
			return addr
		}
	}

	// if we got here no hardware address is set;
	// randomize it
	randomBytes(addr[:])
	return addr
}

// See https://golang.org/pkg/math/rand/#Read
func randomBytes(b []byte) {
	_, err := rand.Read(b)

	if err != nil {
		panic(err) // should panic if rand throws and error
	}
}
