package getter

import (
	"fmt"
	"sync"

	"github.com/gocolly/colly"
	"github.com/liugc/proxypool/proxy"
	"github.com/liugc/proxypool/tool"
)

func init() {
	Register("tgchannel", NewTGChannelGetter)
}

type TGChannelGetter struct {
	c         *colly.Collector
	NumNeeded int
	results   []string
	Url       string
}

func NewTGChannelGetter(options tool.Options) (getter Getter, err error) {
	num, found := options["num"]
	t := 200
	switch num.(type) {
	case int:
		t = num.(int)
	case float64:
		t = int(num.(float64))
	}

	if !found || t <= 0 {
		t = 200
	}
	urlInterface, found := options["channel"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &TGChannelGetter{
			c:         colly.NewCollector(),
			NumNeeded: t,
			Url:       "https://t.me/s/" + url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}

func (g *TGChannelGetter) Get() proxy.ProxyList {
	g.results = make([]string, 0)
	// 找到所有的文字消息
	g.c.OnHTML("div.tgme_widget_message_text", func(e *colly.HTMLElement) {
		g.results = append(g.results, GrepLinksFromString(e.Text)...)
	})

	// 找到之前消息页面的链接，加入访问队列
	g.c.OnHTML("link[rel=prev]", func(e *colly.HTMLElement) {
		if len(g.results) < g.NumNeeded {
			_ = e.Request.Visit(e.Attr("href"))
		}
	})

	g.results = make([]string, 0)
	err := g.c.Visit(g.Url)
	if err != nil {
		_ = fmt.Errorf("%s", err.Error())
	}

	return StringArray2ProxyArray(g.results)
}

func (g *TGChannelGetter) Get2Chan(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := g.Get()
	for _, node := range nodes {
		pc <- node
	}
}
