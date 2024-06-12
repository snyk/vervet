// Package storage provides common functionality supporting Vervet Underground
// storage.
package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// Digest is a sha256 calculation generated
// into a specific string format prefixed by "sha256:" followed by
// the sha256 value generated.
type Digest string

// DigestHeader returns a content digest parsed from a Digest HTTP response
// header as defined in
// https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-digest-headers-05#section-3.
// The returned digest is algorithm-prefixed so that other digest schemes may
// be supported later if needed.
//
// Returns "" if no digest is available.
func DigestHeader(value string) string {
	digests := strings.Split(value, ",")
	for i := range digests {
		digests[i] = strings.TrimSpace(digests[i])
		kv := strings.SplitN(digests[i], "=", 2)
		if len(kv) < 2 {
			continue
		}
		if kv[0] == "id-sha-256" || kv[0] == "sha-256" {
			// Use the no-encoding digest if specified, otherwise assume no
			// encoding as a fallback. HTTP compression is likely to be handled
			// transparently.
			return "sha256:" + kv[1]
		}
	}
	return ""
}

// NewDigest returns the digest of the given contents.
func NewDigest(contents []byte) Digest {
	buf := sha256.Sum256(contents)
	return Digest("sha256:" + base64.StdEncoding.EncodeToString(buf[:]))
}
