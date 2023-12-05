package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"

	"github.com/quic-go/quic-go"
)

const message = "foobar"

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(context.Background(), "127.0.0.1:8080", tlsConf, nil)
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
