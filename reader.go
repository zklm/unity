package unity

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

type Reader struct {
	buf    []byte
	off    int64
	endian binary.ByteOrder
}

func NewReader(b []byte) (*Reader, error) {
	return &Reader{b, 0, binary.BigEndian}, nil
}

func NewReaderFromFilePath(path string) (*Reader, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Reader{b, 0, binary.BigEndian}, nil
}

func (r *Reader) ChangeEndian(isLittleEndian bool) {
	if isLittleEndian {
		r.endian = binary.LittleEndian
	} else {
		r.endian = binary.BigEndian
	}
}

func (r *Reader) Endian() binary.ByteOrder {
	return r.endian
}

func (r *Reader) Reset() {
	r.buf = r.buf[:0]
	r.off = 0
}

func (r *Reader) Renew() {
	r.buf = r.buf[r.off:]
	r.off = 0
}

func (r *Reader) Peek(n int64) ([]byte, error) {
	if r.HasSpace(n) {
		return r.buf[r.off : r.off+n], nil
	}
	return nil, io.EOF
}

func (r *Reader) IsEmpty() bool {
	return r.Len() <= r.off
}

func (r *Reader) HasSpace(sz int64) bool {
	return r.Len() >= r.off+sz
}

func (r *Reader) Len() int64 {
	return int64(len(r.buf))
}

func (r *Reader) Tell() int64 {
	return r.off
}

func (r *Reader) Remaining() int64 {
	return int64(len(r.buf[r.off:]))
}

func (r *Reader) Align() (int64, error) {
	oldPos := r.Len() - r.Remaining()
	newPos := (oldPos + 3) & -4
	if newPos > oldPos {
		return r.SeekStart(int64(newPos - oldPos))
	}
	return r.off, nil
}

func (r *Reader) SeekStart(offset int64) (int64, error) {
	if offset > r.Len() {
		return 0, io.EOF
	} else if offset < 0 {
		return 0, fmt.Errorf("Reader.Seek: invalid offset (%v)", offset)
	}
	r.off = offset
	return r.off, nil
}

func (r *Reader) SeekCurrent(offset int64) (int64, error) {
	if offset > r.Remaining() {
		return 0, io.EOF
	} else if r.off+offset < 0 {
		return 0, fmt.Errorf("Reader.Seek: invalid offset (%v)", offset)
	}
	r.off += offset
	return r.off, nil
}

func (r *Reader) Byte() (byte, error) {
	if r.IsEmpty() {
		return 0, io.EOF
	}
	b := r.buf[r.off]
	r.off++
	return b, nil
}

func (r *Reader) Bytes(n int64) ([]byte, error) {
	if !r.HasSpace(n) {
		return nil, io.EOF
	}
	end := r.off + n
	b := r.buf[r.off:end]
	r.off = end
	return b, nil
}

func (r *Reader) Uint8() (uint8, error) {
	return r.Byte()
}

func (r *Reader) Int8() (int8, error) {
	if r.IsEmpty() {
		return 0, io.EOF
	}
	x := int8(r.buf[r.off])
	r.off++
	return x, nil
}

func (r *Reader) Uint16() (uint16, error) {
	if !r.HasSpace(2) {
		return 0, io.EOF
	}
	x := r.endian.Uint16(r.buf[r.off : r.off+2])
	r.off += 2
	return x, nil
}

func (r *Reader) Int16() (int16, error) {
	if !r.HasSpace(2) {
		return 0, io.EOF
	}
	x := int16(r.endian.Uint16(r.buf[r.off : r.off+2]))
	r.off += 2
	return x, nil

}

func (r *Reader) Uint32() (uint32, error) {
	if !r.HasSpace(4) {
		return 0, io.EOF
	}
	x := r.endian.Uint32(r.buf[r.off : r.off+4])
	r.off += 4
	return x, nil
}

func (r *Reader) Int32() (int32, error) {
	if !r.HasSpace(4) {
		return 0, io.EOF
	}
	x := int32(r.endian.Uint32(r.buf[r.off : r.off+4]))
	r.off += 4
	return x, nil
}

func (r *Reader) Uint64() (uint64, error) {
	if !r.HasSpace(8) {
		return 0, io.EOF
	}
	x := r.endian.Uint64(r.buf[r.off : r.off+8])
	r.off += 8
	return x, nil
}

func (r *Reader) Int64() (int64, error) {
	if !r.HasSpace(8) {
		return 0, io.EOF
	}
	x := int64(r.endian.Uint64(r.buf[r.off : r.off+8]))
	r.off += 8
	return x, nil
}

// https://github.com/andlabs/ohv/blob/master/mshi/reader.go#L140
func (r *Reader) SevenBitEncodedInt() (uint32, error) {
	var out, shift uint32
	for {
		b, err := r.Byte()
		if err != nil {
			return 0, err
		}

		out |= uint32(b&0x7F) << shift
		shift += 7
		if b&0x80 == 0 {
			break
		}
	}
	return out, nil
}

// Reads the data buffer until null byte is found and converts to string
func (r *Reader) StringNull() (string, error) {
	b := []byte{}
	for {
		if c, err := r.Byte(); err != nil {
			return "", err
		} else if c != 0 {
			b = append(b, c)
		} else {
			break
		}
	}
	return string(b), nil
}

// Reads the data buffer by specified limit and converts to string
func (r *Reader) StringLimited(limit int64) (string, error) {
	b := []byte{}
	remain := r.Remaining()

	if remain < limit {
		limit = remain
	}

	for i := int64(0); i < limit; i++ {
		if c, err := r.Byte(); err != nil {
			return "", err
		} else {
			b = append(b, c)
		}
	}
	return string(b), nil
}

func (r *Reader) StringPrefixed() (string, error) {
	len, err := r.SevenBitEncodedInt()
	if err != nil {
		return "", err
	}

	if len == 0 {
		return "", nil
	}

	return r.StringLimited(int64(len))
}
