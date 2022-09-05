package full

import (
	"encoding/binary"
	"fmt"
	"github.com/halacs/twamp"
	"log"
	"net"
)

type TwampFullSession struct {
	connection *TwampFullConnection
	Port       uint16
	Config     twamp.TwampSessionConfig
}

func (s *TwampFullSession) GetConnection() net.Conn {
	return s.connection.GetConnection()
}

func (s *TwampFullSession) GetConfig() twamp.TwampSessionConfig {
	return s.Config
}

func (s *TwampFullSession) Write(buf []byte) {
	s.GetConnection().Write(buf)
}

func (s *TwampFullSession) CreateTest() (*TwampFullTest, error) {
	var pdu []byte = make([]byte, 32)
	pdu[0] = 2

	s.Write(pdu)

	startAckBuffer, err := twamp.ReadFromSocket(s.GetConnection(), 32)
	if err != nil {
		return nil, err
	}

	accept, err := startAckBuffer.ReadByte()
	if err != nil {
		log.Printf("Cannot read: %s\n", err)
		return nil, err
	}

	err = checkAcceptStatus(int(accept), "test setup")
	if err != nil {
		return nil, err
	}

	test := &TwampFullTest{Session: s}
	remoteAddr, err := test.RemoteAddr()
	if err != nil {
		return nil, err
	}
	localAddress := fmt.Sprintf("%s:%d", test.GetLocalTestHost(), s.GetConfig().ReceiverPort)
	localAddr, err := net.ResolveUDPAddr("udp", localAddress)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		return nil, err
	}

	test.SetConnection(conn)

	return test, nil
}

func (s *TwampFullSession) Stop() {
	//	log.Println("Stopping test sessions.")
	var pdu []byte = make([]byte, 32)
	pdu[0] = byte(3)                       // Stop-Sessions Command Number
	pdu[1] = byte(0)                       // Accept Status (0 = OK)
	binary.BigEndian.PutUint16(pdu[4:], 1) // Number of Sessions
	s.GetConnection().Write(pdu)
}
