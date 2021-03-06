package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
)

// VersionNumber is a version number as int
type VersionNumber uint32

// gQUIC version range as defined in the wiki: https://github.com/quicwg/base-drafts/wiki/QUIC-Versions
const (
	gquicVersion0   = 0x51303030
	maxGquicVersion = 0x51303439
)

// The version numbers, making grepping easier
const (
	VersionTLS      VersionNumber = 0x51474fff
	VersionWhatever VersionNumber = 1 // for when the version doesn't matter
	VersionUnknown  VersionNumber = math.MaxUint32
)

// SupportedVersions lists the versions that the server supports
// must be in sorted descending order
var SupportedVersions = []VersionNumber{VersionTLS}

// IsValidVersion says if the version is known to quic
func IsValidVersion(v VersionNumber) bool {
	return v == VersionTLS || IsSupportedVersion(SupportedVersions, v)
}

func (vn VersionNumber) String() string {
	switch vn {
	case VersionWhatever:
		return "whatever"
	case VersionUnknown:
		return "unknown"
	case VersionTLS:
		return "TLS dev version (WIP)"
	default:
		if vn.isGQUIC() {
			return fmt.Sprintf("gQUIC %d", vn.toGQUICVersion())
		}
		return fmt.Sprintf("%#x", uint32(vn))
	}
}

func (vn VersionNumber) isGQUIC() bool {
	return vn > gquicVersion0 && vn <= maxGquicVersion
}

func (vn VersionNumber) toGQUICVersion() int {
	return int(10*(vn-gquicVersion0)/0x100) + int(vn%0x10)
}

// IsSupportedVersion returns true if the server supports this version
func IsSupportedVersion(supported []VersionNumber, v VersionNumber) bool {
	for _, t := range supported {
		if t == v {
			return true
		}
	}
	return false
}

// ChooseSupportedVersion finds the best version in the overlap of ours and theirs
// ours is a slice of versions that we support, sorted by our preference (descending)
// theirs is a slice of versions offered by the peer. The order does not matter.
// The bool returned indicates if a matching version was found.
func ChooseSupportedVersion(ours, theirs []VersionNumber) (VersionNumber, bool) {
	for _, ourVer := range ours {
		for _, theirVer := range theirs {
			if ourVer == theirVer {
				return ourVer, true
			}
		}
	}
	return 0, false
}

// generateReservedVersion generates a reserved version number (v & 0x0f0f0f0f == 0x0a0a0a0a)
func generateReservedVersion() VersionNumber {
	b := make([]byte, 4)
	_, _ = rand.Read(b) // ignore the error here. Failure to read random data doesn't break anything
	return VersionNumber((binary.BigEndian.Uint32(b) | 0x0a0a0a0a) & 0xfafafafa)
}

// GetGreasedVersions adds one reserved version number to a slice of version numbers, at a random position
func GetGreasedVersions(supported []VersionNumber) []VersionNumber {
	b := make([]byte, 1)
	_, _ = rand.Read(b) // ignore the error here. Failure to read random data doesn't break anything
	randPos := int(b[0]) % (len(supported) + 1)
	greased := make([]VersionNumber, len(supported)+1)
	copy(greased, supported[:randPos])
	greased[randPos] = generateReservedVersion()
	copy(greased[randPos+1:], supported[randPos:])
	return greased
}

// StripGreasedVersions strips all greased versions from a slice of versions
func StripGreasedVersions(versions []VersionNumber) []VersionNumber {
	realVersions := make([]VersionNumber, 0, len(versions))
	for _, v := range versions {
		if v&0x0f0f0f0f != 0x0a0a0a0a {
			realVersions = append(realVersions, v)
		}
	}
	return realVersions
}
