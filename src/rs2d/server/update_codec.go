package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"rs2d/cache"
	"time"
)

type UpdateCodec struct {
	client *Client
	requests chan *request
	received chan *received
}

func NewUpdateCodec() (*UpdateCodec) {
	return &UpdateCodec {
		requests: make(chan *request, 20),
		received: make(chan *received, 20),
	}
}

type request struct {
	Cache uint8
	Entry uint16
	Priority uint8
}

type received struct {
	request
	Data []byte
}

func (uc *UpdateCodec) findEntry(req *request) {
	r := new(received)
	r.Cache = req.Cache
	r.Entry = req.Entry
	r.Priority = req.Priority
	r.Data = cache.ReadEntry(int(req.Cache + 2), int(req.Entry))
	uc.received <- r
}

func (uc *UpdateCodec) updateProcess() {
	timeout := make(chan bool, 1)

	p := new(Packet)
	for uc.client.Connected {

		go func() {
			time.Sleep(10 * time.Second)
			timeout <- true
		}()

		select {
		case req := <- uc.requests:
			uc.findEntry(req)
		case rec := <- uc.received:
			data := rec.Data
			if data == nil {
				// Respond with empty
				p := new(Packet)
				p.Write([]byte{
					rec.Cache,
					byte((rec.Entry >> 8) & 0xff),
					byte((rec.Entry >> 0) & 0xff),
					0,
					0,
					0,
				})
				uc.client.Write(p)
				continue
			}
			//data = append(data, byte((ver >> 8) & 0xff), byte((ver >> 0) & 0xff))

			size := len(data)
			//fmt.Printf("Version for [%v][%v]: %v\n", req.Cache, req.Entry, ver)
			header := []byte {
				rec.Cache,
				byte((rec.Entry >> 8) & 0xff),
				byte((rec.Entry >> 0) & 0xff),
				byte((size >> 8) & 0xff),
				byte((size >> 0) & 0xff),
				0,
			}
			block := 0
			for len(data) > 0 {
				header[5] = byte(block)
				block += 1

				p.Clear()
				p.Write(header)
				end := len(data)
				if end > 500 {
					end = 500
				}

				p.Write(data[:end])
				data = data[end:]
				uc.client.Write(p)
			}
		case <-timeout:
			continue

		default:
			time.Sleep(10)
		}
	}
}

func (uc *UpdateCodec) ReadPacket(client *Client) *Packet {
	var in [4]byte;

	err := client.readFully(in[0:4])
	if err != nil {
		return nil
	}

	if uc.client == nil {
		uc.client = client
		go uc.updateProcess()
	}

	reader := bytes.NewReader(in[0:4])

	req := new(request)
	err = binary.Read(reader, binary.BigEndian, req)
	if err != nil {
		client.Disconnect(fmt.Sprintf("Error reading update packet: %v", err))
		return nil
	}

	if req.Priority == 10 {
		return nil // PING, ignore
	}

	uc.requests <- req

	return nil
}

func (*UpdateCodec) WritePacket(c *Client, p *Packet) {
	// Send raw...
	buf := p.Data[0:p.Limit]
	for len(buf) > 0 {
		n, err := (*c.Connection).Write(buf)
		buf = buf[n:]
		if err != nil {
			c.Disconnect(fmt.Sprintf("Error writing to client: %v", err))
			return
		}
	}
}