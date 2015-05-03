package quickcloud

import (
	"net/url"
	"strings"

	"github.com/astaxie/beego/httplib"
)

const (
	HEADER_APP_ID     = "X-QuickCloud-App-Id"
	HEADER_APP_SECRET = "X-QuickCloud-App-Secret"
	HEADER_TOKEN      = "X-QuickCloud-Session-Token"
	PRE_URL           = "https://staging-api.quickcloud.io"
	CORE              = "/core"
	TOKEN             = "/oauth/token"
	GROUPS            = "/groups"
	APPS              = "/apps"
	USERS             = "/users"
)

type QuickCloud struct {
	Endpoint     string
	AppId        string
	AppSecret    string
	SessionToken string
}

func New(endpoint string) *QuickCloud {
	return &QuickCloud{endpoint}
}

func New(endpoint string, appId string, appSecret string) *QuickCloud {
	return &QuickCloud{endpoint, appId, appSecret}
}

func (this *QuickCloud) Register(email string, password string, name string) {
	req := httplib.Post(this.Endpoint + CORE + USERS)

	var data = url.Values{}
	data.Add("email", email)
	data.Add("password", password)
	data.Add("name", name)
	req.Body(strings.NewReader(data.Encode()))

	resp, err := req.String()
	if err == nil {
		panic(err)
	}
}

func (this *QuickCloud) Token(requestCode string) string {
	req := httplib.Post(this.Endpoint + TOKEN)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_APP_SECRET, this.AppSecret)
	req.Header("Content-Type", "application/json")
	req.Body("{code: " + requestCode + "}")

	token, err := req.String()
	if err == nil {
		panic(err)
	}

	this.SessionToken = token
	return token
}

func (this *QuickCloud) Me() {

}
