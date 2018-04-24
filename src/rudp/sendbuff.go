package rudp

import "sync"
import "time"
import "github.com/woodywanghg/gofclog"

type SendBuffItem struct {
	ts      int64
	data    []byte
	retrans int
}

type SendBuff struct {
	seqMap map[int64]SendBuffItem
	lock   sync.Mutex
}

func (s *SendBuff) Init() {
	s.seqMap = make(map[int64]SendBuffItem, 100)
}

func (s *SendBuff) Insert(b []byte, seq int64) {

	s.lock.Lock()
	defer s.lock.Unlock()

	ts := time.Now().UnixNano()
	item := SendBuffItem{ts: ts, data: b, retrans: 0}

	s.seqMap[seq] = item
}

func (s *SendBuff) Delete(seq int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.seqMap, seq)

	fclog.DEBUG("Delete send buffer data! seq=%d", seq)

}

func (s *SendBuff) GetBufferData() map[int64]SendBuffItem {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.seqMap
}
