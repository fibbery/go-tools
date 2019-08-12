package crawl

import (
	"fmt"
	"testing"
)

func TestGetRandomProxy(t *testing.T) {
	var proxy *HttpProxy
	for {
		proxy = GetRandomProxy()
		if proxy.Available() {
			break
		}
	}
	fmt.Println(proxy)
}
