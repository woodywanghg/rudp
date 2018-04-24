package rudp

import "udp"
import "net"
import "github.com/woodywanghg/gofclog"
import "github.com/golang/protobuf/proto"
import "rudpproto"
import "time"
import "sync"

type ReliableUdp struct {
	encrypt    RudpEncrypt
	udpSocket  *udpsocket.UdpSocket
	lock       sync.Mutex
	sessionMap map[int64]*UdpSession
}

func (r *ReliableUdp) Init() {

	r.udpSocket = nil
	r.encrypt.Init()
	r.sessionMap = make(map[int64]*UdpSession, 0)
}

func (r *ReliableUdp) Listen(ip string, port int) error {

	r.udpSocket = new(udpsocket.UdpSocket)
	r.udpSocket.SetUdpReceiver(r)
	err := r.udpSocket.Listen(ip, port)
	if err != nil {
		fclog.ERROR("ReliableUdp init error! err=%s", err.Error())
		return err
	}

	go r.sessionRetransmissionCheck()

	return nil
}

func (r *ReliableUdp) DialUDP(ip string, port int) error {
	r.udpSocket = new(udpsocket.UdpSocket)
	r.udpSocket.SetUdpReceiver(r)
	err := r.udpSocket.DialUDP(ip, port)
	if err != nil {
		fclog.ERROR("ReliableUdp dial udp error!")
		return err
	}

	go r.sessionRetransmissionCheck()

	return nil
}

func (r *ReliableUdp) OnUdpRecv(b []byte, bLen int, ip string, port int) {
	tempBuf := b[0:bLen]
	fclog.DEBUG("Recv data=%d byte=%v", bLen, tempBuf)

	if !r.encrypt.IsValidPacket(tempBuf) {
		fclog.ERROR("Invalid packet")
		return
	}

	packetData := r.encrypt.GetPacketData(tempBuf)

	r.decodePacket(packetData, ip, port)

	//r.udpSocket.SendData(packetData, ip, port)

	//fclog.DEBUG("Send ip=%s port=%d packetData=%v", ip, port, packetData)
}

func (r *ReliableUdp) decodePacket(b []byte, ip string, port int) {

	var msg rudpmsg.RudpMessage

	err := proto.Unmarshal(b, &msg)
	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	var msgType rudpmsg.RudpMsgType = *msg.Type
	fclog.DEBUG("PACKET TYPE=%d", int32(msgType))

	switch {

	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_DATA:
		r.processMsgData(msg.Data, ip, port)
	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_ACK:
		r.processMsgAck(msg.Data, ip, port)
	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_REG:
		r.processMsgReg(msg.Data, ip, port)
	case msgType == rudpmsg.RudpMsgType_MSG_RUDP_REG_RS:
		r.processMsgRegRs(msg.Data, ip, port)
	}

}

func (r *ReliableUdp) processMsgData(b []byte, ip string, port int) {

	fclog.DEBUG("ProcessMsgData")

	var msgData rudpmsg.RudpMsgData
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	sid := int64(*msgData.Sid)
	seq := int64(*msgData.Seq)
	data := msgData.Data

	r.lock.Lock()
	defer r.lock.Unlock()

	udpSession, exist := r.sessionMap[sid]
	if !exist {
		fclog.ERROR("Receive invalid data sid=%d seq=%d", sid, seq)
		return
	}

	r.sendAck(udpSession, seq)

	fclog.DEBUG("Receice udp data: %v", data)

}

func (r *ReliableUdp) processMsgAck(b []byte, ip string, port int) {

	fclog.DEBUG("processMsgAck")

	var msgData rudpmsg.RudpMsgAck
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	sid := int64(*msgData.Sid)
	seq := int64(*msgData.Seq)

	r.lock.Lock()
	defer r.lock.Unlock()

	udpSession, exist := r.sessionMap[sid]
	if !exist {
		fclog.ERROR("Receive invalid data sid=%d seq=%d", sid, seq)
		return
	}

	udpSession.OnAck(seq)
}

func (r *ReliableUdp) processMsgReg(b []byte, ip string, port int) {

	fclog.DEBUG("processMsgReg")

	var msgData rudpmsg.RudpMsgReg
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	sid := int64(*msgData.Sid)
	seq := int64(*msgData.Seq)

	r.lock.Lock()
	_, exist := r.sessionMap[sid]
	r.lock.Unlock()

	if exist {
		fclog.ERROR("Register error!, exist sessioin id! id=%d", sid)
		r.sendInvalidSessionRs(sid, ip, port)
		return
	}

	var udpSession *UdpSession = new(UdpSession)
	udpSession.Init(sid, ip, port, r.udpSocket, r)
	r.sessionMap[sid] = udpSession

	r.sendAck(udpSession, seq)
	r.sendCreateSessionRs(udpSession)

	fclog.DEBUG("Session ID=%d create", sid)

}

