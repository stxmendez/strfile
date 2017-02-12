package strfile

import (
	"encoding/binary"
	"errors"
	"os"
)

const (
	VERSION      = 1
	FLAG_RANDOM  = 1
	FLAG_ORDERED = 2
	FLAG_ROTATED = 3
)

// Header is the data encoded into an strfile.  Note that fields are aligned on 8 byte boundaries when written to
// disk.
type Header struct {
	Version  uint32
	Numstr   uint32
	LongLen  uint32
	ShortLen uint32
	Flags    uint32
	Delim    byte
}

type StrFileReader struct {
	idxFile, strFile string
	idxStat, strStat os.FileInfo
	header           *Header
}

const (
	headerLength     = 48 // sizeof(Header) plus a three byte pad
	indexEntryLength = 8  // Each entry is an 8 byte integer
)

var (
	ErrIsDirectory = errors.New("expected a file, not a directory")
	ErrUnexpected  = errors.New("unexpected error")
)

func checkFile(path string) (os.FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, ErrIsDirectory
	}

	return stat, nil
}

func readUint32(f *os.File) (uint32, error) {
	buf := make([]byte, 4)
	n, err := f.Read(buf)
	if err != nil {
		return 0, err
	}

	if n != cap(buf) {
		return 0, ErrUnexpected
	}

	return binary.BigEndian.Uint32(buf), nil
}

func (f *StrFileReader) Header() (*Header, error) {
	if f.header == nil {
		file, err := os.Open(f.idxFile)
		if err != nil {
			return nil, err
		}

		version, err := readUint32(file)
		if err != nil {
			return nil, err
		}

		// Skip pad
		_, err = readUint32(file)
		if err != nil {
			return nil, err
		}

		numstr, err := readUint32(file)
		if err != nil {
			return nil, err
		}

		// Skip pad
		_, err = readUint32(file)
		if err != nil {
			return nil, err
		}

		longlen, err := readUint32(file)
		if err != nil {
			return nil, err
		}

		// Skip pad
		_, err = readUint32(file)
		if err != nil {
			return nil, err
		}

		shortlen, err := readUint32(file)
		if err != nil {
			return nil, err
		}

		// Skip pad
		_, err = readUint32(file)
		if err != nil {
			return nil, err
		}

		flags, err := readUint32(file)
		if err != nil {
			return nil, err
		}

		// Skip pad
		_, err = readUint32(file)
		if err != nil {
			return nil, err
		}

		b := make([]byte, 1)
		n, err := file.Read(b)
		if err != nil {
			return nil, err
		}

		if n != 1 {
			return nil, ErrUnexpected
		}

		f.header = &Header{
			Delim:    b[0],
			Flags:    flags,
			LongLen:  longlen,
			ShortLen: shortlen,
			Numstr:   numstr,
			Version:  version,
		}
	}

	return f.header, nil
}

func NewStrFileReader(strFile, idxFile string) (*StrFileReader, error) {
	strStat, err := checkFile(strFile)
	if err != nil {
		return nil, err
	}

	idxStat, err := checkFile(idxFile)
	if err != nil {
		return nil, err
	}

	return &StrFileReader{idxFile: idxFile, idxStat: idxStat, strFile: strFile, strStat: strStat}, nil
}

// String returns the string that associated with the index entry ordinal value.
func (f *StrFileReader) String(idx int) (string, error) {
	idxFile, err := os.Open(f.idxFile)
	if err != nil {
		return "", nil
	}
	defer idxFile.Close()

	index_offset := int64((idx * indexEntryLength) + headerLength)
	n, err := idxFile.Seek(index_offset, 0)
	if err != nil {
		return "", err
	}

	if n != index_offset {
		return "", ErrUnexpected
	}

	// Read the quote offset start
	quote_start, err := readUint32(idxFile)
	if err != nil {
		return "", err
	}

	// Skip the lenght field, it is unused
	_, err = readUint32(idxFile)
	if err != nil {
		return "", err
	}

	var quote_end uint32
	if index_offset+indexEntryLength >= f.idxStat.Size() {
		quote_end = uint32(f.strStat.Size())
	} else {
		end, err := readUint32(idxFile)
		if err != nil {
			return "", err
		}

		// Strings are delimited by a '%' character followed by a newline.
		quote_end = end - 2
	}

	buf := make([]byte, quote_end-quote_start)
	strFile, err := os.Open(f.strFile)
	if err != nil {
		return "", err
	}
	defer strFile.Close()

	n, err = strFile.Seek(int64(quote_start), 0)
	if err != nil {
		return "", err
	}

	if n != int64(quote_start) {
		return "", ErrUnexpected
	}

	nBytes, err := strFile.Read(buf)
	if err != nil {
		return "", err
	}

	if nBytes != len(buf) {
		return "", ErrUnexpected
	}

	return string(buf), nil
}

func (f *StrFileReader) StringCount() (uint32, error) {
	h, err := f.Header()
	if err != nil {
		return 0, err
	}

	return h.Numstr, nil
}
