package rudp

import "udp"
import "time"
import "net"
import "rudpproto"
import "github.com/woodywanghg/gofclog"
import "github.com/golang/protobuf/proto"

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
	retransInterval int64
	dstAddr         net.UDPAddr
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
	s.dstAddr = net.UDPAddr{IP: net.ParseIP(dIp), Port: dPort}

}

func (s *UdpSession) Close() {

}

func (s *UdpSession) OnUdpRecv(b []byte, bLen int, ip string, port int) {
	s.reliableUdp.OnUdpRecv(b, bLen, ip, port)
}

func (s *UdpSession) SendData(b []byte) {
	s.sendBuf.Insert(b, s.sendSeq)
	s.sendSeq = (s.sendSeq + 1) % SEQ_MAX_INDEX

	s.udpSocket.SendData(b, &s.dstAddr)
}

func (s *UdpSession) SendAckData(b []byte) {
	s.udpSocket.SendCriticalData(b, &s.dstAddr)
}

func (s *UdpSession) GetSid() int64 {
	return s.sessionId
}

func (s *UdpSession) OnAck(seq int64) {
	s.sendBuf.Delete(seq)
}

func (s *UdpSession) SetMaxRetransmissionCount(count int) {
	s.retransCount = count
	fclog.DEBUG("SetMaxRetransmissionCount count=%d", count)
}

func (s *UdpSession) SetRetransmissionInterval(usecond int) {
	s.retransInterval = int64(usecond)
	fclog.DEBUG("SetRetransmissionInterval interval=%d", usecond)
}

func (s *UdpSession) RetransmissionCheck() {
	if s.retransCount == 0 {
		return
	}

	curTs := time.Now().UnixNano()
	bufferData := s.sendBuf.GetBufferData()

	for seq, v := range bufferData {
		if curTs-v.ts >= s.retransInterval {

			s.udpSocket.SendCriticalData(v.data, &s.dstAddr)
			v.retrans += 1
			fclog.DEBUG("Ack timeout retransmission interval=%d seq=%d retrans count=%d", s.retransInterval, seq, v.retrans)

			if s.retransCount > 0 {
				if v.retrans > s.retransCount {
					fclog.INFO("Packet invalid! rm packet.  max retransmission limit=%d", v.retrans)
					s.sendBuf.Delete(seq)
				}
			}
		}
	}

}

func (s *UdpSession) SendPacketData(b []byte) {
	s.sendBuf.Insert(b, s.sendSeq)
	s.sendSeq = (s.sendSeq + 1) % SEQ_MAX_INDEX

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

	s.udpSocket.SendData(encryptData, &s.dstAddr)
}
