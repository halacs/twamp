package light

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/halacs/twamp/common"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"time"
	"unsafe"
)

/*
TWAMP test connection used for running TWAMP tests.
*/
type TwampLightTest struct {
	Session    *TwampLightSession
	Connection *net.UDPConn
	Sequence   uint32
}

/*
 */
func (t *TwampLightTest) SetConnection(connection *net.UDPConn) {
	c := ipv4.NewConn(connection)

	// RFC recommends IP TTL of 255
	err := c.SetTTL(255)
	if err != nil {
		log.Fatal(err)
	}

	err = c.SetTOS(t.GetSession().GetConfig().TOS)
	if err != nil {
		log.Fatal(err)
	}

	t.Connection = connection
}

/*
Get TWAMP Test UDP connection.
*/
func (t *TwampLightTest) GetConnection() *net.UDPConn {
	return t.Connection
}

/*
Get the underlying TWAMP control Session for the TWAMP test.
*/
func (t *TwampLightTest) GetSession() *TwampLightSession {
	return t.Session
}

/*
Get the remote TWAMP IP/UDP address.
*/
func (t *TwampLightTest) RemoteAddr() (*net.UDPAddr, error) {
	address := fmt.Sprintf("%s:%d", t.GetRemoteTestHost(), t.GetRemoteTestPort())
	return net.ResolveUDPAddr("udp", address)
}

/*
Get the remote TWAMP UDP port number.
*/
func (t *TwampLightTest) GetRemoteTestPort() int {
	return t.GetSession().connection.port
}

/*
Get the local IP address for the TWAMP control Session.
*/
func (t *TwampLightTest) GetLocalTestHost() string {
	//localAddress := t.Session.GetConnection().LocalAddr()
	//return strings.Split(localAddress.String(), ":")[0]
	// TODO implement this correctly for light session
	return ""
}

/*
Get the remote IP address for the TWAMP control Session.
*/
func (t *TwampLightTest) GetRemoteTestHost() string {
	return t.GetSession().connection.hostname
}

type MeasurementPacket struct {
	Sequence            uint32
	Timestamp           common.TwampTimestamp
	ErrorEstimate       uint16
	MBZ                 uint16
	ReceiveTimeStamp    common.TwampTimestamp
	SenderSequence      uint32
	SenderTimeStamp     common.TwampTimestamp
	SenderErrorEstimate uint16
	Mbz                 uint16
	SenderTtl           byte
	//Padding []byte
}

/*
Run a TWAMP test and return a pointer to the TwampResults.
*/
func (t *TwampLightTest) Run() (*common.TwampResults, error) {
	paddingSize := t.GetSession().config.Padding
	senderSeqNum := t.Sequence

	size := t.sendTestMessage(t.GetSession().config.UseAllZeros)

	// receive test packets - allocate a receive buffer of a size we expect to receive plus a bit to know if we get some garbage
	buffer, err := common.ReadFromSocket(t.GetConnection(), (int(unsafe.Sizeof(MeasurementPacket{}))+paddingSize)*2)
	if err != nil {
		return nil, err
	}

	finished := time.Now()

	responseHeader := MeasurementPacket{}
	err = binary.Read(&buffer, binary.BigEndian, &responseHeader)
	if err != nil {
		log.Fatalf("Failed to deserialize measurement package. %v", err)
	}

	responsePadding := make([]byte, paddingSize, paddingSize)
	receivedPaddignSize, err := buffer.Read(responsePadding)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error when receivin padding. %v\n", err))
	}

	if receivedPaddignSize != paddingSize {
		return nil, errors.New(fmt.Sprintf("Incorrect padding. Expected padding size was %d but received %d.\n", paddingSize, receivedPaddignSize))
	}

	// process test results
	r := &common.TwampResults{}
	r.SenderSize = size
	r.SeqNum = responseHeader.Sequence
	r.Timestamp = common.NewTimestamp(responseHeader.Timestamp)
	r.ErrorEstimate = responseHeader.ErrorEstimate
	r.ReceiveTimestamp = common.NewTimestamp(responseHeader.ReceiveTimeStamp)
	r.SenderSeqNum = responseHeader.SenderSequence
	r.SenderTimestamp = common.NewTimestamp(responseHeader.SenderTimeStamp)
	r.SenderErrorEstimate = responseHeader.SenderErrorEstimate
	r.SenderTTL = responseHeader.SenderTtl
	r.FinishedTimestamp = finished

	if senderSeqNum != r.SenderSeqNum {
		return nil, errors.New(
			fmt.Sprintf("Expected Sequence # %d but received %d.\n", senderSeqNum, r.SeqNum),
		)
	}

	return r, nil
}

func (t *TwampLightTest) sendTestMessage(useAllZeros bool) int {
	packetHeader := MeasurementPacket{
		Sequence:            t.Sequence,
		Timestamp:           *common.NewTwampTimestamp(time.Now()),
		ErrorEstimate:       0x0101,
		MBZ:                 0x0000,
		ReceiveTimeStamp:    common.TwampTimestamp{},
		SenderSequence:      0,
		SenderTimeStamp:     common.TwampTimestamp{},
		SenderErrorEstimate: 0x0000,
		Mbz:                 0x0000,
		SenderTtl:           87,
	}

	// seed psuedo-random number generator if requested
	if !useAllZeros {
		rand.NewSource(int64(time.Now().Unix()))
	}

	paddingSize := t.GetSession().config.Padding
	padding := make([]byte, paddingSize, paddingSize)

	for x := 0; x < paddingSize; x++ {
		if useAllZeros {
			padding[x] = 0
		} else {
			padding[x] = byte(rand.Intn(255))
		}
	}

	var binaryBuffer bytes.Buffer
	err := binary.Write(&binaryBuffer, binary.BigEndian, packetHeader)
	if err != nil {
		log.Fatalf("Failed to serialize measurement package. %v", err)
	}

	headerBytes := binaryBuffer.Bytes()
	headerSize := binaryBuffer.Len()
	totalSize := headerSize + paddingSize
	var pdu []byte = make([]byte, totalSize)
	copy(pdu[0:], headerBytes)
	copy(pdu[headerSize:], padding)

	t.GetConnection().Write(pdu)
	t.Sequence++
	return totalSize
}

