package multiio

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReaderAt(t *testing.T) {
	testData1 := "This is a test, "
	testData2 := "this is only"
	testData3 := " a test."
	r1 := strings.NewReader(testData1)
	r2 := strings.NewReader(testData2)
	r3 := strings.NewReader(testData3)

	rdr := NewMultiReader(r1, r2, r3)

	type testParams struct {
		offset int64
		size   int
		expect string
	}

	cases := map[string]testParams{
		"Full":               {offset: 0, size: len(testData1) + len(testData2) + len(testData3), expect: testData1 + testData2 + testData3},
		"SpanFirstAndSecond": {offset: 7, size: 10, expect: " a test, t"},
		"SpanAll":            {offset: int64(len(testData1) - 1), size: len(testData2) + len(testData3), expect: " this is only a test"}, // 1 char from r1 through 1 char less than r3
		"NoSpan r1":          {offset: 0, size: len(testData1), expect: "This is a test, "},
		"NoSpan r2":          {offset: int64(len(testData1)), size: len(testData2), expect: "this is only"},
		"NoSpan r3":          {offset: int64(len(testData1) + len(testData2)), size: len(testData3), expect: " a test."},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {

			buf := make([]byte, c.size)
			n, err := rdr.ReadAt(buf, c.offset)
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}

			if n < c.size {
				t.Fatalf("short read, expected %d, got: %d", c.size, n)
			}

			if string(buf[:n]) != c.expect {
				t.Fatalf("expected %q, got: %s", c.expect, string(buf[:n]))
			}
		})
	}
}

func TestSeek(t *testing.T) {
	testData1 := "This is a test, "
	testData2 := "this is only"
	testData3 := " a test."
	r1 := strings.NewReader(testData1)
	r2 := strings.NewReader(testData2)
	r3 := strings.NewReader(testData3)

	type testParams struct {
		offset     int64
		whence     int
		expectPos  int64
		expectData []byte // just enough info to know we are in the right place...
		expectErr  error
	}

	cases := map[string]testParams{
		"SeekStart0":      {offset: 0, whence: io.SeekStart, expectPos: 0, expectData: []byte("This is a test,")},
		"SeekStartOffset": {offset: 10, whence: io.SeekStart, expectPos: 10, expectData: []byte("test, this is only")},
		"SeekStartBefore": {offset: -1, whence: io.SeekStart, expectPos: -1, expectErr: errOutOfRange{}},
		"SeekStartTooBig": {offset: int64(len(testData1)+len(testData2)+len(testData3)) + 1, whence: io.SeekStart, expectPos: -1, expectErr: errOutOfRange{}},
		"SeekEnd0":        {offset: 0, whence: io.SeekEnd, expectPos: int64(len(testData1) + len(testData2) + len(testData3)), expectData: []byte("")},
		"SeekEndOffset":   {offset: -10, whence: io.SeekEnd, expectPos: 26, expectData: []byte("ly a test.")},
		"SeekEndtAfter":   {offset: 1, whence: io.SeekEnd, expectPos: -1, expectErr: errOutOfRange{}},
		"SeekEndtTooBig":  {offset: (-1 * int64(len(testData1)+len(testData2)+len(testData3))) - 1, whence: io.SeekEnd, expectPos: -1, expectErr: errOutOfRange{}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			rdr := NewMultiReader(r1, r2, r3)

			n, err := rdr.Seek(c.offset, c.whence)
			if err != c.expectErr {
				t.Fatal(err)
			}

			if n != c.expectPos {
				t.Fatalf("expected position %d, got: %d", c.expectPos, n)
			}
			cur, err := rdr.Seek(0, io.SeekCurrent)
			if err != nil {
				t.Fatal(err)
			}
			if n == -1 {
				// SeekCurrent will not return -1 unless there is an error, which there shouldn't be
				// So set this to 0 for comparison
				n = 0
			}
			if cur != n {
				t.Fatalf("exected current position %d, got: %d", n, cur)
			}

			buf := make([]byte, len(c.expectData))
			nr, err := rdr.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}
			if err == io.EOF && rdr.pos != int64(len(testData1)+len(testData2)+len(testData3)) {
				t.Fatal("expected EOF")
			}
			if !bytes.Equal(buf[:nr], c.expectData) {
				t.Fatalf("expected to read %q, got: %s", string(c.expectData), buf[:nr])
			}
		})
	}
}
