package key

// Key ...
type Key []byte

// String describes the key as a human readable string.
func (k Key) String() string {
	return Printable(k)
}
