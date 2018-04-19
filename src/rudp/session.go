package rudp

type UdpSession struct {
	sendBuf SendBuff
	recvBuf RecvBuff
}

func (s *UdpSession) Init(sessionId int64) {

}
