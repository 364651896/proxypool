package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/liugc/proxypool/app/cache"
	"github.com/liugc/proxypool/config"
	"github.com/liugc/proxypool/provider"
	"github.com/liugc/proxypool/proxy"
	"github.com/liugc/proxypool/tool"
	"gopkg.in/yaml.v2"
)

var NeedFetchNewConfigFile = false

func CrawlGo() {
	if NeedFetchNewConfigFile {
		FetchNewConfigFileThenInit()
	}
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)
	for _, g := range Getters {
		wg.Add(1)
		go g.Get2Chan(pc, wg)
	}
	proxies := cache.GetProxies("proxies")
	go func() {
		wg.Wait()
		close(pc)
	}()
	for node := range pc {
		if node != nil {
			proxies = append(proxies, node)
		}
	}
	// 节点去重
	proxies = proxies.Deduplication()
	log.Println("CrawlGo node count:", len(proxies))
	proxies = provider.Clash{Proxies: proxies}.CleanProxies()
	proxies.NameAddCounrty().Sort().NameAddIndex()
	cache.SetProxies("allproxies", proxies)

	// 可用性检测
	proxies = proxy.CleanBadProxies(proxies)
	log.Println("CrawlGo clash useable node count:", len(proxies))
	proxies.NameAddCounrty().Sort().NameAddIndex()
	cache.SetProxies("proxies", proxies)

	cache.SetString("clashproxies", provider.Clash{Proxies: proxies}.Provide())
	cache.SetString("surgeproxies", provider.Surge{Proxies: proxies}.Provide())
}

func FetchNewConfigFileThenInit() {
	fmt.Println("fetch new config file...")
	resp, err := tool.GetHttpClient().Get("https://raw.githubusercontent.com/liugc/proxypool/master/source.yaml")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	err = yaml.Unmarshal(body, &config.SourceConfig)
	if err != nil {
		return
	}
	InitGetters(config.SourceConfig.Sources)
}
