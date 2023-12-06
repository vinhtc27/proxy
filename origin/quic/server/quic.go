package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"proxy/utils"

	"github.com/quic-go/quic-go"
)

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	host := os.Args[1]
	log.Printf("GRPC started at %s\n", host)
	listener, err := quic.ListenAddr(host, utils.GenerateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := listener.Accept(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	}
	// Echo through the loggingWriter
	_, err = io.Copy(loggingWriter{stream}, stream)
	if err != nil {
		panic(err)
	}
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}
