package quickcloud

import (
	"encoding/json"
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
	ROLES             = "/roles"
	LOGIN             = "/login"
	CONFIGURATION     = "/classes/Configuration"
	MEMBERS           = "/members"
)

type QuickCloud struct {
	Endpoint     string
	AppId        string
	AppSecret    string
	SessionToken string
}

func New(endpoint string, appId string, appSecret string) *QuickCloud {
	return &QuickCloud{endpoint, appId, appSecret, ""}
}

func (this *QuickCloud) Register(email string, password string, name string) string {
	req := httplib.Post(this.Endpoint + CORE + USERS)

	var data = url.Values{}
	data.Add("email", email)
	data.Add("password", password)
	data.Add("name", name)
	req.Body(strings.NewReader(data.Encode()))

	type Response struct {
		ObjectId string `json:"objectId"`
	}
	var resp Response
	err := req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	return resp.ObjectId
}

// Direct login and get Session Token
func (this *QuickCloud) Login(email string, password string) string {

	req := httplib.Post(this.Endpoint + CORE + LOGIN)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_APP_SECRET, this.AppSecret)

	var data = url.Values{}
	data.Add("email", email)
	data.Add("password", password)
	req.Body(strings.NewReader(data.Encode()))

	type Response struct {
		Token string `json:"__sessionToken"`
	}

	var resp Response
	err := req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	this.SessionToken = resp.Token
	return resp.Token
}

// Exchange for session token
func (this *QuickCloud) Token(requestCode string) string {

	req := httplib.Post(this.Endpoint + TOKEN)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_APP_SECRET, this.AppSecret)
	req.Header("Content-Type", "application/json")
	req.Body("{code: " + requestCode + "}")

	type Response struct {
		Token string `json:"__sessionToken"`
	}

	var resp Response
	err := req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	this.SessionToken = resp.Token
	return resp.Token
}

// Create a new group / case
func (this *QuickCloud) CreateGroup(name string, desc string) string {
	req := httplib.Post(this.Endpoint + CORE + GROUPS)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_APP_SECRET, this.AppSecret)

	var data = url.Values{}
	data.Add("_name", name)
	data.Add("description", desc)
	req.Body(strings.NewReader(data.Encode()))

	type Response struct {
		GroupId string `json:"objectId"`
	}
	var resp Response
	err := req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	return resp.GroupId
}

// Create Role in given Group
func (this *QuickCloud) CreateRole(groupId string, role string) string {
	req := httplib.Post(this.Endpoint + CORE + ROLES)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)

	var data = url.Values{}
	data.Add("group", groupId)
	data.Add("name", role)
	req.Body(strings.NewReader(data.Encode()))

	type Response struct {
		RoleId string `json:"objectId"`
	}

	var resp Response
	err := req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	return resp.RoleId
}

// Upload given Configuration
func (this *QuickCloud) CreateConfiguration(groupId string, config interface{}) string {

	req := httplib.Post(this.Endpoint + CORE + APPS + "/" + this.AppId + CONFIGURATION)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)

	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	req.Body(data)
	type Response struct {
		ObjectId string `json:"objectId"`
	}

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	return resp.ObjectId
}

// Assign Role
// /groups/{objectId}/members/{userId}/roles
// Body
// roleName (required, string) ... Name of Role to assign (Admin, Supervisor)
func (this *QuickCloud) AssignRole(groupId string, userId string, role string) {
	req := httplib.Post(this.Endpoint + CORE + GROUPS + "/" + groupId + MEMBERS + "/" + userId + "/roles")
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)

	var data = url.Values{}
	data.Add("roleName", role)
	req.Body(strings.NewReader(data.Encode()))

	_, err := req.String()
	if err != nil {
		panic(err)
	}
}

func (this *QuickCloud) GetInvitationCode(groupId string) string {
	req := httplib.Post(this.Endpoint + CORE + GROUPS + "/" + groupId + MEMBERS + "/invite")
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)

	var data = url.Values{}
	data.Add("emailTo", "hi@dwarvesf.com")
	data.Add("message", "New user want to join group "+groupId)
	req.Body(strings.NewReader(data.Encode()))

	type Response struct {
		Code string `json:"code"`
	}

	var resp Response
	err := req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	return resp.Code
}

func (this *QuickCloud) JoinGroup(userToken string, code string) {
	req := httplib.Post(this.Endpoint + CORE + GROUPS + "/join")
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, userToken)

	var data = url.Values{}
	data.Add("code", code)
	req.Body(strings.NewReader(data.Encode()))

	_, err := req.String()
	if err != nil {
		panic(err)
	}
}

func (this *QuickCloud) UploadFile(icon string, public bool) {

}
