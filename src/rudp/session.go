package rudp

import "udp"
import "github.com/woodywanghg/gofclog"

type UdpSession struct {
	sendBuf     SendBuff
	recvBuf     RecvBuff
	sessionId   int64
	dIp         string
	dPort       int
	recv        udpsocket.UdpRecv
	udpSocket   *udpsocket.UdpSocket
	bDial       bool
	reliableUdp *ReliableUdp
	sendSeq     int64
}

func (s *UdpSession) Init(sessionId int64, dIp string, dPort int, udpSocket *udpsocket.UdpSocket, reliableUdp *ReliableUdp) bool {
	s.dIp = dIp
	s.dPort = dPort
	s.sessionId = sessionId
	s.reliableUdp = reliableUdp
	s.sendSeq = 0
	s.sendBuf.Init()
	s.recvBuf.Init()
	if udpSocket == nil {
		s.bDial = true
		if s.dialUDP() != nil {
			return false
		}
	}

	return true
}

func (s *UdpSession) dialUDP() error {
	s.udpSocket = new(udpsocket.UdpSocket)
	err := s.udpSocket.DialUDP(s.dIp, s.dPort)
	if err != nil {
		fclog.ERROR("DIALUDP error! err=%s", err.Error())
		return err
	}

	s.udpSocket.SetUdpReceiver(s)

	return nil
}

func (s *UdpSession) Close() {

	if s.bDial {
		s.udpSocket.Close()
	}
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
