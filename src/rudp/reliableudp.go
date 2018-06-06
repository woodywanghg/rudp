package rudp

import "udp"
import "net"
import "github.com/woodywanghg/gofclog"
import "github.com/golang/protobuf/proto"
import "rudpproto"
import "time"
import "sync"

const (
	UDP_SESSION_RS_OK  = 0
	UDP_SESSION_RS_ERR = 1
)

type ReliableUdp struct {
	encrypt     RudpEncrypt
	udpSocket   *udpsocket.UdpSocket
	lock        sync.Mutex
	sessionMap  map[int64]*UdpSession
	udpInter    RudpInter
	readChan    chan bool
	readTimeOut int
}

var rudp *ReliableUdp = nil

func GetReliableUdp() *ReliableUdp {
	if rudp == nil {
		rudp = new(ReliableUdp)
	}

	return rudp
}

func (r *ReliableUdp) Init() {

	r.udpSocket = nil
	r.encrypt.Init()
	r.sessionMap = make(map[int64]*UdpSession, 0)
	r.readChan = make(chan bool)
	r.readTimeOut = 2
}

func (r *ReliableUdp) SetUdpInterface(udpInter RudpInter) {
	r.udpInter = udpInter
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
	go r.sessionReadCheck()

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
	go r.sessionReadCheck()

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

	udpSession, exist := r.sessionMap[sid]
	if !exist {
		fclog.ERROR("Receive invalid data sid=%d seq=%d", sid, seq)
		return
	}

	udpSession.SendAck(seq)

	fclog.DEBUG("Receice udp data: seq=%d data='%s'", seq, string(data))

	insertOK := udpSession.OnDataRecv(seq, data)

	r.lock.Unlock()

	if insertOK {
		fclog.DEBUG("signal---->")
		r.readChan <- true
		fclog.DEBUG("signal OK---->")
	}
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

	udpSession.SendAck(seq)
	if !udpSession.SendRegisterRs() {

		fclog.ERROR("SendRegisterRs error!")
	}

	fclog.DEBUG("SendRegisterRs OK. ID=%d create", sid)

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
		r.udpInter.OnSessionCreate(sid, UDP_SESSION_RS_ERR)
		return
	}

	udpSession.SendAck(seq)

	fclog.DEBUG("Receive udp session response sessionid=%d code=%d", sid, code)

	r.udpInter.OnSessionCreate(sid, UDP_SESSION_RS_OK)
}

func (r *ReliableUdp) CreateSession(ip string, port int) (int64, error) {

	sid := time.Now().UnixNano()

	var udpSession *UdpSession = new(UdpSession)
	udpSession.Init(sid, ip, port, r.udpSocket, r)

	r.lock.Lock()
	r.sessionMap[sid] = udpSession
	r.lock.Unlock()

	err := udpSession.SendRegister(sid)

	return sid, err
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
		time.Sleep(1000000 * 1000)
		r.lock.Lock()
		for _, session := range r.sessionMap {
			session.RetransmissionCheck()
		}
		r.lock.Unlock()

	}
}

func (r *ReliableUdp) sessionReadCheck() {

	for {
		<-r.readChan
		r.lock.Lock()
		for sid, session := range r.sessionMap {

			for {
				data, bHave := session.ReadCheck()
				if bHave {
					fclog.DEBUG("Find sequence packet")
					r.udpInter.OnRecv(sid, data)
				} else {
					fclog.DEBUG("Break CHECK")
					break
				}
			}
		}
		r.lock.Unlock()
		fclog.DEBUG("Event fire check")

	}
}

func (r *ReliableUdp) sessionReadTimeoutCheck() {

	for {
		select {

		case <-time.After(time.Duration(r.readTimeOut) * time.Second):
			r.lock.Lock()
			for sid, session := range r.sessionMap {

				for {
					data, bHave := session.ReadTimeoutCheck()
					if bHave {
						fclog.DEBUG("find time out packet")
						r.udpInter.OnRecv(sid, data)
					} else {
						break
					}
				}
			}
			r.lock.Unlock()
			fclog.DEBUG("time out check")

		}
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

	udpSession.SendData(b)
}

func (r *ReliableUdp) GetEncrypt() *RudpEncrypt {
	return &r.encrypt
}
