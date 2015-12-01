package cache

import (
	"bytes"
	"encoding/binary"
	"compress/bzip2"
	"strings"
	"io"
)

type entry struct {
	size int
	compressedSize int
	offset int
}

type Archive struct {
	individualCompression bool
	data []byte
	numFiles uint16
	table map[int]entry
}

func (a *Archive) GetEntry(name string) []byte {
	var hash int32 = 0
	name = strings.ToUpper(name)
	for _, r := range name {
		hash = (hash * 61 + int32(r)) - 32
	}

	e, ok := a.table[int(hash)]

	if !ok {
		return nil
	}

	data := a.data[e.offset:e.offset+e.size]
	if a.individualCompression {
		decompressed := make([]byte, e.size)
		reader := bzip2.NewReader(bytes.NewReader(data))
		_, _ = io.ReadFull(reader, decompressed)
		return decompressed
	} else {
		return data
	}
}

func LoadArchive(data []byte) (a *Archive, err error) {
	reader := bytes.NewReader(data)

	var (
		a8 uint8
		a16 uint16
		dataStart int
	)

	a = new(Archive)
	_ = binary.Read(reader, binary.BigEndian, &a16)
	_ = binary.Read(reader, binary.BigEndian, &a8)
	l := (uint(a16) << 8) | uint(a8)

	binary.Read(reader, binary.BigEndian, &a16)
	binary.Read(reader, binary.BigEndian, &a8)
	compressedLen := (uint(a16) << 8) | uint(a8)

	if compressedLen != l {
		bzipReader := bzip2.NewReader(io.MultiReader(bytes.NewReader([]byte{0x42, 0x5a, 'h', '9'}), reader))
		data = make([]byte, l)
		_, err := io.ReadFull(bzipReader, data)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	} else {
		dataStart = 6
		a.individualCompression = true
	}
	a.data = data

	_ = binary.Read(reader, binary.BigEndian, &a.numFiles)

	a.table = make(map[int]entry)

	n := int(a.numFiles)
	off := dataStart + n * 10;
	for i := 0; i < n; i++ {
		var hash int32
		_ = binary.Read(reader, binary.BigEndian, &hash)

		e := new(entry)
		_ = binary.Read(reader, binary.BigEndian, &a16)
		_ = binary.Read(reader, binary.BigEndian, &a8)
		e.size = (int(a16) << 8) | int(a8)
		_ = binary.Read(reader, binary.BigEndian, &a16)
		_ = binary.Read(reader, binary.BigEndian, &a8)
		e.compressedSize = (int(a16) << 8) | int(a8)
		e.offset = off
		a.table[int(hash)] = *e
		off += e.compressedSize
	}
	return a, nil
}

