package main

import "os"
import "fmt"
import "rudp"
import "time"
import "github.com/woodywanghg/gofclog"
import "github.com/woodywanghg/goini"

type TestServer struct {
}

func (t *TestServer) OnSessionCreate(sessionId int64, code int) {
	fmt.Printf("OnSessionCreate  code=%d\n", code)
}

func (t *TestServer) OnRecv(sessionId int64, b []byte) {

	fmt.Printf("OnRecv data len=%d\n", len(b))
}

func (t *TestServer) OnSessionError(sessionId int64, errCode int) {
	fmt.Printf("OnRecv data session id=%d, code=%d\n", sessionId, errCode)
}

func main() {

	fclog.Init(true, true, "rudp.log", 1048576, fclog.LEVEL_DEBUG)

	var iniObj goini.IniFile
	if !iniObj.Init("./demo.ini") {
		os.Exit(0)
		return
	}

	serverIp := iniObj.ReadString("SERVER", "ip", "error")
	serverPort := iniObj.ReadInt("SERVER", "port", -1)

	clientIp := iniObj.ReadString("CLIENT", "ip", "error")
	clientPort := iniObj.ReadInt("CLIENT", "port", -1)

	var obj = rudp.GetReliableUdp()
	obj.Init()

	if serverIp != "error" && serverPort != -1 {
		err := obj.Listen("0.0.0.0", serverPort)

		if err != nil {
			fmt.Printf("Init server error! err=%s\n", err.Error())
			return
		}
	}

	var objTest TestServer
	obj.SetUdpInterface(&objTest)
	var sid int64 = 0
	var err error = nil
	fclog.DEBUG("clientip=%s port=%d", clientIp, clientPort)
	if clientIp != "error" && clientPort != -1 {
		sid, err = obj.CreateSession(clientIp, clientPort)
		fclog.DEBUG("CreateSession id=%d", sid)
		if err != nil {
			os.Exit(0)
			return
		}
	}

	index := 1000
	for {
		time.Sleep(1000000 * 200)
		if sid != 0 {
			buff := fmt.Sprintf("index=%d", index)
			index += 1
			obj.SendData(sid, []byte(buff))
		}

	}
}
