package main

import (
	"fmt"
	. "github.com/PuerkitoBio/goquery"
	. "github.com/fibbery/go-tools/crawl"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	doc, code := GetHtmlDoc("https://wap.f96.net/32/32448/all.html", nil, 3*time.Second)
	if code != http.StatusOK {
		fmt.Println("get url fail")
		return
	}
	var chapter int
	doc.Find("#chapterlist > p > a").Each(func(index int, element *Selection) {
		if index == 0 {
			return
		}
		href, exist := element.Attr("href")
		group := sync.WaitGroup{}
		if exist && strings.Contains(href, "http") {
			group.Add(1)
			chapter++
			chapter := chapter
			go func(link string) {
				defer group.Done()
				writeNovel(link, chapter)
			}(href)
		}
		group.Wait()
	})
}

func writeNovel(link string, chapter int) {
	var retry int
	for {
		if retry > 10 {
			fmt.Printf("retry time exceeded time 10 , can't write chapter %d\n", chapter)
		}
		proxy := GetRandomProxy()
		doc, code := GetHtmlDoc(link, proxy, 10*time.Second)
		if code != http.StatusOK {
			fmt.Printf("through proxy[%s] get chapter %d fail, will retry\n", proxy, chapter)
			DelProxy(proxy)
			retry++
			continue
		}
		content := doc.Find("#chaptercontent").Text()
		if content != "" {
			dir := "/Users/fibbery/Desktop/novel/" + strconv.Itoa(chapter) + ".txt"
			e := ioutil.WriteFile(dir, []byte(content), 0644)
			if e != nil {
				fmt.Printf("write chapter error %d, this error is %s\n", chapter, e.Error())
			} else {
				fmt.Printf("write chapter %d successfully\n", chapter)
			}
			break
		}
	}
}
