package live

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/noaway/juxing/internal/httplib"
	"github.com/noaway/juxing/internal/utils"
	"github.com/noaway/juxing/internal/websocket"
	"github.com/noaway/juxing/internal/worker"
	"github.com/noaway/juxing/store"
	"github.com/sirupsen/logrus"
)

// service const
const (
	KuwoliveService = "https://zhiboserver.kuwo.cn/proxy.p"
)

// cmd constwangya
const (
	SocketCMDChannelID = "6"
	SocketMsgFSystem   = "0"
)

// StartLive fn
func StartLive(ctx context.Context, rid string, js *httplib.Json) error {
	var err error

	entry := logrus.WithFields(map[string]interface{}{
		"room_id": rid,
		"type":    "live",
	})
	if js == nil {
		js, err = getEnterRoom(rid)
		if err != nil {
			entry.Errorf("StartLive.getEnterRoom err, %v", err)
			return err
		}
	}

	switch js.Get("status").MustInt() {
	case 1:
		return enterRoomSuc(ctx, js, rid, entry)
	case 7:
		// enterRoomFail
		err = ErrNoEntryRoome
	case 10:
		// print js.Get("statusdesc").MustString()
	case 29:
		//
		// var t = o.type;
		// if (1 == t)
		// 本房间为加密房，请先登录后再进入
		// 请输入房间密码
		//
		// 	secretRoom.entPwdRoom(e.getRid());
		// else if (2 == t) {
		// 进入扣费房间 只有花钱才能进入
		// 	var n = o.value; 价格(单位星币)
		// 	secretRoom.entPriceRoom(e.getRid(), n)
		// }
		err = ErrEncryptOrChargeRoom
	case 30:
		// 主播给你拉黑名单了
		err = ErrByBlockList
	default:
		// 请稍后再试
		// 进入房间失败
		err = ErrFailToEnterRoom
	}
	return err
}

func getEnterRoom(rid string) (*httplib.Json, error) {
	v := url.Values{}
	v.Add("src", "web")
	v.Add("cmd", "enterroom")
	v.Add("rid", rid)
	v.Add("secrectname", "")
	v.Add("logintype", "")
	v.Add("from", "")
	v.Add("anonyid", "")
	v.Add("macid", "")
	v.Add("auto", "1")
	v.Add("csrc", "4")
	u := fmt.Sprintf("%s?%v", KuwoliveService, v.Encode())
	return tryGetToSimpleJSON(u)
}

func enterRoomSuc(ctx context.Context, js *httplib.Json, rid string, entry *logrus.Entry) error {
	ip := js.Get("chatroom").Get("ip").MustString()
	port := js.Get("chatroom").Get("port").MustInt()
	port = 18233
	addr := fmt.Sprintf("wss://%v:%v/", ip, port)

	signalCloseChan := make(chan struct{})
	c := Conn{
		id:              rid,
		js:              js,
		event:           newEvent(),
		Guardian:        worker.NewGuardian(),
		signalCloseChan: signalCloseChan,
		respAttrPool: sync.Pool{
			New: func() interface{} {
				return NewNode()
			},
		},
	}

	go func() {
		<-ctx.Done()
		c.Close()
	}()

	room := js.Get("room")
	live := store.Live{
		Name:       utils.Unescape(room.Get("name").MustString()),
		RoomID:     utils.ToStr(room.Get("id").MustInt()),
		RoomType:   "normal",
		AnchorUID:  utils.ToStr(room.Get("ownerid").MustInt64()),
		AnchorName: utils.Unescape(room.Get("name").MustString()),
		StartedAt:  time.Unix(room.Get("starttm").MustInt64(), 0),
		CreatedAt:  time.Now(),
	}
	if live.AnchorUID != "" && live.RoomID != "" {
		live.ID = utils.GetMD5HashString(live.AnchorUID + live.RoomID + utils.ToStr(live.StartedAt))
	}
	c.live = &live
	notifyHandler(c.event, &live)
	if err := websocket.NewDialer(addr, &c, entry); err != nil {
		entry.Errorf("enterRoomSuc.NewDialer err, %v", err)
		return err
	}
	<-signalCloseChan
	return c.Err()
}

func liveJSONCheck(params map[string]string) (string, error) {
	u := url.Values{}
	for k, v := range params {
		if v == "" {
			return "", fmt.Errorf("%v", "the parameter value cannot be an empty")
		}
		u.Add(k, v)
	}
	return u.Encode(), nil
}

