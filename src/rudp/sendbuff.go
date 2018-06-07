package rudp

import "time"
import "github.com/woodywanghg/gofclog"

type SendBuffItem struct {
	ts      int64
	data    []byte
	retrans int
}

type SendBuff struct {
	udpSession   *UdpSession
	seqMap       map[int64]*SendBuffItem
	retransCount int64
}

func (s *SendBuff) Init(udpSession *UdpSession) {
	s.retransCount = 0
	s.udpSession = udpSession
	s.seqMap = make(map[int64]*SendBuffItem, 100)
}

func (s *SendBuff) Insert(b []byte, seq int64) {

	ts := time.Now().UnixNano()
	item := new(SendBuffItem)
	item.ts = ts
	item.data = b
	item.retrans = 0

	s.seqMap[seq] = item

	fclog.DEBUG("Send buffer seq=%d len=%d", seq, len(s.seqMap))

	for k, _ := range s.seqMap {
		fclog.DEBUG("----->key=%d", k)
	}
}

func (s *SendBuff) Delete(seq int64) {

	delete(s.seqMap, seq)

	fclog.DEBUG("Delete send buffer data! seq=%d", seq)

}

func (s *SendBuff) Check() {
	if s.udpSession.GetRetransCount() == 0 {
		return
	}

	curTs := time.Now().UnixNano()

	fclog.DEBUG("Check......")
	for seq, v := range s.seqMap {
		fclog.DEBUG("curTs=%d v.ts=%d sub=%d interval=%d", curTs, v.ts, curTs-v.ts, s.udpSession.GetRetransInterval())
		if curTs-v.ts >= s.udpSession.GetRetransInterval() {

			s.retransCount += 1
			v.retrans += 1
			s.udpSession.SendRetransData(v.data)
			fclog.DEBUG("Ack timeout retransmission interval=%d seq=%d retrans count=%d", s.udpSession.GetRetransInterval(), seq, v.retrans)

			if s.udpSession.GetRetransCount() > 0 {
				if v.retrans > s.udpSession.GetRetransCount() {
					fclog.INFO("Packet invalid! rm packet.  max retransmission limit=%d", v.retrans)
					delete(s.seqMap, seq)
				}
			}
		}
	}
	fclog.DEBUG("Check out......")
}

func (s *SendBuff) GetRetransCount() int64 {
	return s.retransCount
}
