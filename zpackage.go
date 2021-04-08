package vhpackage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ZPackage utility class to read and write in binary format.
// Read binary in Little Endian order.
type ZPackage struct {
	r io.Reader
	w io.Writer
}

func NewZPackageFromData(data []byte) *ZPackage {
	return &ZPackage{
		r: bytes.NewReader(data),
	}
}

func NewZPackageReader(r io.Reader) *ZPackage {
	return &ZPackage{
		r: r,
	}
}

func NewZPackageWriter(w io.Writer) *ZPackage {
	return &ZPackage{
		w: w,
	}
}

func NewZPackage(rw io.ReadWriter) *ZPackage {
	return &ZPackage{
		r: rw,
		w: rw,
	}
}

func (p *ZPackage) ReadZDOID() (ZDOID, error) {
	zdoid := ZDOID{}
	if err := p.read(&zdoid.UserID); err != nil {
		return zdoid, err
	}
	if err := p.read(&zdoid.ID); err != nil {
		return zdoid, err
	}
	return zdoid, nil
}

func (p *ZPackage) ReadBool() (bool, error) {
	var b bool
	return b, p.read(&b)
}

func (p *ZPackage) ReadChar() (uint8, error) {
	var c uint8
	return c, p.read(&c)
}

func (p *ZPackage) ReadByte() (uint8, error) {
	var b uint8
	return b, p.read(&b)
}

func (p *ZPackage) ReadSByte() (int8, error) {
	var b int8
	return b, p.read(&b)
}

func (p *ZPackage) ReadInt() (int, error) {
	var i int32
	if err := p.read(&i); err != nil {
		return 0, err
	}
	return int(i), nil
}

func (p *ZPackage) ReadUInt() (uint, error) {
	var i uint32
	if err := p.read(&i); err != nil {
		return 0, err
	}
	return uint(i), nil
}

func (p *ZPackage) ReadLong() (int64, error) {
	var l int64
	return l, p.read(&l)
}

func (p *ZPackage) ReadULong() (uint64, error) {
	var l uint64
	return l, p.read(&l)
}

func (p *ZPackage) ReadSingle() (float32, error) {
	var s float32
	return s, p.read(&s)
}

func (p *ZPackage) ReadDouble() (float64, error) {
	var d float64
	return d, p.read(&d)
}

func (p *ZPackage) ReadString() (string, error) {
	// String length is encoded as 7 bit at a time.
	// Which means it can be stored in 1 byte up-to 4 bytes.
	var length int

	for nb := 0; nb < 4; nb++ {
		b, err := p.ReadByte()
		if err != nil {
			return "", err
		}

		byt7 := b & 0b01111111
		lastbyte := b & 0b10000000
		length = length | (int(byt7) << (7 * nb))

		if lastbyte == 0 {
			break
		}
	}

	data := make([]byte, length, length)
	if err := p.read(&data); err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadPackage reads a package with a count (int32) as header.
func (p *ZPackage) ReadPackage() (*ZPackage, error) {
	data, err := p.ReadByteArray()
	if err != nil {
		return nil, err
	}

	return NewZPackageFromData(data), nil
}

func (p *ZPackage) ReadByteArray() ([]byte, error) {
	var count int32
	if err := p.read(&count); err != nil {
		return nil, err
	}

	data := make([]byte, count, count)
	if err := p.read(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *ZPackage) ReadVector3() (v Vector3, err error) {
	v.X, err = p.ReadSingle()
	v.Y, err = p.ReadSingle()
	v.Z, err = p.ReadSingle()
	return v, err
}

func (p *ZPackage) ReadVector2i() (Vector2i, error) {
	v := Vector2i{}
	x, err := p.ReadInt()
	if err != nil {
		return v, err
	}
	y, err := p.ReadInt()
	if err != nil {
		return v, err
	}
	v.X = int32(x)
	v.Y = int32(y)
	return v, nil
}

func (p *ZPackage) ReadQuaternion() (q Quaternion, err error) {
	q.X, err = p.ReadSingle()
	q.Y, err = p.ReadSingle()
	q.Z, err = p.ReadSingle()
	q.W, err = p.ReadSingle()
	return q, err
}

func (p *ZPackage) ReadIntoList(l interface{}) error {
	count, err := p.ReadInt()
	if err != nil {
		return err
	}

	switch x := l.(type) {
	case (*[]string):
		*x = make([]string, count)
		for i := 0; i < count; i++ {
			(*x)[i], err = p.ReadString()
			if err != nil {
				return err
			}
		}

	case (*[]int):
		int32s := make([]int32, count)
		if err := p.read(&int32s); err != nil {
			return err
		}
		*x = make([]int, count)
		for i, v := range int32s {
			(*x)[i] = int(v)
		}

	default:
		return fmt.Errorf("cannot read into list of type %T", l)
	}

	return nil
}

func (p *ZPackage) read(data interface{}) error {
	return binary.Read(p.r, binary.LittleEndian, data)
}

func (p *ZPackage) WriteByte(b uint8) error {
	return p.write(b)
}

func (p *ZPackage) WriteInt(n int) error {
	return p.write(int32(n))
}

func (p *ZPackage) WriteSingle(f float32) error {
	return p.write(f)
}

func (p *ZPackage) WriteBool(b bool) error {
	return p.write(b)
}

func (p *ZPackage) WriteString(s string) error {
	v := (uint)(len(s))
	for v >= 0x80 {
		p.write((byte)(v | 0x80))
		v >>= 7
	}
	if err := p.write((byte)(v)); err != nil {
		return err
	}
	return p.write([]byte(s))
}

func (p *ZPackage) write(v interface{}) error {
	return binary.Write(p.w, binary.LittleEndian, v)
}
