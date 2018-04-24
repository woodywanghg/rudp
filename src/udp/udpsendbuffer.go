package udpsocket

import "net"
import "sync"
import "github.com/woodywanghg/gofclog"

type BufferItem struct {
	Data    []byte
	DstAddr *net.UDPAddr
}

type UdpSendBuffer struct {
	bufferList []BufferItem
	lockBuff   sync.Mutex
}

func (p *UdpSendBuffer) Add(b []byte, dstAddr *net.UDPAddr) {
	p.lockBuff.Lock()
	defer p.lockBuff.Unlock()

	var item BufferItem = BufferItem{Data: b, DstAddr: dstAddr}
	p.bufferList = append(p.bufferList, item)

	fclog.DEBUG("buffer list len=%d addr=%v", len(p.bufferList), *dstAddr)
}

func (p *UdpSendBuffer) GetLength() int {
	p.lockBuff.Lock()
	defer p.lockBuff.Unlock()

	return len(p.bufferList)
}

func (p *UdpSendBuffer) GetData() []BufferItem {
	p.lockBuff.Lock()
	defer p.lockBuff.Unlock()

	bufferList := make([]BufferItem, len(p.bufferList))

	copy(bufferList, p.bufferList)

	fclog.DEBUG("bufferList=%v", bufferList)

	p.bufferList = make([]BufferItem, 0)

	return bufferList
}
