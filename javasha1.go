package proxy

import (
    "crypto/sha1"
    "hash"
    "fmt"
    "strings"
)

func AuthDigest(serverID string, sharedSecret []byte, encodedPublicKey []byte) string {
    h := sha1.New()
    h.Write([]byte(serverID))
    h.Write(sharedSecret)
    h.Write(encodedPublicKey)
    return JavaDigest(h)
}

// JavaDigest computes a special SHA-1 digest required for Minecraft web
// authentication on Premium servers (online-mode=true).
// Source: http://wiki.vg/Protocol_Encryption#Server
//
// Also many, many thanks to SirCmpwn and his wonderful gist (C#):
// https://gist.github.com/SirCmpwn/404223052379e82f91e6
func JavaDigest(h hash.Hash) string {
    hash := h.Sum(nil)
 
    // Check for negative hashes
    negative := (hash[0] & 0x80) == 0x80
    if negative {
        hash = twosComplement(hash)
    }
 
    // Trim away zeroes
    res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
    if negative {
        res = "-" + res
    }
 
    return res
}
 
// little endian
func twosComplement(p []byte) []byte {
    carry := true
    for i := len(p) - 1; i >= 0; i-- {
        p[i] = byte(^p[i])
        if carry {
            carry = p[i] == 0xff
            p[i]++
        }
    }
    return p
}
