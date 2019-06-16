package live

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"sync"

	"github.com/noaway/juxing/internal/utils"
	"github.com/noaway/juxing/internal/httplib"
	"github.com/noaway/juxing/store"
	"github.com/sirupsen/logrus"
)

const (
	goSaveLive = "savelive"
	goLive     = "live"
)

var once sync.Once

var Cancel context.CancelFunc

var GiftInfo = make(map[string]*store.GiftInfo)

// StartMonitor fn
func StartMonitor(rid string) {
	// 处理全平台所有直播间room id
	// 拿到直播间需要的依赖
	once.Do(func(){GiftInfo = scrapyGiftList()})

	u := KuwoliveService + "?src=web&cmd=enterroom&rid=%s&auto=1"
	js, err := tryGetToSimpleJSON(fmt.Sprintf(u, rid), 4)
	if err != nil {
		logrus.Errorf("scan.get.livestatus.http err %v", err)
		return
	}
	room := js.Get("room")
	status := room.Get("livestatus").MustInt()
	switch status {
	case 1:
		// 没有直播
		EventChan <- EventMessage{Type: "error", Msg: fmt.Sprintf("%v: 没有直播\n", utils.Unescape(room.Get("name").MustString()))}
		return
	case 2:
		// 有直播
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

		ctx, cancel := context.WithCancel(context.Background())
		Cancel = cancel
		if err := StartLive(ctx, rid, js); err != nil {
			EventChan <- EventMessage{Type: "error", Msg: fmt.Sprintf("fnLive.start live err %v", err)}
			return
		}
	}

	sig := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, sig...)
	<-signalChan

	fmt.Println("直播退出")
}


func scrapyGiftList() map[string]*store.GiftInfo {
	u := "http://jx.kuwo.cn/KuwoLive_Mobile/GetGiftList"
	r := httplib.Get(u)
	r.Header("Connection", "keep-alive")
	r.Header("Cache-Control", "max-age=0")
	r.Header("Upgrade-Insecure-Requests", "1")
	r.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	r.Header("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	r.Header("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
	js, err := r.ToSimpleJson()
	if err != nil {
		fmt.Println("err ", err)
		return nil
	}
	data := js.Get("data")
	array := data.MustArray()
	gifts := make(map[string]*store.GiftInfo)
	errs := []error{}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc).Format("2006-01-02")
	for i := range array {
		item := data.GetIndex(i)
		gifts[utils.ToStr(item.Get("id").MustInt())] = &store.GiftInfo{
			ID:       utils.ToStr(item.Get("id").MustInt()),
			Name:     item.Get("name").MustString(),
			Price:    int32(item.Get("coin").MustInt()),
			Type:     utils.ToStr(item.Get("type").MustInt()),
			TypeName: item.Get("typedesc").MustString(),
			Date:     now,
			CreateAt: time.Now(),
		}
	}
	if len(errs) > 0 {
		for i := range errs {
			logrus.Error(errs[i])
		}
	}

	return gifts
}