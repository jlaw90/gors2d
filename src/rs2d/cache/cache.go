package cache

import (
	"fmt"
	"hash/crc32"
	"math"
	"os"
	"bytes"
	"encoding/binary"
)

var data_reader *os.File
var index_reader [5]*os.File

var TitleCrc [10]int32
var VersionTable [4][]uint16
var CrcTable [4][]int32

var Source string

func Initialise(path string) {
	name := "main_file_cache"

	Source = path

	var err error

	checkFile := func(file *os.File, err error) {
		if err != nil {
			fmt.Printf("Error opening cache file '%v' for reading: %v\n", file.Name(), err)
			os.Exit(-2)
		}
	}

	for i := 0; i < 5; i++ {
		index_reader[i], err = os.Open(fmt.Sprintf("%v/%v.idx%v", Source, name, i))
		checkFile(index_reader[i], err)
	}

	data_reader, err = os.Open(fmt.Sprintf("%v/%v.dat", Source, name))
	checkFile(data_reader, err)

	// Compute Title CRC
	for i := 0; i < 9; i++ {
		data := ReadEntry(1, i)
		TitleCrc[i] = int32(crc32.ChecksumIEEE(data))
	}
	var check int32 = 1234
	for i := 0; i < 9; i++ {
		check = (check << 1) + int32(TitleCrc[i])
	}
	TitleCrc[9] = check

	// Load version and CRC data
	types := []string{"model", "anim", "midi", "map"}

	versionlist, err := LoadArchive(ReadEntry(1, 5))

	if err != nil {
		fmt.Printf("Error loading versionlist: %v", err)
		os.Exit(-5)
	}

	for i, s := range types {
		version := versionlist.GetEntry(fmt.Sprintf("%v_version", s))
		reader := bytes.NewReader(version)
		l := len(version) / 2
		VersionTable[i] = make([]uint16, l)
		for j := 0; j < l; j++ {
			err = binary.Read(reader, binary.BigEndian, &VersionTable[i][j])
			if err != nil {
				fmt.Printf("Error loading version table: %v", err)
				os.Exit(-8)
			}
		}

		crc := versionlist.GetEntry(fmt.Sprintf("%v_crc", s))
		reader = bytes.NewReader(crc)
		l = len(crc) / 4
		CrcTable[i] = make([]int32, l)
		for j := 0; j < l; j++ {
			err = binary.Read(reader, binary.BigEndian, &CrcTable[i][j])
			if err != nil {
				fmt.Printf("Error loading CRC table: %v", err)
				os.Exit(-7)
			}
		}
	}
}

func ReadEntry(cache, file int) []byte {
	buffer := make([]byte, 8)
	head := buffer[:6]

	handleReadError := func(e error) bool {
		if e != nil {
			fmt.Printf("Error reading entry from cache (c:%v, f: %v): %v\n", cache, file, e)
			return true
		}
		return false
	}

	readFully := func(file *os.File, dest []byte, from int64, length int) bool {
		total := 0
		for total < length {
			read, err := file.ReadAt(dest[total:length-total], from+int64(total))
			if handleReadError(err) {
				return false
			}
			total += read
		}
		return true
	}

	info, err := data_reader.Stat()
	if err != nil {
		return nil
	}
	data_length := info.Size()
	block_count := int(data_length / int64(520))

	if !readFully(index_reader[cache-1], head, int64(file*6), 6) {
		return nil
	}

	size := ((int(head[0]) & 0xff) << 16) | ((int(head[1]) & 0xff) << 8) | (int(head[2]) & 0xff)
	block := ((int(head[3]) & 0xff) << 16) | ((int(head[4]) & 0xff) << 8) | (int(head[5]) & 0xff)

	if size < 0 || block <= 0 || block >= block_count {
		return nil
	}

	data := make([]byte, size)

	total := 0
	for blocks := 0; total < size; blocks++ {
		if block == 0 {
			return nil
		}

		toRead := int(math.Min(float64(size-total), 512))
		dest := data[total : total+toRead]
		if !readFully(data_reader, buffer, int64(block*520), 8) || !readFully(data_reader, dest, int64(block*520+8), toRead) {
			return nil
		}
		total += toRead
		entryId := ((int(buffer[0]) & 0xff) << 8) | (int(buffer[1]) & 0xff)
		blockNum := ((int(buffer[2]) & 0xff) << 8) | (int(buffer[3]) & 0xff)
		nextBlock := ((int(buffer[4]) & 0xff) << 16) | ((int(buffer[5]) & 0xff) << 8) + (int(buffer[6]) & 0xff)
		cacheId := int(buffer[7]) & 0xff

		if entryId != file || blockNum != blocks || cacheId != cache {
			return nil
		}
		if nextBlock < 0 || nextBlock > block_count {
			return nil
		}

		block = nextBlock
	}

	return data
}
