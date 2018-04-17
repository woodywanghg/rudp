package udpsocket

type UdpRecv interface {
	OnUdpRecv(b []byte, bLen int, ip string, port int)
}
