package light

import (
	"fmt"
	"github.com/halacs/twamp"
	"net"
)

type TwampLightSession struct {
	connection *TwampLightConnection
	config     twamp.TwampSessionConfig
}

func (s *TwampLightSession) GetConfig() twamp.TwampSessionConfig {
	return s.config
}

func (s *TwampLightSession) CreateTest() (*TwampLightTest, error) {
	test := &TwampLightTest{Session: s}
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

func (s *TwampLightSession) Stop() {
}
