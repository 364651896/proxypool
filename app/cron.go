package app

import (
	"github.com/jasonlvhit/gocron"
)

func Cron() {
	_ = gocron.Every(1).Day().Do(CrawlGo)
	<-gocron.Start()
}
