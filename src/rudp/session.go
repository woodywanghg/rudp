package rudp

import "udp"

import "time"
import "net"
import "rudpproto"
import "github.com/woodywanghg/gofclog"
import "github.com/golang/protobuf/proto"

type UdpSession struct {
	sendBuf            SendBuff
	recvBuf            RecvBuff
	sessionId          int64
	dIp                string
	dPort              int
	recv               udpsocket.UdpRecv
	udpSocket          *udpsocket.UdpSocket
	reliableUdp        *ReliableUdp
	sendSeq            int64
	retransCount       int
	retransInterval    int64
	readTimeout        int64
	dstAddr            net.UDPAddr
	lossRate           int
	retransmissionRate int
	statSendCount      int64
	statAckCount       int64
}

func (s *UdpSession) Init(sessionId int64, dIp string, dPort int, udpSocket *udpsocket.UdpSocket, reliableUdp *ReliableUdp) {
	s.dIp = dIp
	s.dPort = dPort
	s.sessionId = sessionId
	s.retransCount = -1
	s.retransInterval = 100
	s.readTimeout = 2000
	s.reliableUdp = reliableUdp
	s.sendSeq = 0
	s.sendBuf.Init(s)
	s.recvBuf.Init(s)
	s.udpSocket = udpSocket
	s.dstAddr = net.UDPAddr{IP: net.ParseIP(dIp), Port: dPort}
	s.lossRate = 0
	s.retransmissionRate = 0
	s.statSendCount = 0
	s.statAckCount = 0
}

func (s *UdpSession) Close() {

}

func (s *UdpSession) OnUdpRecv(b []byte, bLen int, ip string, port int) {
	s.reliableUdp.OnUdpRecv(b, bLen, ip, port)
}

func (s *UdpSession) SendAckData(b []byte) {
	s.udpSocket.SendCriticalData(b, &s.dstAddr)
}

func (s *UdpSession) GetSid() int64 {
	return s.sessionId
}

func (s *UdpSession) OnAck(seq int64) {
	s.sendBuf.Delete(seq)
	s.statAckCount += 1
}

func (s *UdpSession) SetMaxRetransmissionCount(count int) {
	s.retransCount = count
	fclog.DEBUG("SetMaxRetransmissionCount count=%d", count)
}

func (s *UdpSession) SetRetransmissionInterval(usecond int) {
	s.retransInterval = int64(usecond * 1000000)
	fclog.DEBUG("SetRetransmissionInterval interval=%d", usecond)
}

func (s *UdpSession) SetReadTimeout(usecond int) {
	s.readTimeout = int64(usecond * 1000000)
	fclog.DEBUG("SetRetransmissionInterval interval=%d", usecond)
}

func (s *UdpSession) RetransmissionCheck() {
	s.sendBuf.Check()
}

func (s *UdpSession) SendData(b []byte) {

	var msg rudpmsg.RudpMsgData
	msg.Seq = proto.Int64(s.sendSeq)
	msg.Sid = proto.Int64(s.sessionId)
	msg.Data = b

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_DATA)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return
	}

	encryptData := s.reliableUdp.GetEncrypt().EncodePacket(packetData)

	s.sendBuf.Insert(encryptData, s.sendSeq)
	s.sendSeq = (s.sendSeq + 1) % SEQ_MAX_INDEX
	fclog.DEBUG("SendData seq++")

	s.udpSocket.SendData(encryptData, &s.dstAddr)

	s.statSendCount += 1
}

func (s *UdpSession) SendAck(seq int64) {

	var msg rudpmsg.RudpMsgAck
	msg.Seq = proto.Int64(seq)
	msg.Sid = proto.Int64(s.sessionId)

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

	encryptData := s.reliableUdp.GetEncrypt().EncodePacket(packetData)
	s.SendAckData(encryptData)
}

func (s *UdpSession) SendRegister(sessionId int64) error {

	var msg rudpmsg.RudpMsgReg
	msg.Seq = proto.Int64(s.sendSeq)
	msg.Sid = proto.Int64(sessionId)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return err
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return err
	}

	encryptData := s.reliableUdp.GetEncrypt().EncodePacket(packetData)

	s.sendBuf.Insert(encryptData, s.sendSeq)

	s.udpSocket.SendData(encryptData, &s.dstAddr)

	return nil
}

func (s *UdpSession) SendRegisterRs() bool {

	var msg rudpmsg.RudpMsgRegRs
	msg.Seq = proto.Int64(0)
	msg.Sid = proto.Int64(s.sessionId)
	msg.Code = proto.Int64(0)

	data, err := proto.Marshal(&msg)
	if err != nil {
		fclog.ERROR("Marshal message error!")
		return false
	}

	packetData := rudpmsg.EncodePacket(data, rudpmsg.RudpMsgType_MSG_RUDP_REG_RS)

	if len(packetData) <= 0 {
		fclog.ERROR("EncodePacket error!")
		return false
	}

	encryptData := s.reliableUdp.GetEncrypt().EncodePacket(packetData)

	s.sendBuf.Insert(encryptData, s.sendSeq)

	s.udpSocket.SendData(encryptData, &s.dstAddr)

	return true
}

func (s *UdpSession) GetRetransCount() int {
	return s.retransCount
}

func (s *UdpSession) GetRetransInterval() int64 {
	return s.retransInterval
}

func (s *UdpSession) SendRetransData(b []byte) {
	s.udpSocket.SendCriticalData(b, &s.dstAddr)
}

func (s *UdpSession) OnDataRecv(seq int64, b []byte) bool {
	return s.recvBuf.Insert(seq, b)
}

func (s *UdpSession) ReadCheck() (b []byte, bRead bool) {
	return s.recvBuf.GetData()
}

func (s *UdpSession) ReadTimeoutCheck() (b []byte, bRead bool) {
	curTs := time.Now().UnixNano()
	return s.recvBuf.GetTimeoutData(curTs, s.readTimeout*1000000)
}

func (s *UdpSession) GetLossrate() int {

	s.lossRate = int((1 - float64(s.statAckCount)/float64(s.statSendCount+s.sendBuf.GetRetransCount())) * 100)
	fclog.DEBUG("Ackcount=%d statSendCount=%d lossrate=%d", s.statAckCount, s.statSendCount+s.sendBuf.GetRetransCount(), s.lossRate)

	return s.lossRate
}

func (s *UdpSession) GetRetransmissionrate() int {
	s.retransmissionRate = int((float64(s.sendBuf.GetRetransCount()) / float64(s.statSendCount)) * 100)
	fclog.DEBUG("retrans cout=%d sendcount=%d retransrate=%d", s.sendBuf.GetRetransCount(), s.statSendCount, s.retransmissionRate)
	return s.retransmissionRate
}
