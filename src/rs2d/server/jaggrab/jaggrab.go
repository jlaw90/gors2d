package jaggrab

import (
	"fmt"
	"net"
	"os"
	"rs2d/cache"
	"rs2d/config"
)

var fileMap = map[string]int{
	"title":       1,
	"config":      2,
	"interface":   3,
	"media":       4,
	"versionlist": 5,
	"textures":    6,
	"wordenc":     7,
	"sounds":      8,
}


func Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Config.JaggrabPort))
	if err != nil {
		fmt.Printf("Failed to initiate JAGGRAB server: %v", err)
		os.Exit(-1)
	}

	// Start a go-routine for accepting connections...
	go listen(listener)
}


// Internal package methods...
func listen(listener net.Listener) {
	fmt.Printf("JAGGRAB server listening on %v\n", listener.Addr())

	for {
		client, err := listener.Accept()

		if err != nil {
			fmt.Printf("Error during JAGGRAB accept: %v", err)
		} else {
			// Start a goroutine for responding to requests...
			go process(client)
		}
	}
}

func process(client net.Conn) {
	buf := make([]byte, 1)

	request := make([]byte, 0)

	for {
		read, err := client.Read(buf)
		if err != nil {
			fmt.Printf("Error reading from JAGGRAB client, disconnecting: %v\n", err)
			_ = client.Close()
			break
		} else {
			if read == 1 {
				request = append(request, buf[0])

				if len(request) > 2 && string(request[len(request) - 2:]) == "\n\n" {
					str := string(request[:len(request) - 2])

					processRequest(str, client)
					_ = client.Close()
					break
				}
			}
		}
	}
}

func processRequest(str string, client net.Conn) {
	if len(str) > 8 && str[:8] != "JAGGRAB " {
		fmt.Printf("Invalid request to JAGGRAB, disconnecting: %v", str)
		return
	}

	requested := str[8:]

	if requested[0:1] == "/" {
		requested = requested[1:]
	}

	crcidx := -1

	for i, codepoint := range requested {
		if codepoint >= '0' && codepoint <= '9' || codepoint == '-' {
			crcidx = i
			break
		}
	}

	file := requested[:crcidx]
	//crc := requested[crcidx:]

	//fmt.Printf("JAGGRAB request for: %v, File: %v, Expected CRC: %v\n", requested, file, crc)

	var data []byte

	switch {
	case file == "crc":
		data = make([]byte, 40)

		for i := 0; i < 10; i++ {
			c := cache.TitleCrc[i]
			b1, b2, b3, b4 := byte((c >> 24) & 0xff), byte((c >> 16) & 0xff), byte((c >> 8) & 0xff), byte(c & 0xff)
			data[i*4+0] = b1
			data[i*4+1] = b2
			data[i*4+2] = b3
			data[i*4+3] = b4
		}
	default:
		id := fileMap[string(file)]
		data = cache.ReadEntry(1, id)

		//realCrc := cache.TitleCrc[id]
		// Todo: check CRC matches?
	}

	for total := 0; total < len(data); {
		written, err := client.Write(data)
		if err != nil {
			fmt.Printf("Error responding to JAGGRAB request, disconnecting: %v", err)
			return
		}
		total += written
	}
}


