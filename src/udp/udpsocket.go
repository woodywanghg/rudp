package udpsocket

import "net"
import "time"

import "github.com/woodywanghg/gofclog"

type UdpSocket struct {
	port       int
	ip         string
	conn       *net.UDPConn
	recv       UdpRecv
	buff       []byte
	sendBuffer UdpSendBuffer
}

func (u *UdpSocket) Listen(ip string, port int) error {

	u.buff = make([]byte, 1024)
	u.ip = ip
	u.port = port

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
	u.buff = make([]byte, 1024)
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP(ip), Port: port}

	var err error = nil
	u.conn, err = net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return err
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

		time.Sleep(1000000 * 1000)
		u.sendUdpDataToPeer()
	}
}

func (u *UdpSocket) sendUdpDataToPeer() {
	bufferList := u.sendBuffer.GetData()
	fclog.DEBUG("BufferList len=%d list=%v", len(bufferList), bufferList)
	for _, v := range bufferList {
		dstAddr := &net.UDPAddr{IP: net.ParseIP(v.Ip), Port: v.Port}
		fclog.DEBUG("try to send: ip=%s port=%d data=%v", v.Ip, v.Port, v.Data)
		sLen, err := u.conn.WriteToUDP(v.Data, dstAddr)
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

func (u *UdpSocket) SendData(b []byte, ip string, port int) {
	u.sendBuffer.Add(b, ip, port)
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
