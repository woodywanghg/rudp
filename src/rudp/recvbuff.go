package rudp

import "sync"
import "time"

type RecvBuffItem struct {
	data []byte
	ts   int64
}

type RecvBuff struct {
	seq        int64
	seqMap     map[int64]*RecvBuffItem
	nextSeq    int64
	lock       sync.Mutex
	udpSession *UdpSession
}

func (s *RecvBuff) Init(udpSession *UdpSession) {
	s.seq = 0
	s.nextSeq = 0
	s.udpSession = udpSession
	s.seqMap = make(map[int64]*RecvBuffItem, 100)
}

func (s *RecvBuff) Insert(seq int64, b []byte) {

	s.lock.Lock()
	defer s.lock.Unlock()

	item := new(RecvBuffItem)
	item.data = b
	item.ts = time.Now().UnixNano()

	s.seqMap[seq] = item
}

func (s *RecvBuff) Delete(seq int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.seqMap, seq)
}

func (s *RecvBuff) GetData() ([]byte, bool) {

	s.lock.Lock()
	defer s.lock.Unlock()

	b, have := s.seqMap[s.nextSeq]
	if have {
		s.nextSeq = (s.nextSeq + 1) % SEQ_MAX_INDEX
	}

	return b.data, have
}
