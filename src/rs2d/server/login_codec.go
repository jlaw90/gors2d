package server

import (
	"fmt"
	"encoding/binary"
	"errors"
	"rs2d/player/login"
	"math/rand"
	"io"
	"bytes"
	"rs2d/cache"
	"encoding/hex"
	"rs2d/config"
	"math/big"
	"rs2d/player"
)

type LoginCodec struct{

}

func NewLoginCodec() *LoginCodec {
	return &LoginCodec {
	}
}

func (lc *LoginCodec) ReadPacket(client *Client) *Packet {
	resp := new(Packet)

	respondWith := func(code login.LoginResponseCode, err error) *Packet {
		resp.Clear()
		b := byte(code)
		resp.Write([]byte{b,b,b,b,b,b,b,b,b}) // incase we're in auth stage, we need 8 bytes for session key
		client.Write(resp)
		client.Disconnect(fmt.Sprintf("Login failed: %v\n", err))
		return nil
	}

	// Read initial login request
	var hash uint8 // 5 bit representation of users encoded name
	var err error
	if err = binary.Read(*client.Connection, binary.BigEndian, &hash); err != nil {
		return respondWith(login.Fail, err)
	}


	resp.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8}) // 8 dummy bytes ignored by client
	resp.Write([]byte{byte(login.Begin)}) // Request client send login info

	// Respond to initial login request - can reply with maintenance or general error messages here before authentication stage
	client.Write(resp)

	resp.Clear()

	// Authentication
	serverSessionKey := rand.Int63()
	if err = binary.Write(resp, binary.BigEndian, serverSessionKey); err != nil {
		return respondWith(login.Fail, err)
	}
	client.Write(resp)
	resp.Clear()

	var flagsize uint16

	if err = binary.Read(*client.Connection, binary.BigEndian, &flagsize); err != nil {
		return respondWith(login.Fail, err)
	}
	reconnecting := (flagsize >> 8) == 18
	if !reconnecting && ((flagsize >> 8) != 16) {
		return respondWith(login.Fail, errors.New(fmt.Sprintf("unexpected login type: %v, expecting 16 or 18", (flagsize>>8))))
	}

	packetSize := (flagsize & 0xff)

	data := make([]byte, packetSize) // packetSize includes itself...
	_, err = io.ReadFull(*client.Connection, data)
	if err != nil {
		return respondWith(login.Fail, err)
	}

	reader := bytes.NewReader(data)

	var header int32

	if err = binary.Read(reader, binary.BigEndian, &header); err != nil {
		return respondWith(login.Fail, err)
	}

	if ((header >> 24) & 0xff) != 0xff {
		return respondWith(login.Fail, errors.New("invalid start of authentication packet"))
	}

	version := (header >> 8) & 0xffff
	if version != 317 {
		return respondWith(login.Updated, errors.New(fmt.Sprintf("invalid client version: %v", version)))
	}

	_ = (header & 0xff) == 1 // lowmem

	var titleCrc int32
	for i := 0; i < 9; i++ {
		if err = binary.Read(reader, binary.BigEndian, &titleCrc); err != nil {
			return respondWith(login.Fail, err)
		}
		if titleCrc != cache.TitleCrc[i] {
			return respondWith(login.Fail, errors.New("title CRC mismatch"))
		}
	}

	// The rest of the buffer is the login packet!

	rsaLen, _ := reader.ReadByte()

	data = data[41:]

	if int(rsaLen) != len(data) {
		return respondWith(login.Fail, errors.New(fmt.Sprintf("RSA length mismatch - real: %v, reported: %v", len(data), rsaLen)))
	}

	// Decrypt RSA portion of login block
	if config.Config.UseRSAEncryption {
		priv := config.Config.PrivateKey
		c := new(big.Int).SetBytes(data);
		m := new(big.Int).Exp(c, priv.D, priv.N);
		data = m.Bytes()
	}

	buf := bytes.NewBuffer(data) // We use buffer here so we can use ReadString

	if check, err := buf.ReadByte(); err != nil || check != 10 {
		fmt.Printf("%v\n", hex.EncodeToString(data))
		if err == nil {
			err = errors.New(fmt.Sprintf("invalid decoded RSA block start, got %v", data[0]))
		}
		return respondWith(login.Fail, err)
	}

	var clientSessionKey, reportedServerSessionKey int64
	var uid int32
	var username, password string

	if err = binary.Read(buf, binary.BigEndian, &clientSessionKey); err != nil {
		return respondWith(login.Fail, err)
	}

	if err = binary.Read(buf, binary.BigEndian, &reportedServerSessionKey); err != nil {
		return respondWith(login.Fail, err)
	}

	if reportedServerSessionKey != serverSessionKey {
		return respondWith(login.BadSessionId, errors.New(fmt.Sprintf("session key mismatch.  Expected: %x, got: %x", serverSessionKey, reportedServerSessionKey)))
	}

	if err = binary.Read(buf, binary.BigEndian, &uid); err != nil {
		return respondWith(login.Fail, err)
	}

	if username, err = buf.ReadString(10); err != nil {
		return respondWith(login.Fail, err)
	}

	if password, err = buf.ReadString(10); err != nil {
		return respondWith(login.Fail, err)
	}

	username = username[:len(username)-1]
	password = password[:len(password)-1]

	fmt.Printf("Login request for %v\n", username)

	// Todo: authentication!

	player.Authenticate(username, player.HashPassword(username, password))

	resp.Write([]byte{byte(login.WorldFull)})
	client.Write(resp)

	return nil
}


func (*LoginCodec) WritePacket(c *Client, p *Packet) {
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