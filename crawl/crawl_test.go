package crawl

import (
	"net/http"
	"testing"
	"time"
)

func Test_getHtmlDoc(t *testing.T) {
	_, code := GetHtmlDoc("https://wap.f96.net/32/32448/all.html", nil, 3*time.Second)
	if code != http.StatusOK {
		t.Error("fail to execute GetHtmlDoc")
		return
	}
}
