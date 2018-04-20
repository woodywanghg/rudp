package rudp

import "udp"
import "github.com/woodywanghg/gofclog"

type UdpSession struct {
	sendBuf   SendBuff
	recvBuf   RecvBuff
	sessionId int64
	dIp       string
	dPort     int
	recv      udpsocket.UdpRecv
	udpSocket *udpsocket.UdpSocket
}

func (s *UdpSession) Init(sessionId int64, dIp string, dPort int) bool {
	s.dIp = dIp
	s.dPort = dPort
	s.sessionId = sessionId
	s.sendBuf.Init()
	s.recvBuf.Init()
	s.udpSocket = nil

	return true
}

func (s *UdpSession) DialUDP() error {
	s.udpSocket = new(udpsocket.UdpSocket)
	err := s.udpSocket.DialUDP(s.dIp, s.dPort)
	if err != nil {
		fclog.ERROR("DIALUDP error! err=%s", err.Error())
		return err
	}

	return nil
}

func (s *UdpSession) Close() {

	if s.udpSocket != nil {
		s.udpSocket.Close()
	}
}

func (u *UdpSession) SetUdpReceiver(recv udpsocket.UdpRecv) {
	u.recv = recv
}

func (u *UdpSession) GetSocket() *udpsocket.UdpSocket {
	return u.udpSocket
}
