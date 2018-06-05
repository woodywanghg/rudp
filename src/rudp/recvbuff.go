package rudp

import "sync"
import "time"
import "sort"
import "math"
import "github.com/woodywanghg/gofclog"

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
	seqInts    []int
}

func (s *RecvBuff) Init(udpSession *UdpSession) {
	s.seq = 0
	s.nextSeq = 0
	s.udpSession = udpSession
	s.seqMap = make(map[int64]*RecvBuffItem, 100)
}

func (s *RecvBuff) Insert(seq int64, b []byte) bool {

	s.lock.Lock()
	defer s.lock.Unlock()

	if seq < s.nextSeq && math.Abs(float64(seq-s.nextSeq)) < (SEQ_MAX_INDEX-3000)*1.0 {
		fclog.ERROR("Find invalid timeout packet! seq=%d", seq)
		return false
	}

	_, have := s.seqMap[seq]
	if have {
		fclog.INFO("Find duplicate key. return!")
		return false
	}

	item := new(RecvBuffItem)
	item.data = b
	item.ts = time.Now().UnixNano()

	s.seqMap[seq] = item
	s.seqInts = append(s.seqInts, int(seq))
	sort.Ints(s.seqInts)
	fclog.DEBUG("Sort ints=%v map=%v", s.seqInts, s.seqMap)

	return true
}

func (s *RecvBuff) GetData() ([]byte, bool) {

	s.lock.Lock()
	defer s.lock.Unlock()

	var data []byte

	if len(s.seqInts) < 1 {
		return data, false
	}

	b, have := s.seqMap[s.nextSeq]
	fclog.DEBUG("sort int have=%v map=%v", have, s.seqMap)

	if have {
		fclog.DEBUG("Find index packet seq=%d", s.nextSeq)
		delete(s.seqMap, s.nextSeq)
		s.nextSeq = (s.nextSeq + 1) % SEQ_MAX_INDEX
		if len(s.seqInts) > 1 {
			s.seqInts = s.seqInts[1:]
		} else if len(s.seqInts) == 1 {
			s.seqInts = make([]int, 0)
		}

		return b.data, true
	}

	return data, false
}

func (s *RecvBuff) GetTimeoutData(curTs int64, timeoutSec int64) ([]byte, bool) {

	s.lock.Lock()
	defer s.lock.Unlock()

	var data []byte

	if len(s.seqInts) < 1 {
		return data, false
	}

	v := int64(s.seqInts[0])
	b, have := s.seqMap[v]

	if have {
		if curTs-b.ts > timeoutSec {
			fclog.DEBUG("Find invalid timeout packet seq=%d", s.nextSeq)
			delete(s.seqMap, s.nextSeq)
			s.nextSeq = (v + 1) % SEQ_MAX_INDEX

			if len(s.seqInts) > 1 {
				s.seqInts = s.seqInts[1:]
			} else if len(s.seqInts) == 1 {
				s.seqInts = make([]int, 0)
			}

			return b.data, true
		}
	}

	return data, false
}
