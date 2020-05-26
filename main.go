package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func rgb565toRGBA(rgb565 uint16) color.RGBA {
	rConverted := uint8(((rgb565 & 0xF800) >> 8) | ((rgb565 & 0xF800) >> 13))
	gConverted := uint8(((rgb565 & 0x7E0) >> 3) | ((rgb565 & 0x7E0) >> 9))
	bConverted := uint8(((rgb565 & 0x1F) << 3) | ((rgb565 & 0x1F) >> 2))

	return color.RGBA{
		R: rConverted,
		G: gConverted,
		B: bConverted,
		A: 255,
	}
}

func main() {
	log.Println("phone-debug-client")

	outputFormat := flag.String("output", "bin", "The output format (either 'bin' or 'png')")

	flag.Parse()

	if *outputFormat != "bin" && *outputFormat != "png" {
		log.Fatalf("unknown output format '%s' -- must be either 'bin' or 'png'", *outputFormat)
	}

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

	if *outputFormat == "bin" {
		ioutil.WriteFile("buffer.bin", finalBuffer, 777)

		log.Println("wrote to buffer.bin")
	} else if *outputFormat == "png" {
		// we make the (rather large) assumption that, if the output format is png
		// that the buffer is a 128x128 rgb565 image

		outputFile, err := os.Create("buffer.png")
		if err != nil {
			panic(err)
		}

		width := 128
		height := 128

		outputImage := image.NewRGBA(image.Rect(0, 0, width, height))

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dataIndex := 2 * ((y * width) + x)
				color565 := (uint16(finalBuffer[dataIndex]) << 8) | uint16(finalBuffer[dataIndex+1])
				outputImage.SetRGBA(x, y, rgb565toRGBA(color565))
			}
		}

		png.Encode(outputFile, outputImage)

		outputFile.Close()

		log.Println("wrote to buffer.png")
	}
}
