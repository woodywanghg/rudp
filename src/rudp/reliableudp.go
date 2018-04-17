package rudp

import "udp"
import "github.com/woodywanghg/gofclog"
import "github.com/golang/protobuf/proto"
import "rudpproto"
import "time"
import "sync"

type ReliableUdp struct {
	encrypt    RudpEncrypt
	udpSocket  udpSocket.UdpSocket
	sessionMap map[int64]RecvBuff
	sessionMap map[int64]SendBuff
	lock       sync.Mutex
}

func (r *ReliableUdp) Init(ip string, port int) error {

	fclog.DEBUG("this=%p", r)
	r.encrypt.Init()
	r.sessionMap = make(map[int64]SessionBuff, 0)
	r.udpSocket.SetUdpReceiver(r)

	err := r.udpSocket.Listen(ip, port)
	if err != nil {
		fclog.ERROR("ReliableUdp init error! err=%s", err.Error())
		return err
	}

	return nil
}

func (r *ReliableUdp) OnUdpRecv(b []byte, bLen int, ip string, port int) {
	tempBuf := b[0:bLen]
	fclog.DEBUG("Recv data=%d byte=%s", bLen, string(tempBuf))

	if !r.encrypt.IsValidPacket(tempBuf) {
		fclog.ERROR("Invalid packet")
		return
	}

	packetData := r.encrypt.GetPacketData(tempBuf)

	r.DecodePacket(packetData)

	r.udpSocket.SendData(packetData, ip, port)

	fclog.DEBUG("Send ip=%s port=%d packetData=%v", ip, port, packetData)
}

func (r *ReliableUdp) DecodePacket(b []byte) {

	var msg rudpmsg.RudpMessage

	err := proto.Unmarshal(b, &msg)
	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	var msgType rudpmsg.RudpMsgType

	switch {

	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_DATA:
	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_ACK:
	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_REG:
	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_REG_RS:

	}

}

func (r *ReliableUdp) ProcessMsgData(b []byte) {

	var msgData rudpmsg.RudpMsgData
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

}

func (r *ReliableUdp) ProcessMsgAck(b []byte) {
	var msgData rudpmsg.RudpMsgAck
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}
}

func (r *ReliableUdp) ProcessMsgReg(b []byte) {
	var msgData rudpmsg.RudpMsgReg
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}
}

func (r *ReliableUdp) ProcessMsgRegRs(b []byte) {
	var msgData rudpmsg.RudpMsgRegRs
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	ssrc := int64(*msgData.Ssrc)
	var dataBuf SessionBuff
	dataBuf.Init(ssrc)

	r.lock.Lock()
	r.sessionMap[ssrc] = dataBuf
	r.lock.Unlock()

}

func (r *ReliableUdp) CreateSession() {

	var msg rudpmsg.RudpMsgReg
	msg.Seq = proto.Int64(0)
	msg.Ssrc = proto.Int64(0)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return
	}

	packetData := r.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return
	}

	encryptData := r.encrypt.EncodePacket(packetData)

	r.udpSocket.SendData(encryptData, r.udpSocket.GetIp(), r.udpSocket.GetPort())
}

func (r *ReliableUdp) EncodePacket(b []byte, msgType rudpmsg.RudpMsgType) []byte {

	var packet rudpmsg.RudpMessage
	packet.Type = &msgType
	packet.Data = b

	packetData, err := proto.Marshal(&packet)

	if err != nil {
		return make([]byte, 0)
	}

	return packetData
}

func (r *ReliableUdp) SendCreateSessionRs() {

	ssrc := time.Now().UnixNano()

	var msg rudpmsg.RudpMsgRegRs
	msg.Seq = proto.Int64(0)
	msg.Ssrc = proto.Int64(ssrc)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return
	}

	packetData := r.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return
	}

	var dataBuf SessionBuff
	dataBuf.Init(ssrc)

	r.lock.Lock()
	r.sessionMap[ssrc] = dataBuf
	r.lock.Unlock()

	encryptData := r.encrypt.EncodePacket(packetData)

	r.udpSocket.SendData(encryptData, r.udpSocket.GetIp(), r.udpSocket.GetPort())
}
