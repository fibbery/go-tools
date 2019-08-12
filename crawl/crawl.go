package crawl

import (
	"crypto/tls"
	"fmt"
	. "github.com/PuerkitoBio/goquery"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	DefaultTimeOut = 3 * time.Second
)

var (
	semaphore   = make(chan struct{}, 50)
)



func GetHtmlDoc(url string, proxy *HttpProxy, timeout time.Duration) (*Document, int) {
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	request, _ := http.NewRequest("GET", url, nil)
	// set header and transport
	request.Header.Set("User-Agent", getAgent())
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Set("Connection", "keep-alive")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //fix x509 certificate problem
	}
	// set proxy
	if proxy != nil {
		tr.Proxy = http.ProxyURL(proxy.toUrl())
	}

	client := &http.Client{Transport: tr, Timeout: timeout,}
	response, err := client.Do(request)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return nil, -1
	}
	defer func() { _ = response.Body.Close() }()
	doc, err := NewDocumentFromReader(response.Body)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return nil, -1
	}
	return doc, response.StatusCode
}



func getAgent() string {
	agent := [...]string{
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
		"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
		"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
		"User-Agent,Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"User-Agent, Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
		"User-Agent,Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return agent[r.Intn(len(agent))]
}


