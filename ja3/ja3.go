package ja3

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strconv"

	"github.com/dreadl0ck/tlsx"
)

var (
	// Debug indicates whether we run in debug mode.
	Debug        = false
	sepValueByte = byte(45)
	sepFieldByte = byte(44)

	// Ciphers, extensions and elliptic curves, should be filtered so GREASE values are not added to the ja3 digest
	// GREASE RFC: https://tools.ietf.org/html/draft-davidben-tls-grease-01
	greaseValues = map[uint16]bool{
		0x0a0a: true, 0x1a1a: true,
		0x2a2a: true, 0x3a3a: true,
		0x4a4a: true, 0x5a5a: true,
		0x6a6a: true, 0x7a7a: true,
		0x8a8a: true, 0x9a9a: true,
		0xaaaa: true, 0xbaba: true,
		0xcaca: true, 0xdada: true,
		0xeaea: true, 0xfafa: true,
	}
)

// BareToDigestHex converts a bare []byte to a hex string.
func BareToDigestHex(bare []byte) string {
	sum := md5.Sum(bare)
	return hex.EncodeToString(sum[:])
}

func Bare(hello *tlsx.ClientHelloBasic, shouldSort bool) []byte {
	var (
		maxPossibleBufferLength = 5 + 1 + // Version = uint16 => maximum = 65536 = 5chars + 1 field sep
			(5+1)*len(hello.CipherSuites) + // CipherSuite = uint16 => maximum = 65536 = 5chars
			(5+1)*len(hello.AllExtensions) + // uint16 = 2B => maximum = 65536 = 5chars
			(5+1)*len(hello.SupportedGroups) + // uint16 = 2B => maximum = 65536 = 5chars
			(3+1)*len(hello.SupportedPoints) // uint8 = 1B => maximum = 256 = 3chars

		buffer = make([]byte, 0, maxPossibleBufferLength)
	)

	buffer = strconv.AppendInt(buffer, int64(hello.HandshakeVersion), 10)
	buffer = append(buffer, sepFieldByte)

	/*
	 *	Cipher Suites
	 */

	// collect cipher suites
	lastElem := len(hello.CipherSuites) - 1
	if len(hello.CipherSuites) > 1 {
		for _, e := range hello.CipherSuites[:lastElem] {
			// filter GREASE values
			if !greaseValues[uint16(e)] {
				buffer = strconv.AppendInt(buffer, int64(e), 10)
				buffer = append(buffer, sepValueByte)
			}
		}
	}
	// append last element if cipher suites are not empty
	if lastElem != -1 {
		// filter GREASE values
		if !greaseValues[uint16(hello.CipherSuites[lastElem])] {
			buffer = strconv.AppendInt(buffer, int64(hello.CipherSuites[lastElem]), 10)
		}
	}
	buffer = bytes.TrimSuffix(buffer, []byte{sepValueByte})
	buffer = append(buffer, sepFieldByte)

	/*
	 *	Extensions
	 */
	// sort extensions
	if shouldSort {
		sort.Slice(hello.AllExtensions, func(i, j int) bool {
			return hello.AllExtensions[i] < hello.AllExtensions[j]
		})
	}
	// collect extensions
	lastElem = len(hello.AllExtensions) - 1
	if len(hello.AllExtensions) > 1 {
		for _, e := range hello.AllExtensions[:lastElem] {
			// filter GREASE values
			if !greaseValues[uint16(e)] {
				buffer = strconv.AppendInt(buffer, int64(e), 10)
				buffer = append(buffer, sepValueByte)
			}
		}
	}
	// append last element if extensions are not empty
	if lastElem != -1 {
		// filter GREASE values
		if !greaseValues[uint16(hello.AllExtensions[lastElem])] {
			buffer = strconv.AppendInt(buffer, int64(hello.AllExtensions[lastElem]), 10)
		}
	}
	buffer = bytes.TrimSuffix(buffer, []byte{sepValueByte})
	buffer = append(buffer, sepFieldByte)

	/*
	 *	Supported Groups
	 */

	// collect supported groups
	lastElem = len(hello.SupportedGroups) - 1
	if len(hello.SupportedGroups) > 1 {
		for _, e := range hello.SupportedGroups[:lastElem] {
			// filter GREASE values
			if !greaseValues[uint16(e)] {
				buffer = strconv.AppendInt(buffer, int64(e), 10)
				buffer = append(buffer, sepValueByte)
			}
		}
	}
	// append last element if supported groups are not empty
	if lastElem != -1 {
		// filter GREASE values
		if !greaseValues[uint16(hello.SupportedGroups[lastElem])] {
			buffer = strconv.AppendInt(buffer, int64(hello.SupportedGroups[lastElem]), 10)
		}
	}
	buffer = bytes.TrimSuffix(buffer, []byte{sepValueByte})
	buffer = append(buffer, sepFieldByte)

	/*
	 *	Supported Points
	 */

	// collect supported points
	lastElem = len(hello.SupportedPoints) - 1
	if len(hello.SupportedPoints) > 1 {
		for _, e := range hello.SupportedPoints[:lastElem] {
			buffer = strconv.AppendInt(buffer, int64(e), 10)
			buffer = append(buffer, sepValueByte)
		}
	}
	// append last element if supported points are not empty
	if lastElem != -1 {
		buffer = strconv.AppendInt(buffer, int64(hello.SupportedPoints[lastElem]), 10)
	}

	return buffer
}
