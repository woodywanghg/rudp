package rudp

import "github.com/woodywanghg/gofclog"

type UdpSession struct {
	sendBuf   SendBuff
	recvBuf   RecvBuff
	sessionId int64
	dIp       string
	dPort     int64
	recv      UdpRecv
	udpSocket *udpsocket.UdpSocket
}

func (s *UdpSession) Init(sessionId int64, dIp string, dPort int, bClient bool) bool {
	s.dIp = dIp
	s.dPort = dPort
	s.sessionId = sessionId
	s.sendBuf.Init()
	s.recvBuf.Init()
	s.udpSocket = nil

	if bClient {
		s.udpSocket = new(udpsocket.UdpSocket)
		err := s.udpSocket.DialUDP(dIp, dPort)
		if err != nil {
			fclog.ERROR("DIALUDP error! err=%s", err.Error())
			return false
		}
	}

	return true
}

func (s *UdpSession) Close() {

	if s.udpSocket != nil {
		s.udpSocket.Close()
	}
}

func (u *UdpSession) SetUdpReceiver(recv UdpRecv) {
	u.recv = recv
}

func (u *UdpSession) UpdateSessionId(recv UdpRecv) {
	u.recv = recv
}
