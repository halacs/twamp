package full

import (
	"bytes"
	"encoding/binary"
	"github.com/halacs/twamp"
	"log"
	"net"
	"time"
)

type TwampFullConnection struct {
	connection net.Conn
}

func NewTwampFullConnection(conn net.Conn) *TwampFullConnection {
	return &TwampFullConnection{connection: conn}
}

func (c *TwampFullConnection) GetConnection() net.Conn {
	return c.connection
}

func (c *TwampFullConnection) Close() {
	c.GetConnection().Close()
}

func (c *TwampFullConnection) LocalAddr() net.Addr {
	return c.connection.LocalAddr()
}

func (c *TwampFullConnection) RemoteAddr() net.Addr {
	return c.connection.RemoteAddr()
}

/*
TWAMP client session negotiation message.
*/
type TwampClientSetUpResponse struct {
	Mode     uint32
	KeyID    [80]byte
	Token    [64]byte
	ClientIV [16]byte
}

/*
TWAMP server greeting message.
*/
type TwampServerGreeting struct {
	Mode      uint32   // modes (4 bytes)
	Challenge [16]byte // challenge (16 bytes)
	Salt      [16]byte // salt (16 bytes)
	Count     uint32   // count (4 bytes)
}

func (c *TwampFullConnection) sendTwampClientSetupResponse() {
	// negotiate TWAMP session configuration
	response := &TwampClientSetUpResponse{}
	response.Mode = ModeUnauthenticated
	binary.Write(c.GetConnection(), binary.BigEndian, response)
}

func (c *TwampFullConnection) getTwampServerGreetingMessage() (*TwampServerGreeting, error) {
	// check the greeting message from TWAMP server
	buffer, err := twamp.ReadFromSocket(c.connection, 64)
	if err != nil {
		log.Printf("Cannot read: %s\n", err)
		return nil, err
	}

	// decode the TwampServerGreeting PDU
	greeting := &TwampServerGreeting{}
	_ = buffer.Next(12)
	greeting.Mode = binary.BigEndian.Uint32(buffer.Next(4))
	copy(greeting.Challenge[:], buffer.Next(16))
	copy(greeting.Salt[:], buffer.Next(16))
	greeting.Count = binary.BigEndian.Uint32(buffer.Next(4))

	return greeting, nil
}

type TwampServerStart struct {
	Accept    byte
	ServerIV  [16]byte
	StartTime twamp.TwampTimestamp
}

func (c *TwampFullConnection) getTwampServerStartMessage() (*TwampServerStart, error) {
	// check the start message from TWAMP server
	buffer, err := twamp.ReadFromSocket(c.connection, 48)
	if err != nil {
		return nil, err
	}

	// decode the TwampServerStart PDU
	start := &TwampServerStart{}
	_ = buffer.Next(15)
	start.Accept = byte(buffer.Next(1)[0])
	copy(start.ServerIV[:], buffer.Next(16))
	start.StartTime.Integer = binary.BigEndian.Uint32(buffer.Next(4))
	start.StartTime.Fraction = binary.BigEndian.Uint32(buffer.Next(4))

	return start, nil
}

/* Byte offsets for Request-TW-Session TWAMP PDU */
const ( // TODO these constants should be removed as part of a refactor when control changel messages are refactored to use "struct based" messaging which is clearer
	offsetRequestTwampSessionCommand         = 0
	offsetRequestTwampSessionIpVersion       = 1
	offsetRequestTwampSessionSenderPort      = 12
	offsetRequestTwampSessionReceiverPort    = 14
	offsetRequestTwampSessionPaddingLength   = 64
	offsetRequestTwampSessionStartTime       = 68
	offsetRequestTwampSessionTimeout         = 76
	offsetRequestTwampSessionTypePDescriptor = 84
)

type RequestTwSession []byte

func (b RequestTwSession) Encode(c twamp.TwampSessionConfig) {
	start_time := twamp.NewTwampTimestamp(time.Now())
	b[offsetRequestTwampSessionCommand] = byte(5)
	b[offsetRequestTwampSessionIpVersion] = byte(4) // As per RFC, this value can be 4 (IPv4) or 6 (IPv6).
	binary.BigEndian.PutUint16(b[offsetRequestTwampSessionSenderPort:], uint16(c.SenderPort))
	binary.BigEndian.PutUint16(b[offsetRequestTwampSessionReceiverPort:], uint16(c.ReceiverPort))
	binary.BigEndian.PutUint32(b[offsetRequestTwampSessionPaddingLength:], uint32(c.Padding))
	binary.BigEndian.PutUint32(b[offsetRequestTwampSessionStartTime:], start_time.Integer)
	binary.BigEndian.PutUint32(b[offsetRequestTwampSessionStartTime+4:], start_time.Fraction)
	binary.BigEndian.PutUint32(b[offsetRequestTwampSessionTimeout:], uint32(c.Timeout))
	binary.BigEndian.PutUint32(b[offsetRequestTwampSessionTimeout+4:], 0)
	binary.BigEndian.PutUint32(b[offsetRequestTwampSessionTypePDescriptor:], uint32(c.TOS))
}

func (c *TwampFullConnection) CreateFullSession(config twamp.TwampSessionConfig) (*TwampFullSession, error) {
	var pdu RequestTwSession = make(RequestTwSession, 112)

	var session *TwampFullSession

	pdu.Encode(config)

	c.GetConnection().Write(pdu)

	acceptBuffer, err := twamp.ReadFromSocket(c.GetConnection(), 48)
	if err != nil {
		return nil, err
	}

	acceptSession := NewTwampAcceptSession(acceptBuffer)

	err = checkAcceptStatus(int(acceptSession.accept), "session")
	if err != nil {
		return nil, err
	}

	session = &TwampFullSession{connection: c, Port: acceptSession.port, Config: config}

	return session, nil
}

type TwampAcceptSession struct {
	accept byte
	port   uint16
	sid    [16]byte
}

func NewTwampAcceptSession(buf bytes.Buffer) *TwampAcceptSession {
	message := &TwampAcceptSession{}
	message.accept = byte(buf.Next(1)[0])
	_ = buf.Next(1) // mbz
	message.port = binary.BigEndian.Uint16(buf.Next(2))
	copy(message.sid[:], buf.Next(16))
	return message
}
