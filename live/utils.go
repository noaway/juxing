package live

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/noaway/juxing/internal/httplib"
	"github.com/noaway/juxing/internal/utils"
)

type DecodeExt struct {
	U1 struct {
		NN string `json:"nn"`
	} `json:"u1"`
}

// NewNode fn
func NewNode() *Node {
	return &Node{Attr: make(map[string]string)}
}

// Node struct
type Node struct {
	Name  string
	Attr  map[string]string
	Text  string
	Child *Node // TODO slice
	Raw   string
	Ext   string
	DExt  DecodeExt
}

func (n *Node) getElem(name string) *Node {
	node := n
	for {
		if node.Name == name {
			return node
		}
		if node.ChildNode() == nil {
			return &Node{}
		}
		node = node.Child
	}
}

func (n *Node) getAttr(attr string) string {
	if n == nil {
		return ""
	}
	return utils.Unescape(n.Attr[attr])
}

func (n *Node) getText() string {
	if n == nil || n.Text == "" {
		return ""
	}
	return utils.Unescape(n.Text)
}

func (n *Node) getName() string {
	if n == nil || n.Name == "" {
		return ""
	}
	return utils.Unescape(n.Name)
}

// ChildNode fn
func (n *Node) ChildNode() *Node {
	return n.Child
}

// String fn
func (n *Node) String() string {
	i := n
	content := ""
	for {
		content += fmt.Sprintf("Name: %v  Attr: %v Text: %v", i.getName(), i.Attr, i.getText())
		if i.Child == nil {
			break
		}
		i = i.Child
	}
	return content
}

func (n *Node) UnmarshalExt() error {
	if n.Ext == "" {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(n.Ext)
	if err != nil {
		return err
	}
	ret := DecodeExt{}
	if err := json.Unmarshal(data, &ret); err != nil {
		return err
	}
	n.DExt = ret
	return nil
}

func get(u string) *httplib.Request {
	r := httplib.Get(u)
	r.SetTimeout(time.Second*2, time.Second*10)
	r.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36")
	r.Header("Host", "zhiboserver.kuwo.cn")
	r.Header("Referer", "http://jx.kuwo.cn")
	return r
}

func tryGetToSimpleJSON(u string, trycounts ...int) (js *httplib.Json, err error) {
	count := 3
	if len(trycounts) > 0 {
		count = trycounts[0]
	}
	for i := 0; i < count; i++ {
		r := get(u)
		js, err = r.ToSimpleJson()
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				continue
			}
			netErr, ok := err.(net.Error)
			if !ok {
				return nil, err
			}
			if netErr.Timeout() {
				continue
			}

			opErr, ok := netErr.(*net.OpError)
			if !ok {
				return nil, err
			}

			switch opErr.Err.(type) {
			case *net.DNSError:
				time.Sleep(time.Second)
				continue
			case *os.SyscallError:
				time.Sleep(time.Second * 2)
				continue
			}

			return nil, err
		}
		return
	}
	return
}
