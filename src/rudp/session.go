package rudp

import "udp"
import "github.com/woodywanghg/gofclog"

type UdpSession struct {
	sendBuf         SendBuff
	recvBuf         RecvBuff
	sessionId       int64
	dIp             string
	dPort           int
	recv            udpsocket.UdpRecv
	udpSocket       *udpsocket.UdpSocket
	reliableUdp     *ReliableUdp
	sendSeq         int64
	retransCount    int
	retransInterval int
}

func (s *UdpSession) Init(sessionId int64, dIp string, dPort int, udpSocket *udpsocket.UdpSocket, reliableUdp *ReliableUdp) {
	s.dIp = dIp
	s.dPort = dPort
	s.sessionId = sessionId
	s.retransCount = -1
	s.retransInterval = 100
	s.reliableUdp = reliableUdp
	s.sendSeq = 0
	s.sendBuf.Init()
	s.recvBuf.Init()
	s.udpSocket = udpSocket

}

func (s *UdpSession) Close() {

}

func (s *UdpSession) OnUdpRecv(b []byte, bLen int, ip string, port int) {
	s.reliableUdp.OnUdpRecv(b, bLen, ip, port)
}

func (s *UdpSession) SendData(b []byte) {
	s.sendBuf.Insert(b, s.sendSeq)
	s.sendSeq = (s.sendSeq + 1) % SEQ_MAX_INDEX

	s.udpSocket.SendData(b, s.dIp, s.dPort)
}

func (s *UdpSession) GetSid() int64 {
	return s.sessionId
}

func (s *UdpSession) OnAck(sid int64) {
	s.sendBuf.Delete(sid)
}

func (s *UdpSession) SetMaxRetransmissionCount(count int) {
	s.retransCount = count
	fclog.DEBUG("SetMaxRetransmissionCount count=%d", count)
}

func (s *UdpSession) SetRetransmissionInterval(usecond int) {
	s.retransInterval = usecond
	fclog.DEBUG("SetRetransmissionInterval interval=%d", usecond)
}

func (s *UdpSession) RetransmissionCheck() {
	if s.retransCount == 0 {
		return
	}

}
