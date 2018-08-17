package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Filenames can be 15 chars max in the AGG FAT
const fatSizeName int = 15

type header struct {
	ItemCount uint16
}

type aggFat struct {
	crc    uint32
	offset uint32
	size   uint32
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("\nAGG Extractor. Copyright (c) 2015. Rob Morgan.")
	fmt.Println("Based on code from the fheroes2 project")
	fmt.Println("https://github.com/robmorgan")
	fmt.Println(strings.Repeat("*", 80) + "\n")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Please specify the location of an AGG to extract")
		return
	}
	aggFilename := flag.Arg(0)
	fmt.Printf("Found file: %s\n", aggFilename)

	outputDir := strings.TrimSuffix(aggFilename, filepath.Ext(aggFilename))
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err2 := os.Mkdir(outputDir, 0700)
		if err2 != nil {
			panic(err2)
		}
	}
	fmt.Printf("Output Directory: %s\n", outputDir)

	f, err := os.Open(aggFilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(statsErr)
	}
	size := stats.Size()
	fmt.Printf("Size of Agg (in bytes): %d\n", size)

	data := readNextBytes(f, 2)

	header := header{}
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.LittleEndian, &header)
	if err != nil {
		panic(err)
	}

	fmt.Printf("total items: %d\n", header.ItemCount)

	fatData := readNextBytes(f, int(header.ItemCount)*4*3) // crc, offset, size
	fatDataBuffer := bytes.NewBuffer(fatData)

	// Seek to filenames for building FAT map
	f.Seek(size-int64(fatSizeName)*int64(header.ItemCount), 0)

	m := make(map[string]*aggFat)
	total := 0
	for i := 0; i < int(header.ItemCount); i++ {
		nameData := readNextBytes(f, fatSizeName)
		nameBuffer := bytes.NewBuffer(nameData)
		nameBytes, err := nameBuffer.ReadBytes(0x00) // read until 0x00, for some reason they pad with random chars in the 15 byte chunk
		if err != nil {
			panic(err)
		}

		fileName := strings.ToLower(string(nameBytes[0 : len(nameBytes)-1]))

		crc, err := readLEU32(fatDataBuffer)
		if err != nil {
			panic(err)
		}
		offset, err := readLEU32(fatDataBuffer)
		if err != nil {
			panic(err)
		}
		size, err := readLEU32(fatDataBuffer)
		if err != nil {
			panic(err)
		}

		fat := &aggFat{
			crc:    crc,
			offset: offset,
			size:   size,
		}

		m[fileName] = fat
		total++
	}

	// Extract file data and write
	// dont bother sorting the map to avoid random disk reads
	// TODO - actually it turns out the maps are returning elements randomly
	dir, err := filepath.Abs(outputDir)
	fmt.Printf("going to extract to: %s\n", dir)

	for k, v := range m {
		f.Seek(int64(v.offset), 0)
		fileData := readNextBytes(f, int(v.size))

		// make sure deferred function calls are actually called to avoid too many open files
		func() {
			file, err := os.Create(fmt.Sprintf(outputDir+"/%s", k))
			if err != nil {
				panic(err)
			}
			defer file.Close()

			_, err = file.Write(fileData)
			if err != nil {
				panic(err)
			}

			fmt.Printf("extract: %s\n", k)
			defer file.Close()
		}()
	}
	fmt.Printf("total extracted: %d\n", total)
}

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

// readLEU32 reads a uint32 value with little endianess.
func readLEU32(r io.Reader) (v uint32, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}
