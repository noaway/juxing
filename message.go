package main

import (
	"fmt"
	"time"
	"sync"

	"github.com/noaway/juxing/internal/utils"
	"github.com/noaway/juxing/live"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
)

var once sync.Once

// handleMessages handles messages
func handleMessages(w *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	rid:=m.Name
	once.Do(func(){go onMsg(w)})
	
	if utils.StrTo(rid).MustInt() == 0 {
		live.EventChan <- live.EventMessage{Type: "error", Msg: fmt.Sprintf("%v",`直播间id输入错误
例子: 如果你的直播间地址是: http://jx.kuwo.cn/123456
那么你的id就是123456`)}
		return 
	}
	if live.Cancel!=nil{
		time.Sleep(time.Second)
		live.Cancel()
	}
	live.StartMonitor(rid)
	return
}

func onMsg(w *astilectron.Window){
	for event:=range live.EventChan{
		w.SendMessage(event)
	}
}