package light

import (
	"github.com/halacs/twamp"
)

type TwampLightConnection struct {
	hostname string
	port     uint16
}

func NewTwampLightConnection(hostname string, port uint16) *TwampLightConnection {
	return &TwampLightConnection{
		hostname: hostname,
		port:     port,
	}
}

func (c *TwampLightConnection) CreateLightSession(config twamp.TwampSessionConfig) (*TwampLightSession, error) {
	session := &TwampLightSession{connection: c, config: config}
	return session, nil
}

func (c *TwampLightConnection) Close() {
}
