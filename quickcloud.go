package quickcloud

import (
	"github.com/astaxie/beego/httplib"
)

const (
	URL = "https://staging-api.quickcloud.io/"
)

type QuickCloud struct {
}

func (this *QuickCloud) ExchangeToken(requestCode string) {
	req := httplib.Get("http://beego.me/")
}
