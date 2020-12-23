package main

import (
	"lolkor-monitoring/util"
	"sync"
)

func main() {
	c := make(chan []util.ArticleID, 256)
	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		defer wait.Done()
		monitoring := util.NewControlData(4444, c)
		monitoring.FindFilterArticle()
	}()

	go func() {
		defer wait.Done()
		findFilter := util.NewControlData(4445, c)
		findFilter.FilterArticle()
	}()

	wait.Wait()
}