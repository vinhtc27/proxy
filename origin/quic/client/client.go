package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"

	"github.com/quic-go/quic-go"
)

const message = "foobar"

func main() {
	host := os.Args[1]
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(context.Background(), host, tlsConf, nil)
	if err != nil {
		return
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return
	}

	fmt.Printf("Client: Sending '%s'\n", message)
	_, err = stream.Write([]byte(message))
	if err != nil {
		return
	}

	buf := make([]byte, len(message))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return
	}
	fmt.Printf("Client: Got '%s'\n", buf)
}