// Conn struct
type Conn struct {
	websocket.Conn
	worker.Guardian
	*event

	live *store.Live

	id              string
	js              *httplib.Json
	respAttrPool    sync.Pool
	signalCloseChan chan struct{}

	key1           string
	loginMsg       string
	joinMsg        string
	userid, websid int64
	firstCome      int64
	chatid         string
	chatnum        uint32
	roomID         int64
	systm          int64
}

// Close fn
func (c *Conn) Close() {
	c.Conn.Close()
}

// OnClose func
func (c *Conn) OnClose() error {
	if c.signalCloseChan != nil {
		close(c.signalCloseChan)
	}
	c.Guardian.Close()
	return c.Err()
}

func (c *Conn) getNode() *Node {
	n := c.respAttrPool.Get().(*Node)
	for k := range n.Attr {
		delete(n.Attr, k)
	}
	return n
}

func (c *Conn) pushNode(n *Node) {
	c.respAttrPool.Put(n)
}

func (c *Conn) respParseAttr(doc *xml.Decoder) *Node {
	node := c.getNode()
	n := node
	for {
		token, err := doc.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			c.Logger.Error("respParseAttr.xml.token err ", err)
			break
		}
		switch elem := token.(type) {
		case xml.StartElement:
			if n.Name != "" && n.Name != elem.Name.Local {
				n.Child = NewNode()
				n = n.Child
			}
			n.Name = elem.Name.Local
			for _, a := range elem.Attr {
				n.Attr[a.Name.Local] = utils.Unescape(a.Value)
			}
		case xml.CharData:
			n.Text = string(elem)
		}
	}
	return node
}

// OnOpen func
func (c *Conn) OnOpen() error {
	if c.js == nil {
		return fmt.Errorf("%v", "js is nil")
	}
	js := c.js

	login, err := liveJSONCheck(map[string]string{
		"id":      "1",
		"sig":     strconv.Itoa(js.Get("chatroom").Get("channel").Get("login").Get("sig").MustInt()),
		"t":       strconv.Itoa(js.Get("chatroom").Get("tm").MustInt()),
		"channel": "0",
		"type":    "0",
		"content": "login:" + js.Get("chatid").MustString() + ":" + js.Get("chatroom").Get("channel").Get("login").Get("chatname").MustString(),
	})
	if err != nil {
		return fmt.Errorf("login OnOpen.liveJSONCheck err, %v", err)
	}

	join, err := liveJSONCheck(map[string]string{
		"id":      "2",
		"sig":     strconv.Itoa(js.Get("chatroom").Get("channel").Get("join").Get("sig").MustInt()),
		"t":       strconv.Itoa(js.Get("chatroom").Get("tm").MustInt()),
		"channel": strconv.Itoa(js.Get("chatroom").Get("channel").Get("join").Get("id").MustInt()),
		"type":    "0",
		"content": "join:" + js.Get("chatid").MustString(),
	})
	if err != nil {
		return fmt.Errorf("join OnOpen.liveJSONCheck err, %v", err)
	}

	c.loginMsg = login
	c.joinMsg = join
	me := js.Get("myenterinfo").Get("me")
	c.userid = me.Get("uid").MustInt64()
	c.websid = me.Get("sid").MustInt64()
	c.firstCome = time.Now().Unix() * 1000
	c.chatid = js.Get("chatid").MustString()
	c.roomID = js.Get("room").Get("id").MustInt64()
	c.systm = js.Get("systm").MustInt64()

	c.SendMsg(c.loginMsg)

	EventChan <- EventMessage{Type: "退出房间", Msg: fmt.Sprintf("欢迎来到【%v】的直播间\n", c.live.Name)}
	return nil
}

// OnMessage func
// this function does not run in the same goroutines as the context
func (c *Conn) OnMessage(messageType int, p []byte) {
	for _, str := range strings.Split(string(p), "\r\n") {
		if str != "" {
			data, _ := utils.GbkToUtf8([]byte(str))
			doc := xml.NewDecoder(bytes.NewReader(data))
			node := c.respParseAttr(doc)
			node.Raw = string(data)
			switch node.getAttr("id") {
			case "1":
				status := node.getElem("result").getAttr("status")
				if status == "ok" {
					c.key1 = node.getElem("key").getAttr("key1")
					// 初始化心跳
					c.loginChatServerSuccess()
				} else {
					EventChan <- EventMessage{Type: "error", Msg: fmt.Sprintf("%v", "进入直播间失败")}
				}
			case "2":
			case "3":
			default:
				c.socketData(node)
			}
			c.pushNode(node)
		}
	}
}

