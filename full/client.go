package full

import (
	"errors"
	"fmt"
	"net"
	"time"
)

/*
Security modes for TWAMP session.
*/
const (
	ModeUnspecified     = 0
	ModeUnauthenticated = 1
	ModeAuthenticated   = 2
	ModeEncypted        = 4
)

type TwampFullClient struct{}

func NewFullClient() *TwampFullClient {
	return &TwampFullClient{}
}

func (c *TwampFullClient) Connect(hostname string, port int) (*TwampFullConnection, error) {
	// connect to remote host
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), time.Second*5)
	if err != nil {
		return nil, err
	}

	// create a new TwampFullConnection
	twampConnection := NewTwampFullConnection(conn)

	// check for greeting message from TWAMP server
	greeting, err := twampConnection.getTwampServerGreetingMessage()
	if err != nil {
		return nil, err
	}

	// check greeting mode for errors
	switch greeting.Mode {
	case ModeUnspecified:
		return nil, errors.New("The TWAMP server is not interested in communicating with you.")
	case ModeUnauthenticated:
	case ModeAuthenticated:
		return nil, errors.New("Authentication is not currently supported.")
	case ModeEncypted:
		return nil, errors.New("Encyption is not currently supported.")
	}

	// negotiate TWAMP session configuration
	twampConnection.sendTwampClientSetupResponse()

	// check the start message from TWAMP server
	serverStartMessage, err := twampConnection.getTwampServerStartMessage()
	if err != nil {
		return nil, err
	}

	err = checkAcceptStatus(int(serverStartMessage.Accept), "connection")
	if err != nil {
		return nil, err
	}

	return twampConnection, nil
}

/*
TWAMP Accept Field Status Code
*/
const (
	OK                          = 0
	Failed                      = 1
	InternalError               = 2
	NotSupported                = 3
	PermanentResourceLimitation = 4
	TemporaryResourceLimitation = 5
)

/*
Convenience function for checking the accept code contained in various TWAMP server
response messages.
*/
func checkAcceptStatus(accept int, context string) error {
	switch accept {
	case OK:
		return nil
	case Failed:
		return errors.New(fmt.Sprintf("ERROR: The %s failed.", context))
	case InternalError:
		return errors.New(fmt.Sprintf("ERROR: The %s failed: internal error.", context))
	case NotSupported:
		return errors.New(fmt.Sprintf("ERROR: The %s failed: not supported.", context))
	case PermanentResourceLimitation:
		return errors.New(fmt.Sprintf("ERROR: The %s failed: permanent resource limitation.", context))
	case TemporaryResourceLimitation:
		return errors.New(fmt.Sprintf("ERROR: The %s failed: temporary resource limitation.", context))
	}
	return nil
}