func (t *TwampLightTest) FormatJSON(r *common.PingResults) {
	doc, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", string(doc))
}

func (t *TwampLightTest) ReturnJSON(r *common.PingResults) string {
	doc, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s\n", string(doc))
}

func (t *TwampLightTest) Ping(count int, isRapid bool, interval int) *common.PingResults {
	Stats := &common.PingResultStats{}
	Results := &common.PingResults{Stat: Stats}
	var TotalRTT time.Duration = 0

	packetSize := 14 + t.GetSession().GetConfig().Padding

	fmt.Printf("TWAMP PING %s: %d data bytes\n", t.GetRemoteTestHost(), packetSize)

	for i := 0; i < count; i++ {
		Stats.Transmitted++
		results, err := t.Run()
		if err != nil {
			// TODO Do we need error logging here? I guess not because dot represents the sort error message here but should be double checked.
			if isRapid {
				fmt.Printf(".")
			}
		} else {
			if i == 0 {
				Stats.Min = results.GetRTT()
				Stats.Max = results.GetRTT()
			}
			if Stats.Min > results.GetRTT() {
				Stats.Min = results.GetRTT()
			}
			if Stats.Max < results.GetRTT() {
				Stats.Max = results.GetRTT()
			}

			TotalRTT += results.GetRTT()
			Stats.Received++
			Results.Results = append(Results.Results, results)

			if isRapid {
				fmt.Printf("!")
			} else {
				fmt.Printf("%d bytes from %s: twamp_seq=%d ttl=%d time=%0.03f ms\n",
					packetSize,
					t.GetRemoteTestHost(),
					results.SenderSeqNum,
					results.SenderTTL,
					(float64(results.GetRTT()) / float64(time.Millisecond)),
				)
			}
		}

		if !isRapid {
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}

	if isRapid {
		fmt.Printf("\n")
	}

	Stats.Avg = time.Duration(int64(TotalRTT) / int64(count))
	Stats.Loss = float64(float64(Stats.Transmitted-Stats.Received)/float64(Stats.Transmitted)) * 100.0
	Stats.StdDev = Results.StdDev(Stats.Avg)

	fmt.Printf("--- %s twamp ping statistics ---\n", t.GetRemoteTestHost())
	fmt.Printf("%d packets transmitted, %d packets received, %0.1f%% packet loss\n",
		Stats.Transmitted,
		Stats.Received,
		Stats.Loss)
	fmt.Printf("round-trip min/avg/max/stddev = %0.3f/%0.3f/%0.3f/%0.3f ms\n",
		(float64(Stats.Min) / float64(time.Millisecond)),
		(float64(Stats.Avg) / float64(time.Millisecond)),
		(float64(Stats.Max) / float64(time.Millisecond)),
		(float64(Stats.StdDev) / float64(time.Millisecond)),
	)
	defer t.Connection.Close()

	return Results
}

func (t *TwampLightTest) updateStats(TotalRTT time.Duration, count int, stats *common.PingResultStats, Results *common.PingResults) {
	stats.Avg = time.Duration(int64(TotalRTT) / int64(count))
	stats.Loss = float64(float64(stats.Transmitted-stats.Received)/float64(stats.Transmitted)) * 100.0
	stats.StdDev = Results.StdDev(stats.Avg)
}

func (t *TwampLightTest) RunX(count int, callback common.TwampTestCallbackFunction, doneSignal chan bool) *common.PingResults {
	defer t.Connection.Close()

	Stats := &common.PingResultStats{}
	Results := &common.PingResults{Stat: Stats}
	var TotalRTT time.Duration = 0

	terminationRequested := false
	for i := 0; i < count && !terminationRequested; i++ {
		select {
		case <-doneSignal:
			terminationRequested = true
		default:
			Stats.Transmitted++
			results, err := t.Run()

			if err != nil {
				// Packet lost somehow
				log.Printf("%v\n", err)
			} else {
				// Packet received
				if i == 0 {
					Stats.Min = results.GetRTT()
					Stats.Max = results.GetRTT()
				}
				if Stats.Min > results.GetRTT() {
					Stats.Min = results.GetRTT()
				}
				if Stats.Max < results.GetRTT() {
					Stats.Max = results.GetRTT()
				}

				TotalRTT += results.GetRTT()
				Stats.Received++
				Results.Results = append(Results.Results, results)
			}

			t.updateStats(TotalRTT, count, Stats, Results)
			if callback != nil {
				callback(count, results, *Stats)
			}

			// Wait in a way can be interrupted by user
			d := t.GetSession().GetConfig().Interval
			for i := 0; int64(i) < d.Milliseconds() && !terminationRequested; i++ {
				select {
				case <-doneSignal:
					terminationRequested = true
				default:
					time.Sleep(1 * time.Millisecond)
				}
			}
		}
	}

	return Results
}
