package rudp

import "bytes"
import "github.com/woodywanghg/gofclog"

type RudpEncrypt struct {
	preKey   []byte
	endKey   []byte
	preLen   int
	endLen   int
	checkLen int
}

func (r *RudpEncrypt) Init() {
	r.preKey = []byte{0x31, 0x31, 0x31}
	r.endKey = []byte{0x32, 0x32, 0x32, 0x0A}

	r.preLen = len(r.preKey)
	r.endLen = len(r.endKey)
	r.checkLen = r.preLen + r.endLen
}

func (r *RudpEncrypt) IsValidPacket(b []byte) bool {

	packetLen := len(b)

	if packetLen < r.checkLen {
		fclog.ERROR("Invalid packet len")
		return false
	}

	preSub := b[0:r.preLen]
	endSub := b[packetLen-r.endLen:]

	fclog.DEBUG("preSub=%v endSub=%v packetLen=%d endLen=%d", preSub, endSub, packetLen, r.endLen)

	if !bytes.Equal(preSub, r.preKey) || !bytes.Equal(endSub, r.endKey) {
		fclog.ERROR("Invalid packet value")
		return false
	}

	return true
}

func (r *RudpEncrypt) EncodePacket(b []byte) []byte {

	encodeData := make([]byte, 0)

	encodeData = append(encodeData, r.preKey...)
	encodeData = append(encodeData, b...)
	encodeData = append(encodeData, r.endKey...)

	return encodeData
}

func (r *RudpEncrypt) GetPacketData(b []byte) []byte {
	packetLen := len(b)
	packetData := b[r.preLen : packetLen-r.endLen]

	fclog.DEBUG("GetPacketData data:%v", packetData)

	return packetData
}
