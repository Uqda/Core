package main

import (
	"crypto/ed25519"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/blake2b"
)

const (
	metaVersionMajor uint16 = iota
	metaVersionMinor
	metaPublicKey
	metaPriority
)

func main() {
	listen := flag.String("listen", "127.0.0.1:12105", "TCP address to listen on")
	password := flag.String("password", "", "legacy link password")
	flag.Parse()

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		fatal(err)
	}
	response, err := legacyHello(privateKey, publicKey, []byte(*password))
	if err != nil {
		fatal(err)
	}

	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		fatal(err)
	}
	defer listener.Close()
	fmt.Printf("legacy 0.5 peer listening on %s\n", *listen)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fatal(err)
		}
		go handle(conn, response)
	}
}

func handle(conn net.Conn, response []byte) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	header := make([]byte, 6)
	if _, err := io.ReadFull(conn, header); err != nil {
		return
	}
	if string(header[:4]) != "meta" {
		return
	}
	bodyLen := int(binary.BigEndian.Uint16(header[4:6]))
	if bodyLen < ed25519.SignatureSize || bodyLen > 4096 {
		return
	}
	body := make([]byte, bodyLen)
	if _, err := io.ReadFull(conn, body); err != nil {
		return
	}

	// A deployed 0.5 peer ignores the new extension TLVs and responds with the
	// original public-key-authenticated hello, without the secure-confirmation
	// capability. A secure=required UQDA peer must reject this response.
	_, _ = conn.Write(response)
	buf := make([]byte, 1)
	_, _ = conn.Read(buf)
}

func legacyHello(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, password []byte) ([]byte, error) {
	body := make([]byte, 0, 128)
	body = appendField(body, metaVersionMajor, []byte{0, 0})
	body = appendField(body, metaVersionMinor, []byte{0, 5})
	body = appendField(body, metaPublicKey, publicKey)
	body = appendField(body, metaPriority, []byte{0})

	hasher, err := blake2b.New512(password)
	if err != nil {
		return nil, err
	}
	if _, err := hasher.Write(publicKey); err != nil {
		return nil, err
	}
	body = append(body, ed25519.Sign(privateKey, hasher.Sum(nil))...)

	message := append([]byte{'m', 'e', 't', 'a', 0, 0}, body...)
	binary.BigEndian.PutUint16(message[4:6], uint16(len(body)))
	return message, nil
}

func appendField(dst []byte, op uint16, value []byte) []byte {
	dst = binary.BigEndian.AppendUint16(dst, op)
	dst = binary.BigEndian.AppendUint16(dst, uint16(len(value)))
	return append(dst, value...)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
