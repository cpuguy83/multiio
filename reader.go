package multiio

import (
	"io"
)

// SizedReaderAt is a ReaderAt that also implements a function to get its size.
type SizedReaderAt interface {
	Size() int64
	io.ReaderAt
}

// MultiReader is a reader which implements various io reader functions for
// but spans across multiple readers.
//
// It is undefined what will occur if the underlying readers are mutated.
// The expectation should be that you'll see data corruption when this happens.
// If the final reader is appeneded to it's probably ok, but still undefined.
type MultiReader struct {
	r1, r2 SizedReaderAt
	pos    int64
}

type nullReader struct{}

func (nullReader) Size() int64 {
	return 0
}

func (nullReader) ReadAt(_ []byte, _ int64) (int, error) {
	return 0, io.EOF
}

// NewMultiReader creates a new Reader that is the logical concatencation of
// the passed in readeers.
func NewMultiReader(readers ...SizedReaderAt) *MultiReader {
	if len(readers) == 0 {
		return nil
	}
	if len(readers) == 1 {
		return &MultiReader{r1: readers[0], r2: nullReader{}}
	}

	rdr := &MultiReader{r1: readers[0], r2: nullReader{}}
	for _, r := range readers[1:] {
		rdr = &MultiReader{r1: rdr, r2: r}
	}
	return rdr
}

// Size gets the size of the data in all the readsers reader
func (r *MultiReader) Size() int64 {
	return r.r1.Size() + r.r2.Size()
}

// ReadAt implements io.ReaderAt but spans across multiple readers
func (r *MultiReader) ReadAt(p []byte, offset int64) (n int, err error) {
	r1Size := r.r1.Size()
	r2Size := r.r2.Size()

	if offset > r1Size+r2Size {
		return 0, errOutOfRange{}
	}

	if offset < r1Size {
		n1, err := r.r1.ReadAt(p, offset)
		if err == nil || err != io.EOF {
			return n1, err
		}
		n2, err := r.r2.ReadAt(p[n1:], 0)
		return n1 + n2, err
	}

	offset -= r1Size
	return r.r2.ReadAt(p, offset)
}

// Seek implements io.Seeker but spans across multiple readers.
func (r *MultiReader) Seek(offset int64, whence int) (int64, error) {
	r1Size := r.r1.Size()
	r2Size := r.r2.Size()

	if offset > r1Size+r2Size {
		return -1, errOutOfRange{}
	}

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return -1, errOutOfRange{}
		}
		r.pos = offset
	case io.SeekCurrent:
		newPos := r.pos + offset
		if newPos < 0 || newPos > r1Size+r2Size {
			return -1, errOutOfRange{}
		}
		r.pos = newPos
	case io.SeekEnd:
		newPos := r1Size + r2Size + offset
		if offset > 0 || newPos < 0 {
			return -1, errOutOfRange{}
		}
		r.pos = newPos
	default:
		return -1, errOutOfRange{}
	}
	return r.pos, nil
}

// Read implements io.Reader across multiple readers.
func (r *MultiReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadAt(p, r.pos)
	r.pos += int64(n)
	return n, err
}
