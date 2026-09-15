package kdf

// Zeroize clears a sensitive byte slice by overwriting it with zeros.
func Zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
