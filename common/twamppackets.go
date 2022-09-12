package common

type MeasurementPacket struct {
	Sequence            uint32
	Timestamp           TwampTimestamp
	ErrorEstimate       uint16
	MBZ                 uint16
	ReceiveTimeStamp    TwampTimestamp
	SenderSequence      uint32
	SenderTimeStamp     TwampTimestamp
	SenderErrorEstimate uint16
	Mbz                 uint16
	SenderTtl           byte
	//Padding []byte
}
