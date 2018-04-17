package udpsocket

import "sync"
import "github.com/woodywanghg/gofclog"

type BufferItem struct {
	Data []byte
	Ip   string
	Port int
}

type UdpSendBuffer struct {
	bufferList []BufferItem
	lockBuff   sync.Mutex
}

func (p *UdpSendBuffer) Add(b []byte, ip string, port int) {
	p.lockBuff.Lock()
	defer p.lockBuff.Unlock()

	var item BufferItem = BufferItem{Data: b, Ip: ip, Port: port}
	p.bufferList = append(p.bufferList, item)

	fclog.DEBUG("buffer list=%d ip=%s port=%d b=%v", len(p.bufferList), ip, port, b)
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
