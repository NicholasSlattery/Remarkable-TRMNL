//go:build linux

package protocol

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// dialPair mirrors AppLoad's transport: a SOCK_SEQPACKET Unix socket where the
// Qt/C++ side is the listener and the Go backend dials in.
func dialPair(t *testing.T) (*net.UnixConn, *Connection) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "appload.sock")
	listener, err := net.ListenUnix("unixpacket", &net.UnixAddr{Name: path, Net: "unixpacket"})
	if err != nil {
		t.Skipf("unixpacket is unavailable: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(path) })

	type accepted struct {
		conn *net.UnixConn
		err  error
	}
	incoming := make(chan accepted, 1)
	go func() {
		c, err := listener.AcceptUnix()
		incoming <- accepted{conn: c, err: err}
	}()

	client, err := Dial(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	select {
	case a := <-incoming:
		if a.err != nil {
			t.Fatalf("accept: %v", a.err)
		}
		t.Cleanup(func() { _ = a.conn.Close() })
		return a.conn, client
	case <-time.After(5 * time.Second):
		t.Fatal("timed out accepting the AppLoad connection")
		return nil, nil
	}
}

// sendLikeAppLoad reproduces the Qt sender, which always emits a second
// sequence packet even when the declared payload length is zero.
func sendLikeAppLoad(t *testing.T, conn *net.UnixConn, typ uint32, contents string) {
	t.Helper()
	header := []byte{byte(typ), byte(typ >> 8), byte(typ >> 16), byte(typ >> 24), 0, 0, 0, 0}
	payload := []byte(contents)
	n := uint32(len(payload))
	header[4], header[5], header[6], header[7] = byte(n), byte(n>>8), byte(n>>16), byte(n>>24)
	if _, err := conn.Write(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
}

func TestReceiveConsumesEmptyPayloadPacket(t *testing.T) {
	peer, client := dialPair(t)
	// An empty QString followed by a real message: if the empty second packet
	// is not consumed, the next header read misreads it as EOF.
	sendLikeAppLoad(t, peer, 1, "")
	sendLikeAppLoad(t, peer, 2, `{"api_key":"x"}`)

	first, err := client.Receive()
	if err != nil {
		t.Fatalf("receive empty message: %v", err)
	}
	if first.Type != 1 || first.Contents != "" {
		t.Fatalf("first message = %+v, want type 1 with no contents", first)
	}
	second, err := client.Receive()
	if err != nil {
		t.Fatalf("receive follow-up message: %v", err)
	}
	if second.Type != 2 || second.Contents != `{"api_key":"x"}` {
		t.Fatalf("second message = %+v", second)
	}
}

func TestReceiveRejectsOversizedPayload(t *testing.T) {
	peer, client := dialPair(t)
	header := []byte{7, 0, 0, 0, 0, 0, 0, 0}
	oversize := uint32(MaxPayload + 1)
	header[4], header[5], header[6], header[7] = byte(oversize), byte(oversize>>8), byte(oversize>>16), byte(oversize>>24)
	if _, err := peer.Write(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := client.Receive(); err == nil {
		t.Fatal("oversized declared payload was accepted")
	}
}

func TestSendFramesHeaderAndPayload(t *testing.T) {
	peer, client := dialPair(t)
	body := `{"message":"Updated"}`
	if err := client.Send(msgStatusForTest, body); err != nil {
		t.Fatalf("send: %v", err)
	}
	if err := peer.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	header := make([]byte, 8)
	if n, err := peer.Read(header); err != nil || n != 8 {
		t.Fatalf("read header: n=%d err=%v", n, err)
	}
	gotType := uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16 | uint32(header[3])<<24
	gotLen := uint32(header[4]) | uint32(header[5])<<8 | uint32(header[6])<<16 | uint32(header[7])<<24
	if gotType != msgStatusForTest || int(gotLen) != len(body) {
		t.Fatalf("header type=%d length=%d, want %d and %d", gotType, gotLen, msgStatusForTest, len(body))
	}
	payload := make([]byte, gotLen)
	if _, err := peer.Read(payload); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	if string(payload) != body {
		t.Fatalf("payload = %q", payload)
	}
}

func TestSendRejectsOversizedPayload(t *testing.T) {
	_, client := dialPair(t)
	err := client.Send(1, strings.Repeat("a", MaxPayload+1))
	if err == nil {
		t.Fatal("oversized payload was accepted")
	}
}

func TestReceiveReportsTerminate(t *testing.T) {
	peer, client := dialPair(t)
	sendLikeAppLoad(t, peer, SystemTerminate, "")
	m, err := client.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if m.Type != SystemTerminate {
		t.Fatalf("type = %d, want SystemTerminate", m.Type)
	}
}

const msgStatusForTest uint32 = 103
