package uuid

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"sync"
)

const (
	uuidSize = 16

	// https://tools.ietf.org/html/rfc4122#section-4.1.1
	rfc4122 = 0x04
	future  = 0x07
)

var (
	mu         = sync.Mutex{}   // global mutex to prevent races on timeSource and clockSeq
	timeSource timestamp        // please see timestamp.go for info
	addr       [6]byte          // hardware address used for v1 and v2
	clockSeq   = clockSeqInit() // used for v1 and v2

	uuidRegex = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")

	// ErrUUIDSize makes sure byte array is the correct size
	ErrUUIDSize = errors.New("UUID Size should 16 bytes")

	// ErrUUIDFormat will return if UUID does not pass uuidRegex
	ErrUUIDFormat = errors.New("UUID is not in the proper format")
)

func init() {
	addr = hardwareAddr()

	if err := initNamespace(); err != nil {
		panic(err)
	}
}

// UUID is 128 bits used to create a A Universally Unique IDentifier (UUID) URN Namespace
// Its specifications are described in RFC4122 and can be found https://tools.ietf.org/html/rfc4122
type UUID [uuidSize]byte

// NewV1 See https://tools.ietf.org/html/rfc4122#section-4.2.1
func NewV1() UUID {

	var uuid UUID

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidTime{}

	insertTimestamp(uuid[:], timeSource.timestamp())
	uuid.version(1)

	clockSeq++

	binary.BigEndian.PutUint16(uuid[8:], clockSeq)
	uuid.variant(rfc4122) // must set after setting clockSeq

	copy(uuid[10:], addr[:])

	return uuid
}

// NewV2 See http://pubs.opengroup.org/onlinepubs/9629399/apdxa.htm
func NewV2() UUID {

	var uuid UUID

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidDCE{}
	insertTimestamp(uuid[:], timeSource.timestamp())
	uuid.version(2)

	clockSeq++

	binary.BigEndian.PutUint16(uuid[8:], clockSeq)
	uuid.variant(rfc4122) // must set after setting clockSeq
	copy(uuid[10:], addr[:])

	return uuid

}

// NewV3 See https://tools.ietf.org/html/rfc4122#section-4.3
func NewV3(namespace UUID, name string) (UUID, error) {

	var uuid UUID

	h := md5.New()
	_, err := h.Write(namespace[:])

	if err != nil {
		return uuid, err
	}

	_, err = h.Write([]byte(name))

	if err != nil {
		return uuid, err
	}

	copy(uuid[:], h.Sum(nil))

	uuid.version(3)
	uuid.variant(rfc4122)

	return uuid, nil
}

// NewV4 See https://tools.ietf.org/html/rfc4122#section-4.4
func NewV4() UUID {

	var uuid UUID

	mu.Lock()
	defer mu.Unlock()

	timeSource = &uuidRand{}
	insertTimestamp(uuid[:], timeSource.timestamp())
	uuid.version(4)

	uuid.variant(rfc4122)
	// From Doc: Set all the other bits to randomly (or pseudo-randomly) chosen values
	randomBytes(uuid[9:])

	return uuid
}

// NewV5 See https://tools.ietf.org/html/rfc4122#section-4.3
func NewV5(namespace UUID, name string) (UUID, error) {

	var uuid UUID

	h := sha1.New()
	_, err := h.Write(namespace[:])

	if err != nil {
		return uuid, err
	}

	_, err = h.Write([]byte(name))

	if err != nil {
		return uuid, err
	}

	copy(uuid[:], h.Sum(nil))

	uuid.version(5)
	uuid.variant(rfc4122)

	return uuid, nil
}

// FromString will attempt to convert a uuid hex string into a uuid byte array
// if string does not pass regex text ErrUUIDFormat will be returned
func FromString(s string) (UUID, error) {

	var uuid UUID

	s = strings.Replace(s, "-", "", -1) //remove the dashes as they will cause an error with hex decode

	b, err := hex.DecodeString(s)

	if err != nil {
		return uuid, err
	}

	return FromBytes(b)

}

// FromBytes will take a in a slice of bytes and attempts to convert into
// a UUID. If bytes does not pass format or is wrong size and error will be returned
func FromBytes(b []byte) (UUID, error) {

	var uuid UUID

	if len(b) != uuidSize {
		return uuid, ErrUUIDSize
	}

	copy(uuid[:], b)

	if !uuidRegex.MatchString(uuid.String()) {
		return uuid, ErrUUIDFormat
	}

	return uuid, nil
}

// Format in bytes 4-2-2-2-6
func (u *UUID) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(u[:4]), hex.EncodeToString(u[4:6]), hex.EncodeToString(u[6:8]), hex.EncodeToString(u[8:10]), hex.EncodeToString(u[10:16]))
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