func (r *ReliableUdp) processMsgRegRs(b []byte, ip string, port int) {

	fclog.DEBUG("processMsgRegRs")

	var msgData rudpmsg.RudpMsgRegRs
	err := proto.Unmarshal(b, &msgData)

	if err != nil {
		fclog.ERROR("Unmarshal error! err=%s", err.Error())
		return
	}

	sid := int64(*msgData.Sid)
	seq := int64(*msgData.Seq)
	code := int64(*msgData.Code)

	r.lock.Lock()
	defer r.lock.Unlock()

	udpSession, exist := r.sessionMap[sid]
	if !exist {
		fclog.ERROR("Receive invalid data sid=%d seq=%d", sid, seq)
		return
	}

	r.sendAck(udpSession, seq)

	fclog.DEBUG("Receive udp session response sessionid=%d code=%d", sid, code)
}

func (r *ReliableUdp) CreateSession(ip string, port int) (int64, error) {

	Sid := time.Now().UnixNano()

	var msg rudpmsg.RudpMsgReg
	msg.Seq = proto.Int64(0)
	msg.Sid = proto.Int64(Sid)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return 0, err
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return 0, err
	}

	var udpSession *UdpSession = new(UdpSession)
	udpSession.Init(Sid, ip, port, r.udpSocket, r)

	r.lock.Lock()
	r.sessionMap[Sid] = udpSession
	r.lock.Unlock()

	encryptData := r.encrypt.EncodePacket(packetData)

	udpSession.SendData(encryptData)

	return Sid, nil
}

func (r *ReliableUdp) sendCreateSessionRs(udpSession *UdpSession) {

	var msg rudpmsg.RudpMsgRegRs
	msg.Seq = proto.Int64(0)
	msg.Sid = proto.Int64(udpSession.GetSid())
	msg.Code = proto.Int64(0)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG_RS)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return
	}

	encryptData := r.encrypt.EncodePacket(packetData)
	udpSession.SendData(encryptData)
}

func (r *ReliableUdp) sendInvalidSessionRs(sid int64, ip string, port int) {

	var msg rudpmsg.RudpMsgRegRs
	msg.Seq = proto.Int64(0)
	msg.Sid = proto.Int64(sid)
	msg.Code = proto.Int64(10001)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG_RS)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return
	}

	encryptData := r.encrypt.EncodePacket(packetData)
	dstAddr := &net.UDPAddr{IP: net.ParseIP(ip), Port: port}
	r.udpSocket.SendData(encryptData, dstAddr)
}

func (r *ReliableUdp) sendAck(udpSession *UdpSession, seq int64) {

	var msg rudpmsg.RudpMsgAck
	msg.Seq = proto.Int64(seq)
	msg.Sid = proto.Int64(udpSession.GetSid())

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_ACK)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return
	}

	encryptData := r.encrypt.EncodePacket(packetData)
	udpSession.SendAckData(encryptData)
}

func (r *ReliableUdp) SetMaxRetransmissionCount(sessionId int64, count int) {

	r.lock.Lock()
	defer r.lock.Unlock()

	udpSession, exist := r.sessionMap[sessionId]
	if !exist {
		fclog.ERROR("SetMaxRetransmissionCount error! sid=%d count=%d", sessionId, count)
		return
	}

	udpSession.SetMaxRetransmissionCount(count)
}

func (r *ReliableUdp) SetRetransmissionInterval(sessionId int64, usecond int) {

	r.lock.Lock()
	defer r.lock.Unlock()

	udpSession, exist := r.sessionMap[sessionId]
	if !exist {
		fclog.ERROR("SetRetransmissionInterval error! sid=%d count=%d", sessionId, usecond)
		return
	}

	udpSession.SetRetransmissionInterval(usecond)
}

func (r *ReliableUdp) sessionRetransmissionCheck() {

	for {
		time.Sleep(1000000 * 100000)
		r.lock.Lock()
		for _, session := range r.sessionMap {
			session.RetransmissionCheck()
		}
		r.lock.Unlock()

	}
}

func (r *ReliableUdp) SendData(sessionId int64, b []byte) {
	r.lock.Lock()
	defer r.lock.Unlock()

	udpSession, exist := r.sessionMap[sessionId]
	if !exist {
		fclog.ERROR("SendData error! sid=%d", sessionId)
		return
	}

	udpSession.SendPacketData(b)
}

func (r *ReliableUdp) GetEncrypt() *RudpEncrypt {
	return &r.encrypt
}
