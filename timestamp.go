package uuid

import (
	"math/rand"
	"os/user"
	"strconv"
	"time"
)

const (
	epochOffset = 122192928000000000 // See uuidTime below
)

// Timestamp https://tools.ietf.org/html/rfc4122#section-4.1.4 and https://tools.ietf.org/html/rfc4122#section-4.1.2
// The timestamp is a 60-bit value: so why are we returning 64?
// The timestamp is a 64 bit value that its last byte is multiplexed with version number (i.e. 1-5)
type timestamp interface {
	timestamp() uint64
}

func getUUIDEpochTime() uint64 {
	return uint64((time.Now().UnixNano() + epochOffset) / 100) // 100 nano second intervals
}

// V1
// From Doc: For UUID version 1, this is represented by Coordinated Universal Time (UTC)
// as a count of 100-nanosecond intervals since 00:00:00.00, 15 October 1582 (the date of
// Gregorian reform to the Christian calendar). This is date requires and offset between
// unix epoch time and and uuid epoch time: thus the epochOffset above (see const)
type uuidTime struct{}

func (u *uuidTime) timestamp() uint64 {
	return getUUIDEpochTime()
}

// V2 is similiar to V1, but with some a couple of differences. First, v2 does not fall under
// RFC4122. Instead it is defined by DCE1.1 (http://pubs.opengroup.org/onlinepubs/9629399/apdxa.htm)
// V2 does take a timestamp but the time_low is to be replaced by UID or GID (atm UID is being used)
type uuidDCE struct{}

func (u *uuidDCE) timestamp() uint64 {
	t := getUUIDEpochTime()
	uId := getUser()
	return t ^ (0xFFFFFFFF) | uint64(uId)
}

//To DO: handle panics
func getUser() int {

	us, err := user.Current()

	if err != nil {
		panic(err)
	}

	i, err := strconv.Atoi(us.Uid)

	if err != nil {
		panic(err)
	}

	return i
}

//V4
// For UUID version 4, the timestamp is a randomly or pseudo-randomly
// generated 60-bit value, as described in https://tools.ietf.org/html/rfc4122#section-4.4 Section 4.4.
type uuidRand struct{}

func (u *uuidRand) timestamp() uint64 {
	rand.Seed(time.Now().UnixNano())
	return uint64(rand.Int63())
}
