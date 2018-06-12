package main

import "os"
import "fmt"
import "rudp"
import "time"
import "github.com/woodywanghg/gofclog"
import "github.com/woodywanghg/goini"
import "net/http"
import _ "net/http/pprof"

type TestClient struct {
}

func (t *TestClient) OnSessionCreate(sessionId int64, code int) {
	fmt.Printf("OnSessionCreate  code=%d\n", code)
}

func (t *TestClient) OnRecv(sessionId int64, b []byte) {

	fmt.Printf("OnRecv data len=%d\n", len(b))
}

func (t *TestClient) OnSessionError(sessionId int64, errCode int) {
	fmt.Printf("OnRecv data session id=%d, code=%d\n", sessionId, errCode)
}

func main() {

	go func() {
		fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	fclog.Init(true, true, "rudp.log", 1048576, fclog.LEVEL_DEBUG)

	var iniObj goini.IniFile
	if !iniObj.Init("./client.ini") {
		os.Exit(0)
		return
	}

	serverIp := iniObj.ReadString("SERVER", "ip", "error")
	serverPort := iniObj.ReadInt("SERVER", "port", -1)

	clientIp := iniObj.ReadString("CLIENT", "ip", "error")
	clientPort := iniObj.ReadInt("CLIENT", "port", -1)
	statAddr := iniObj.ReadString("STAT", "addr", "error")

	var obj = rudp.GetReliableUdp()
	obj.Init()

	if serverIp != "error" && serverPort != -1 {
		err := obj.Listen("0.0.0.0", serverPort)

		if err != nil {
			fmt.Printf("Init server error! err=%s\n", err.Error())
			return
		}
	}

	obj.Stat(statAddr)

	var objTest TestClient
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
		time.Sleep(1000000 * 20)
		if sid != 0 {
			buff := fmt.Sprintf("index=%d", index)
			index += 1
			obj.SendData(sid, []byte(buff))
		}

	}
}
