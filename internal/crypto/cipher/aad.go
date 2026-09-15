package cipher

import (
	"fmt"
)

// AADContext represents contextual authenticated metadata bound to ciphertexts to prevent transplant attacks.
type AADContext struct {
	VaultID    string
	RecordID   string
	RecordType string
	Version    uint32
}

// Bytes returns the deterministic canonical byte representation of the AAD context.
func (c AADContext) Bytes() []byte {
	return []byte(fmt.Sprintf("govault/v1|v:%s|r:%s|t:%s|ver:%d", c.VaultID, c.RecordID, c.RecordType, c.Version))
}

// SimpleAAD formats an arbitrary label into standard domain-separated AAD bytes.
func SimpleAAD(label string) []byte {
	return []byte(fmt.Sprintf("govault/v1|label:%s", label))
}
