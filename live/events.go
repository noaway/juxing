package live

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/noaway/juxing/internal/utils"
	"github.com/noaway/juxing/store"
)

// EventMessage struct
type EventMessage struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// EventChan chan
var EventChan = make(chan EventMessage, 1000)

type eventFunc func(cmd string, n *Node)

func newEvent() *event { return &event{register: make(map[string]entrust)} }

type entrust struct {
	fn     eventFunc
	effect bool
}

type event struct {
	register map[string]entrust
}

func (e *event) notify(cmd string, node *Node) {
	for _, c := range []string{cmd, "*"} {
		if entrust, ok := e.register[c]; ok || entrust.effect {
			entrust.fn(cmd, node)
		}
	}
}

func (e *event) listen(cmd string, fn eventFunc) { e.register[cmd] = entrust{fn: fn, effect: true} }

func notifyHandler(e *event, live *store.Live) {
	// e.listen("notifyaudience", func(cmd string, n *Node) {
	// fmt.Println("~~~ cmd: ", cmd, " n: ", n.getElem("userlist").getElem("userlist"))
	// })

	// e.listen("notifykimenter", func(cmd string, n *Node) {
	// 	fmt.Println("~~~ cmd: ", cmd, " n: ", n.String(), " ~ ", n.Raw)
	// })

	// e.listen("*", func(cmd string, n *Node) {
	// 	switch cmd {
	// 	case "notifyglobalgift",
	// 		"notifybroadcast", "notifytextmsg",
	// 		"notifyenter", "notifygift",
	// 		"notifymobileglobalgift", "notifysw1906chipacmp", "notifydaywaverank", "notifyglobaltopgift",
	// 		"notifyusercnt":
	// 		return
	// 	}
	// 	fmt.Println("****** cmd: ", cmd, " n: ", n.String())
	// })

	// 礼物通知
	e.listen("notifygift", func(cmd string, n *Node) {
		gift := n.getElem("gift")
		if err := n.UnmarshalExt(); err != nil {
			fmt.Println(err)
			return
		}
		name := utils.Unescape(n.DExt.U1.NN)
		if name == "" {
			return
		}
		if gift.getAttr("fn") == "神秘人" && name != "jx"{
			gid := gift.getAttr("gid")
			info,ok := GiftInfo[gid]
			if !ok{
				return
			}
			EventChan <- EventMessage{Type: "神秘人发礼物", Msg: fmt.Sprintf("神秘人「%v」发送%v个%v\n",name,gift.getAttr("cnt"),info.Name)}
		}
	})

	// 进入房间
	e.listen("notifyenter", func(cmd string, n *Node) {
		userlist := n.getElem("userlist")
		nickname := userlist.getAttr("nickname")
		if err := n.UnmarshalExt(); err != nil {
			fmt.Println(err)
			return
		}
		if nickname == "神秘人" || (userlist.getAttr("id") == "9999" && userlist.getAttr("rid") == "0") {
			EventChan <- EventMessage{Type: "神秘人", Msg: fmt.Sprintf("神秘人「%v」进入房间\n", utils.Unescape(n.DExt.U1.NN))}
		} else {
			EventChan <- EventMessage{Type: "普通人", Msg: fmt.Sprintf("「%v」进入房间\n", nickname)}
		}
	})

	// 观众席通知
	// 离开房间
	e.listen("notifyaudience", func(cmd string, n *Node) {
		ele := n.getElem("userlist")
		child := ele.ChildNode()
		if ele.getAttr("type") == "1" && child.getAttr("type") == "2" {
			if err := n.UnmarshalExt(); err != nil {
				fmt.Println(err)
				return
			}
			name := utils.Unescape(n.DExt.U1.NN)
			if name == "" {
				return
			}
			EventChan <- EventMessage{Type: "退出房间", Msg: fmt.Sprintf("「%v」退出房间\n", name)}
		}
	})

	// 文本消息通知
	// e.listen("notifytextmsg", func(cmd string, n *Node) {
	// 	senderUID, err := getUserID(n.getAttr("f"))
	// 	if err != nil {
	// 		fmt.Println("notifytextmsg.getUserID err ", err)
	// 		return
	// 	}

	// 	msg := &store.Message{
	// 		SenderUid:    utils.ToStr(senderUID),
	// 		SenderName:   n.getAttr("n"),
	// 		ReceiverUid:  n.getAttr("c"),
	// 		ReceiverName: "",
	// 		Content:      strings.TrimPrefix(n.getText(), "#|"),
	// 		RoomID:       live.RoomID,
	// 		LiveID:       live.ID,
	// 		CreatedAt:    time.Now(),
	// 	}
	// 	_ = msg
	// 	// fmt.Println("---- ", n.String())
	// })

	// 房间用户数通知
	// e.listen("notifyusercnt", func(cmd string, n *Node) {
	// 	ruc := &store.RoomUserCount{
	// 		RoomID:    live.RoomID,
	// 		LiveID:    live.ID,
	// 		Count:     int32(utils.StrTo(n.getAttr("cnt")).MustInt()),
	// 		CreatedAt: time.Now(),
	// 	}
	// 	_ = ruc
	// 	// fmt.Println(*ruc)
	// })
}

func getUserID(fs string) (int64, error) {
	c := 0
	n := 0
	values := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}
	list := []string{}
	opt := make([]int, len(fs))
	for i := 0; i < len(fs); i++ {
		v, err := strconv.Atoi(string(fs[i]))
		if err != nil {
			return 0, err
		}
		opt[i] = v
	}
	for i := 0; i < len(fs); c = 0 {
		for j := i; j < len(fs); j++ {
			n = opt[j] + 10*c
			opt[j] = int(n / 16)
			c = n % 16
		}
		for list = append(list, values[c]); i < len(fs) && opt[i] == 0; i++ {
		}
	}
	reverse(list)
	str := strings.Join(list, "")
	return strconv.ParseInt(str[len(list)-8:], 16, 32)
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
