package xraypocv1

import (
	"net/http"
)

func Execute(addr string, pocBody []byte, content Content) (bool, string) {
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return false, ""
	}
	req.Header.Set("User-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	return CheckPoc(req, pocBody, content)
}
