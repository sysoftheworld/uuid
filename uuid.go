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
	rfc4122 = 0x04
	future  = 0x07
)

var (
	mu         = sync.Mutex{} // global mutex to prevent races on timeSource and clockSeq
	timeSource timestamp      // please see timestamp.go for info
	addr       [6]byte        // hardware address used for v1 and v2
	clockSeq   uint16         // used for v1 and v2
)

func init() {
	addr = hardwareAddr()
	clockSeq = clockSeqInit()
}

// UUID ...
type UUID [uuidSize]byte

// New and String are the only two Public functions

// New returns a copy of a UUID
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

	clockSeq++

	binary.BigEndian.PutUint16(u[8:], clockSeq)
	u.variant(rfc4122) // must set after setting clockSeq

	copy(u[10:], addr[:])
}

func (u *UUID) v2() {

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidDCE{}
	insertTimestamp(u[:], timeSource.timestamp())
	u.version(2)

	clockSeq++

	binary.BigEndian.PutUint16(u[8:], clockSeq)
	u.variant(rfc4122) // must set after setting clockSeq
	copy(u[10:], addr[:])

}

// See https://tools.ietf.org/html/rfc4122#section-4.4
func (u *UUID) v4() {

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidRand{}
	insertTimestamp(u[:], timeSource.timestamp())
	u.version(4)

	u.variant(rfc4122)
	// From Doc: Set all the other bits to randomly (or pseudo-randomly) chosen values
	randomBytes(u[9:])
}

// https://tools.ietf.org/html/rfc4122 (Section: 4.1.3)
// The version number is in the most significant 4 bits of the time
// stamp (bits 4 through 7 of the time_hi_and_version field).
func (u *UUID) version(v byte) {
	u[6] = (u[6] & 0x0F) | (v << 4)
}

// https://tools.ietf.org/html/rfc4122#section-4.1.1
func (u *UUID) variant(v byte) {
	var mask byte

	//0x3F clear top 2
	//0x1F clear top 3

	switch v {
	default:
		mask = 0x3F
	case rfc4122:
		mask = 0x3F
	case future:
		mask = 0x1F
	}

	u[8] = (u[8] & mask) | (v << 5)
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

// Set the clock to random bytes
func clockSeqInit() uint16 {
	var b [2]byte
	randomBytes(b[:])
	return binary.BigEndian.Uint16(b[:])
}

// See https://golang.org/pkg/math/rand/#Read
func randomBytes(b []byte) {
	_, err := rand.Read(b)

	if err != nil {
		panic(err) // should panic if rand throws and error
	}
}
