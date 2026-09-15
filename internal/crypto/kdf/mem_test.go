package kdf

import "testing"

func TestZeroize(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	Zeroize(data)
	for i, b := range data {
		if b != 0 {
			t.Fatalf("expected byte at index %d to be 0, got %d", i, b)
		}
	}
}

func TestZeroizeEmpty(t *testing.T) {
	var data []byte
	Zeroize(data) // should not panic
}
