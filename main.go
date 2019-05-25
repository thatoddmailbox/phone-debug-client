package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"
)

func main() {
	log.Println("phone-debug-client")

	listener, err := net.Listen("tcp", ":8520")
	if err != nil {
		panic(err)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	conn.SetDeadline(time.Time{})

	// get first 4 bytes, which tells us how long the buffer is
	numBuffer := make([]byte, 4)

	n, err := conn.Read(numBuffer)
	if err != nil {
		panic(err)
	}

	if n < 4 {
		log.Fatalf("got wrong number of bytes, expected 4 but got %d", n)
	}

	bufferSize := binary.LittleEndian.Uint32(numBuffer)

	if bufferSize > 1*1024*1024 {
		bufferSize = 10
	}

	log.Printf("ready for buffer of size %d", bufferSize)

	finalBuffer := []byte{}
	rxBuffer := make([]byte, 128)
	readBytes := uint32(0)
	for readBytes < bufferSize {
		n, err := conn.Read(rxBuffer)
		if err != nil {
			panic(err)
		}

		finalBuffer = append(finalBuffer, rxBuffer[0:n]...)
		readBytes += uint32(n)
	}

	if readBytes > bufferSize {
		log.Printf("WARNING: got %d more bytes than expected!", readBytes-bufferSize)
	}

	if bufferSize < 2048 {
		fmt.Print(hex.Dump(finalBuffer))
	}

	ioutil.WriteFile("buffer.bin", finalBuffer, 777)

	log.Println("wrote to buffer.bin")
}
