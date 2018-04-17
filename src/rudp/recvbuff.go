package rudp

import "sync"

type RecvBuffItem struct {
	have bool
	data []byte
}

type RecvBuff struct {
	sessionId int64
	seq       int64
	seqMap    map[int64]RecvBuffItem
	nextSeq   int64
	lock      sync.Mutex
}

func (s *RecvBuff) Init(sessionId int64) {
	s.seq = 0
	s.nextSeq = 0
	s.seqMap = make(map[int64]RecvBuffItem, 100)
	s.sessionId = sessionId
}

func (s *RecvBuff) Insert(seq int64, b []byte) {

	s.lock.Lock()
	defer s.lock.Unlock()

	item := RecvBuffItem{have: true, data: b}

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
