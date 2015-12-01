package server

import "fmt"

type InitialCodec struct{}

func (*InitialCodec) ReadPacket(client *Client) *Packet {
	var buf = make([]byte, 1)
	err := client.readFully(buf)
	if err != nil {
		return nil
	}


	switch buf[0] {
	case 15: // Update server
		p := new(Packet)
		data := []byte{1,2,3,4,5,6,7,8}
		p.Write(data)
		client.Write(p)

		client.Codec = NewUpdateCodec()
		return client.Codec.ReadPacket(client)

		case 14: // Initial login
		client.Codec = NewLoginCodec()
		return client.Codec.ReadPacket(client)
	default:
		fmt.Printf("Unknown connection type: %v\n", buf[0])
	}

	return nil
}

func (*InitialCodec) WritePacket(c *Client, p *Packet) {
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