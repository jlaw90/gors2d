package server

import (
	"net"
	"fmt"
)

type Client struct {
	Connected bool
	Connection *net.Conn
	Codec PacketCodec


}

func (c *Client) Disconnect(reason string) {
	c.Connected = false
	fmt.Printf("Client %v disconnected: %v\n", c, reason)
}

func (c *Client) readFully(dest []byte) (err error) {
	for len(dest) > 0 {
		var n int
		n, err = (*c.Connection).Read(dest)
		if err != nil {
			c.Disconnect(fmt.Sprintf("Read error: %v", err))
			return
		}
		dest = dest[n:]
	}
	return
}

func (c *Client) Write(packet *Packet) {
	c.Codec.WritePacket(c, packet)
}

func (c *Client) String() string {
	return (*c.Connection).RemoteAddr().String()
}

func NewClient(conn *net.Conn) *Client {
	c := &Client{
		Connection: conn,
		Codec: new(InitialCodec),
		Connected: true,
	}

	return c
}