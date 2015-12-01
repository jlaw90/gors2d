package server

import (
	"math"
	"io"
)

type PacketCodec interface {
	ReadPacket(client *Client) *Packet

	WritePacket(*Client, *Packet)
}

type Packet struct {
	Data  []byte
	Position   uint
	Limit uint
}

func (p *Packet) checkBounds(end int) (error) {
	cap := len(p.Data)
	if end < 0 || end >= int(p.Limit) || end >= cap {
		return io.EOF
	}
	return nil
}

func (p *Packet) ensureBounds(end uint) {
	capacity := uint(len(p.Data))
	if end >= capacity {
		d := int(end - capacity)
		for i := 0; i < d; i++ {
			p.Data = append(p.Data, 0)
		}
	}
}

func (p *Packet) Clear() {
	p.Limit = 0
	p.Position = 0
	p.Data = p.Data[:0]
}

func (p *Packet) Read(dest []byte) (n int, err error) {
	n, err = p.ReadAt(dest, int64(p.Position))
	p.Position += uint(n)

	return n, err
}

func (p *Packet) ReadAt(dest []byte, off int64) (n int, err error) {
	of32 := uint(off)
	end := uint(math.Min(float64(int(off) + len(dest)), float64(len(p.Data))))

	if p.Position == end {
		return 0, io.EOF
	}

	var i uint = 0
	for ; of32 < end; i++ {
		dest[i] = p.Data[of32+i]
	}
	n = int(i)

	return
}

func (p *Packet) Write(data []byte) (n int, err error) {
	n, err = p.WriteAt(data, int64(p.Position))
	if p.Limit == p.Position {
		p.Limit += uint(n)
	}
	p.Position += uint(n)

	return
}

func (p *Packet) WriteAt(data []byte, off int64) (n int, err error) {
	of32 := uint(off)
	l := len(data)
	p.ensureBounds(of32 + uint(l))

	for _, b := range data {
		p.Data[of32] = b
		of32 += 1
	}

	return l, nil
}

func (p *Packet) Uint8() (uint8, error) {
	if err := p.checkBounds(int(p.Position)); err != nil {
		return 0, err
	}

	b := p.Data[p.Position]
	p.Position += 1
	return b, nil
}

func (p *Packet) Uint8At(pos int) (uint8, error) {
	if err := p.checkBounds(pos); err != nil {
		return 0, err
	}

	return p.Data[pos], nil
}

func (p *Packet) WriteUint8(v uint8) {
	p.ensureBounds(p.Position)
	p.Data[p.Position] = v
	if p.Limit == p.Position {
		p.Limit += 1
	}
	p.Position += 1
}