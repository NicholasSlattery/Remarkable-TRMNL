// Package protocol implements AppLoad's two-packet SOCK_SEQPACKET backend protocol.
package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

const (
	SystemTerminate      uint32 = 0xffffffff
	SystemNewCoordinator uint32 = 0xfffffffe
	MaxPayload                  = 10 * 1024 * 1024
)

type Message struct {
	Type     uint32
	Contents string
}

type Connection struct {
	c       *net.UnixConn
	writeMu sync.Mutex
}

func Dial(path string) (*Connection, error) {
	addr := &net.UnixAddr{Name: path, Net: "unixpacket"}
	c, err := net.DialUnix("unixpacket", nil, addr)
	if err != nil {
		return nil, err
	}
	return &Connection{c: c}, nil
}

func (c *Connection) Close() error { return c.c.Close() }

func (c *Connection) Receive() (Message, error) {
	h := make([]byte, 8)
	n, _, err := c.c.ReadFromUnix(h)
	if err != nil {
		return Message{}, err
	}
	if n != 8 {
		return Message{}, fmt.Errorf("short AppLoad header: %d", n)
	}
	typ := binary.LittleEndian.Uint32(h[:4])
	length := binary.LittleEndian.Uint32(h[4:])
	if length > MaxPayload {
		return Message{}, errors.New("AppLoad payload exceeds limit")
	}
	// AppLoad's Qt/C++ sender always emits a second SOCK_SEQPACKET record,
	// including when the declared payload length is zero. Consume that record;
	// otherwise it is mistaken for EOF on the next header read.
	// Go short-circuits reads into a zero-length slice, so use one byte of
	// capacity to force recvmsg(2) to consume AppLoad's empty packet.
	payloadSize := int(length)
	if payloadSize == 0 {
		payloadSize = 1
	}
	b := make([]byte, payloadSize)
	n, _, err = c.c.ReadFromUnix(b)
	if err != nil {
		// net.UnixConn reports a consumed zero-length SOCK_SEQPACKET record as
		// io.EOF. It is the expected payload for an empty AppLoad QString.
		if length == 0 && errors.Is(err, io.EOF) && n == 0 {
			return Message{Type: typ}, nil
		}
		return Message{}, err
	}
	if n != int(length) {
		return Message{}, io.ErrUnexpectedEOF
	}
	return Message{Type: typ, Contents: string(b[:length])}, nil
}

func (c *Connection) Send(typ uint32, contents string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	b := []byte(contents)
	if len(b) > MaxPayload {
		return errors.New("AppLoad payload exceeds limit")
	}
	h := make([]byte, 8)
	binary.LittleEndian.PutUint32(h[:4], typ)
	binary.LittleEndian.PutUint32(h[4:], uint32(len(b)))
	if _, err := c.c.Write(h); err != nil {
		return err
	}
	if len(b) > 0 {
		_, err := c.c.Write(b)
		return err
	}
	return nil
}
