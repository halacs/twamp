package light

import (
	"github.com/halacs/twamp/common"
)

type TwampLightConnection struct {
	hostname string
	port     int
}

func NewTwampLightConnection(hostname string, port int) *TwampLightConnection {
	return &TwampLightConnection{
		hostname: hostname,
		port:     port,
	}
}

func (c *TwampLightConnection) CreateLightSession(config common.TwampSessionConfig) (*TwampLightSession, error) {
	session := &TwampLightSession{connection: c, config: config}
	return session, nil
}

func (c *TwampLightConnection) Close() {
}
