package rudp

import "sync"
import "time"

type SendBuffItem struct {
	ts      int64
	data    []byte
	retrans int
}

type SendBuff struct {
	seq    int64
	seqMap map[int64]SendBuffItem
	lock   sync.Mutex
}

func (s *SendBuff) Init() {
	s.seq = 0
	s.seqMap = make(map[int64]SendBuffItem, 100)
}

func (s *SendBuff) Insert(b []byte) {

	s.lock.Lock()
	defer s.lock.Unlock()

	ts := time.Now().UnixNano()
	item := SendBuffItem{ts: ts, data: b, retrans: 0}

	s.seqMap[s.seq] = item

	s.seq = (s.seq + 1) % SEQ_MAX_INDEX
}

func (s *SendBuff) Delete(seq int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.seqMap, seq)
}
