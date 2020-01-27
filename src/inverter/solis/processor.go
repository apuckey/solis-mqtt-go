package solis

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"net"
	cfg "solis-go/config"
	"strconv"
	"time"
)

var (
	FrameHeader  = "680241b1"
	FrameCommand = "0100"
	FrameEnd     = "16"
)

type Solis struct {
	connectchan    chan bool
	sendchan       chan *InverterData
	mq             mqtt.Client
	conn           net.Conn
	status         uint32
	frame          []byte
	host           string
	connecttimeout time.Duration
	readtimeout    time.Duration
	readinterval   time.Duration
}

func New(mq mqtt.Client, config *cfg.Config) *Solis {
	s := &Solis{
		mq:             mq,
		sendchan:       make(chan *InverterData),
		host:           config.Inverter.Host,
		connecttimeout: time.Duration(config.Inverter.ConnectTimeout) * time.Second,
		readtimeout:    time.Duration(config.Inverter.ReadTimeout) * time.Second,
		readinterval:   time.Duration(config.Inverter.ReadInterval) * time.Second,
	}

	s.generateFrame(config.Inverter.Serial)

	go s.sendToMQTT()

	return s
}

func (inv *Solis) Run() {
	for {
		msg, err := inv.getLiveData()
		if err != nil {
			// this happens a lot with my wifi dongle. unsure if its broken or its bad firmware.
			// for now im just going to ignore the error as theres nothing i can do about it
			// and dont want my pi syslog filling up :-)
			//fmt.Println(fmt.Sprintf("Unable to get data from inverter: %s", err.Error()))
		} else {
			d, err := inv.Process(msg)
			if err == nil {
				inv.sendchan <- d
			}
		}
		time.Sleep(inv.readinterval)
	}
}

func (inv *Solis) Process(b []byte) (*InverterData, error) {

	bindata := &BinaryData{}
	buffer := bytes.NewBuffer(b)

	err := binary.Read(buffer, binary.BigEndian, bindata)
	if err != nil {
		return nil, fmt.Errorf("unable to parse data")
	}

	// convert to readable format
	inverterData := &InverterData{}
	inverterData.ParseBinaryData(bindata)
	return inverterData, nil
}

func (inv *Solis) sendToMQTT() {
	for {
		msg := <-inv.sendchan

		jsondata, err := json.Marshal(msg)
		if err == nil {
			inv.sendToQueue("/solar/payload", string(jsondata), true)
		}
	}
}

func (inv *Solis) calcChecksum(src []byte) []byte {
	var checksum uint64
	for _, b := range src[:len(src)-8] {
		i, _ := strconv.ParseUint(string(b), 16, 64)
		checksum += i & 255
	}
	return []byte(fmt.Sprintf("%x", checksum&255))

}

func (inv *Solis) sendDataToInverter(conn net.Conn) error {
	_, err := conn.Write(inv.frame)
	return err
}

// My solis inverter sends 2 blocks of data then stops.
// I think either my wifi dongle is broken or the firmware doesnt like me,
// If i either continue to use the same connection or disconnect and reconnect each time
// the wifi dongle reboots itself.
// for now i just let it disconnect me and then try again, this seems to have the best outcome.
func (inv *Solis) readDataFromInverter(conn net.Conn) ([]byte, error) {
	_ = conn.SetReadDeadline(time.Now().Add(inv.readtimeout))
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 512)
	//var count int
	for {
		//if count == 2 {
		//	break
		//}
		n, err := conn.Read(tmp)
		if err != nil {
			//if err != io.EOF {
				// return the error so we can reconnect.
				//return nil, err
				break
			//}
		}
		//count++
		buf = append(buf, tmp[:n]...)
	}
	return buf, nil
}

func (inv *Solis) getLiveData() ([]byte, error) {
	conn, err := net.DialTimeout("tcp", inv.host, inv.connecttimeout)

	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	err = inv.sendDataToInverter(conn)
	if err != nil {
		return nil, err
	}
	return inv.readDataFromInverter(conn)
}

func (inv *Solis) generateFrame(serial uint64) {
	sn := []byte(fmt.Sprintf("%x", serial))
	rsn := fmt.Sprintf("%s%s%s%s", sn[6:8], sn[4:6], sn[2:4], sn[0:2])

	src := []byte(fmt.Sprintf(
		"%s%s%s%s%s%s",
		FrameHeader,
		rsn,
		rsn,
		FrameCommand,
		inv.calcChecksum([]byte(fmt.Sprintf("%s%s%s%s%s%s", FrameHeader, rsn, rsn, FrameCommand, "00", FrameEnd))),
		FrameEnd))
	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		fmt.Printf("err: %s", err.Error())
	}

	inv.frame = dst[:n]
}

func (inv *Solis) sendToQueue(queue string, msg interface{}, retained bool) {
	t := inv.mq.Publish(queue, 0, retained, fmt.Sprintf("%v", msg))
	if t.Wait() && t.Error() != nil {
		fmt.Printf("Error sending data to the broker: %v", t.Error())
	}
}
