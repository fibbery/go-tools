package crawl

import (
	"encoding/json"
	"fmt"
	. "github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis"
	"net/http"
	. "net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// redis keys
	KeyHasUpdateToday = "proxy:has:update:today"
	KeyProxyIpPool    = "proxy:ip:pool"
)

var (
	proxyGetUrl = "https://www.kuaidaili.com/free/inha"
	client      = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		OnConnect: func(conn *redis.Conn) error {
			_, err := conn.Ping().Result()
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("connect to redis successfully !!!!")
			return nil
		},
	})
)

type HttpProxy struct {
	Protocol string
	Ip       string
	Port     int
}

func (p *HttpProxy) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *HttpProxy) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

func (p *HttpProxy) String() string {
	return p.Protocol + "://" + p.Ip + ":" + strconv.Itoa(p.Port)
}

func (p *HttpProxy) toUrl() *URL {
	url, e := Parse(p.String())
	if e != nil {
		return nil
	}
	return url
}

func (p *HttpProxy) Available() bool {
	if p == nil {
		return false
	}
	_, code := GetHtmlDoc("https://www.baidu.com", p, DefaultTimeOut)
	return code == http.StatusOK
}

func InitProxy() {
	var count int
	hasUpdate, _ := client.Get(KeyHasUpdateToday).Int()
	if hasUpdate == 1 {
		fmt.Println("不需要更新代理库")
		return
	}
	var page = 100
	for {
		if count >= 100 {
			//更新完毕
			fmt.Println("update proxy pool successfully!!!!")
			client.Set(KeyHasUpdateToday, 1, 24*time.Hour)
			break
		}
		page++
		//设置请求间隔时间防止被封
		time.Sleep(100 * time.Millisecond)
		contentUrl := strings.Join([]string{proxyGetUrl, strconv.Itoa(page)}, string(os.PathSeparator))
		doc, code := GetHtmlDoc(contentUrl, nil, DefaultTimeOut)
		if code != http.StatusOK {
			continue
		}
		//获取具体代理地址
		var group sync.WaitGroup
		doc.Find("#list > table tbody tr").Each(func(index int, element *Selection) {
			if index == 0 {
				return
			}
			port, _ := strconv.Atoi(element.Find("td").Eq(1).Text())
			proxy := &HttpProxy{
				Ip:       element.Find("td").Eq(0).Text(),
				Port:     port,
				Protocol: strings.ToLower(element.Find("td").Eq(3).Text()),
			}
			group.Add(1)
			go func() {
				if proxy.Available() {
					client.SAdd(KeyProxyIpPool, proxy)
					count++
					fmt.Printf("load proxy url in redis : %s\n", proxy)
				}
				defer group.Done()
			}()
		})
		group.Wait()
	}
}

// 获取随机代
func GetRandomProxy() *HttpProxy {
	data, err := client.SRandMember(KeyProxyIpPool).Result()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}
	var proxy HttpProxy
	_ = proxy.UnmarshalBinary([]byte(data))
	return &proxy
}

func DelProxy(proxy *HttpProxy) {
	client.SRem(KeyProxyIpPool, proxy)
}
