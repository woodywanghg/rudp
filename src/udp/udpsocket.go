package udpsocket

import "net"

//import "time"
import "strings"
import "strconv"

import "github.com/woodywanghg/gofclog"

type UdpSocket struct {
	port       int
	ip         string
	conn       *net.UDPConn
	recv       UdpRecv
	buff       []byte
	sendBuffer UdpSendBuffer
	localIp    string
	localPort  int
	writeChan  chan bool
}

func (u *UdpSocket) Listen(ip string, port int) error {

	u.buff = make([]byte, 1024)
	u.ip = ip
	u.port = port
	u.localIp = ip
	u.localPort = port
	u.writeChan = make(chan bool)

	var err error = nil
	u.conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(u.ip), Port: u.port})
	if err != nil {
		return err
	}

	fclog.DEBUG("ListUDP OK. IP=%s PORT:%d", u.ip, u.port)

	go u.goRecv()
	go u.goSend()

	return nil
}

func (u *UdpSocket) DialUDP(ip string, port int) error {

	u.ip = ip
	u.port = port
	u.localIp = ""
	u.localPort = 0
	u.buff = make([]byte, 1024)
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP(ip), Port: port}

	var err error = nil
	u.conn, err = net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return err
	}

	localAddr := u.conn.LocalAddr().String()
	addrAry := strings.Split(localAddr, ":")
	if len(addrAry) > 1 {
		u.localIp = addrAry[0]
		u.localPort, _ = strconv.Atoi(addrAry[1])
	}

	go u.goRecv()
	go u.goSend()

	return nil
}

func (u *UdpSocket) goRecv() {

	for {
		rLen, addr, err := u.conn.ReadFromUDP(u.buff)
		if err != nil {
			fclog.ERROR("ReadFromUDP error! err=%s", err.Error())
			continue
		}

		fclog.DEBUG("recv=%p rLen=%d addr=%v", u.recv, rLen, addr)
		u.recv.OnUdpRecv(u.buff, rLen, addr.IP.String(), addr.Port)

	}
}

func (u *UdpSocket) goSend() {
	for {

		<-u.writeChan
		fclog.DEBUG("CHANNEL write")
		u.sendUdpDataToPeer()
	}
}

func (u *UdpSocket) sendUdpDataToPeer() {
	bufferList := u.sendBuffer.GetData()
	for _, v := range bufferList {
		sLen, err := u.conn.WriteToUDP(v.Data, v.DstAddr)
		if err != nil {
			fclog.ERROR("SendData error! err=%s, sLen=%d", err.Error(), sLen)
			continue
		}
	}
}

func (u *UdpSocket) SetUdpReceiver(recv UdpRecv) {
	u.recv = recv
	fclog.DEBUG("u.recv=%p", u.recv)
}

func (u *UdpSocket) SendData(b []byte, dstAddr *net.UDPAddr) {
	u.sendBuffer.Add(b, dstAddr)
	u.writeChan <- true
}

func (u *UdpSocket) SendCriticalData(b []byte, dstAddr *net.UDPAddr) {
	sLen, err := u.conn.WriteToUDP(b, dstAddr)
	if err != nil {
		fclog.ERROR("SendCriticalData error! err=%s, sLen=%d", err.Error(), sLen)
	}
}

func (u *UdpSocket) Close() {
	u.conn.Close()
}

func (u *UdpSocket) GetIp() string {
	return u.ip
}

func (u *UdpSocket) GetPort() int {
	return u.port
}

func (u *UdpSocket) GetLocalIp() string {
	return u.localIp
}

func (u *UdpSocket) GetLocalPort() int {
	return u.localPort
}
