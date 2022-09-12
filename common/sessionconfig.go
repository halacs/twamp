package common

import "time"

type TwampSessionConfig struct {
	// According to RFC 4656, if Conf-Receiver is not set, Receiver port
	// is the UDP port OWAMP-Test to which packets are
	// requested to be sent.
	ReceiverPort int
	// According to RFC 4656, if Conf-Sender is not set, Sender port is the
	// UDP port from which OWAMP-Test packets will be sent.
	SenderPort int
	// According to RFC 4656, Padding length is the number of octets to be
	// appended to the normal OWAMP-Test packet (see more on
	// padding in discussion of OWAMP-Test).
	Padding int
	// According to RFC 4656, Timeout (or a loss threshold) is an interval of time
	// (expressed as a timestamp). A packet belonging to the test session
	// that is being set up by the current Request-Session command will
	// be considered lost if it is not received during Timeout seconds
	// after it is sent.
	Timeout int
	TOS     int
	// If true, padding will be filled with zeros instead of random data.
	UseAllZeros bool
	// Interval between sending out two measurement packet
	Interval time.Duration
}
