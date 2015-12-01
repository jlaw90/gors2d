package server

import (
	"fmt"
	"net"
	"os"
	"rs2d/config"
	"time"
)

func Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Config.Port))

	if err != nil {
		fmt.Printf("Failed to initiate primary server: %v", err)
		os.Exit(-3)
	}

	go serverListen(listener)
}

func serverListen(listener net.Listener) {
	fmt.Printf("Server listening on %v\n", listener.Addr())

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("Error during server accept: %v\n", err)
		} else {
			go serverProcess(&conn)
		}
	}
}

func serverProcess(conn *net.Conn) {
	client := NewClient(conn)

	// We are now in a separate go-routine for this client and can perform blocking operations
	fmt.Printf("Client connected: %v\n", client)

	// Packet read loop to send to channel
	packetChan := make(chan *Packet)

	go serverReadLoop(client, packetChan)

	for client.Connected {
		select {
		case packet := <- packetChan:
			if packet != nil {
				fmt.Printf("Packet received: %v\n", packet)
				// Todo: process packet
			}

		default:
			time.Sleep(10)
		}
	}
	_ = (*client.Connection).Close()
}

func serverReadLoop(client *Client, channel chan *Packet) {
	for client.Connected {
		(*client.Connection).SetReadDeadline(time.Now().Add(time.Duration(config.Config.ReadTimeout)*time.Second))
		channel <- client.Codec.ReadPacket(client)
	}
}