// SendMsg fn
func (c *Conn) SendMsg(msg string) {
	c.Conn.Send(1, []byte(msg+"\r\n"))
}

func (c *Conn) socketData(node *Node) {
	cval := node.getElem("resp").getAttr("c")
	tval := node.getElem("resp").getAttr("t")
	resp := node.getElem("resp")
	if cval == SocketCMDChannelID || tval == SocketMsgFSystem {
		text := node.getElem("resp").Text
		if text != "" && strings.Index(text, "cmd=") != -1 {
			n, err := c.txtToNode(text)
			if err != nil {
				c.Logger.Error("socketData.txtToNode err, ", err)
				return
			}
			n.Raw = node.Raw
			n.Ext = node.getAttr("ext")
			cmd := n.getAttr("cmd")
			if cmd == "" {
				c.Logger.Error("cmd is empty")
				return
			}
			c.notify(cmd, n)
		}
	} else if resp.Text != "touch:"+c.chatid {
		if utils.StrTo(tval).MustInt() == 2 {
			// 私有消息不处理
		} else {
			c.notify("notifytextmsg", resp)
		}
	}
}

func (c *Conn) loginChatServerSuccess() {
	c.Do(time.Second*300, c.sendHeartBeat)
	c.joinChennel()
}

func (c *Conn) sendHeartBeat() error {
	v := url.Values{}
	v.Add("src", "web")
	v.Add("cmd", "heartbeat")
	v.Add("uid", fmt.Sprintf("%v", c.userid))
	v.Add("sid", fmt.Sprintf("%v", c.websid))
	v.Add("chatid", c.chatid)
	v.Add("cookie", fmt.Sprintf("%v", c.firstCome))
	v.Add("rid", fmt.Sprintf("%v", c.roomID))
	v.Add("chatnum", fmt.Sprintf("%v", c.chatnum))
	v.Add("strnum", "0")
	v.Add("r", fmt.Sprintf("%v", c.roomID))
	v.Add("tm", fmt.Sprintf("%v", time.Now().Unix()-c.systm))
	v.Add("src2", "")
	v.Add("key1", c.key1)
	u := fmt.Sprintf("%v?%v", KuwoliveService, v.Encode())
	r := httplib.Get(u)
	r.Header("Referer", "http://jx.kuwo.cn/"+utils.ToStr(c.roomID))
	r.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36")
	resp, err := r.SetTimeout(time.Second*2, time.Second*10).Response()
	if err != nil {
		c.Logger.Error("sendHeartBeat.http err ", err)
		// 禁止重试
		return nil
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.Logger.Error("http heard beat err ", resp)
	}
	return nil
}

// joinChennel this function calls other resources
// that require locking
func (c *Conn) joinChennel() {
	time.Sleep(time.Second * 2)
	c.SendMsg(c.joinMsg)
}

// parse const
const (
	GiftKeyVal = "giftkeyval"
)

func (c *Conn) txtToNode(msg string) (*Node, error) {
	var (
		strs       = strings.Split(msg, "\n")
		brace      = "off"
		key, value = []string{}, []string{}
		status     = ""
	)
	node := NewNode()
	node.Name = "kuwo"
	n := node

	for _, str := range strs {
		if status == GiftKeyVal {
			value = strings.Split(str, "|")
			if len(key) != len(value) {
				return nil, fmt.Errorf("%v", "parse error k,v not equal")
			}
			for i := range key {
				n.Attr[key[i]] = utils.Unescape(value[i])
			}
			status = ""
			key, value = []string{}, []string{}
			continue
		}

		i := strings.Index(str, "=")
		if i != -1 {
			n.Attr[str[:i]] = utils.Unescape(str[i+1:])
		} else if i := strings.Index(str, "{"); i != -1 {
			if brace == "off" {
				n.Child = NewNode()
				n = n.ChildNode()
				n.Name = str[:i]
				brace = "on"
				continue
			}

			if l, g := strings.Index(str, "<"), strings.Index(str, ">"); l != -1 && g != -1 {
				n.Child = NewNode()
				n = n.ChildNode()
				n.Name = str[:l]
				key = strings.Split(str[l+1:g], "|")
				status = GiftKeyVal
				continue
			}

		} else if i := strings.Index(str, "}"); i != -1 {
			if brace == "on" {
				brace = "off"
			}
		}
	}
	return node, nil
}
