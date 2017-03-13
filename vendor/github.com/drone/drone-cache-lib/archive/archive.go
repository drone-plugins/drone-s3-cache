package archive

import (
	"io"
)

// Archive is an interface for packing and unpacking archive formats.
type Archive interface {
	// Pack writes an archive containing the source
	Pack(srcs []string, w io.Writer) error

	// Unpack reads the archive and restores it to the destination
	Unpack(dst string, r io.Reader) error
}
