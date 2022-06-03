package mesh

import "testing"

func TestParseHttp(t *testing.T) {
	h := "GET http://www.flysnow.org/ HTTP/1.1\nHost: www.flysnow.org\nProxy-Connection: keep-alive\nUpgrade-Insecure-Requests: 1\nUser-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36"
	ParseHttp([]byte(h))
}
