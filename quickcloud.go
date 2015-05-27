package quickcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/httplib"
)

const (
	HEADER_APP_ID       = "X-QuickCloud-App-Id"
	HEADER_APP_SECRET   = "X-QuickCloud-App-Secret"
	HEADER_TOKEN        = "X-QuickCloud-Session-Token"
	HEADER_CONTENT_TYPE = "Content-Type"
	APPLICATION_JSON    = "application/json"
	PRE_URL             = "https://staging-api.quickcloud.io"
	CORE                = "/core"
	SEARCH              = "/search"
	TOKEN               = "/oauth/token"
	GROUPS              = "/groups"
	APPS                = "/apps"
	USERS               = "/users"
	ROLES               = "/roles"
	LOGIN               = "/login"
	CONFIGURATION       = "/classes/Configuration"
	MEMBERS             = "/members"
	FILES               = "/files"
)

type Role struct {
	Name     string `json:"_name"`
	Group    string `json:"_group"`
	ObjectId string `json:"objectId"`
}

type Acl struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

type User struct {
	Name  string `json:"_name"`
	Email string `json:"_email"`
	Token string `json:"__sessionToken"`
	Response
}

type Response struct {
	ObjectId  string `json:"objectId"`
	CreatedAt string `json:"_createdAt"`
	Error     string `json:"error"`
}

type QuickCloud struct {
	Endpoint     string
	AppId        string
	AppSecret    string
	SessionToken string
	AdminToken   string
}

type File struct {
	ObjectId string `json:"objectId"`
	Url      string `json:"url"`
	Name     string `json:"_name"`
	Folder   string `json:"_folder"`
	GroupId  string `json:"groudId"`
}

func New(endpoint string, appId string, appSecret string) *QuickCloud {
	return &QuickCloud{endpoint, appId, appSecret, "", ""}
}

func Trace() {
	pc := make([]uintptr, 10) // at least 1 entry needed

	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])

	var p = strings.Split(file, "/")
	var n = strings.Split(f.Name(), "/")

	log.Infoln("[" + p[len(p)-1] + ":" + strconv.Itoa(line) + "] " + n[len(n)-1])
}

// Register an QuickCloud account
func (this *QuickCloud) Register(email string, password string, name string) string {

	Trace()
	var u = this.Endpoint + CORE + USERS
	req := httplib.Post(u)

	data := map[string]interface{}{
		"email":    email,
		"password": password,
		"name":     name,
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)
	req.Body(jsonStr)

	var resp Response
	err = req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	return resp.ObjectId
}

// Direct login and get Session Token
func (this *QuickCloud) Login(email string, password string, isAdmin bool) (string, string) {

	Trace()
	log.Infoln("Email: " + email)
	log.Infoln("Password: " + password)

	req := httplib.Post(this.Endpoint + CORE + LOGIN)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_APP_SECRET, this.AppSecret)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req.Body(jsonStr)

	var resp User
	err = req.ToJson(&resp)

	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	this.SessionToken = resp.Token

	if isAdmin {
		this.AdminToken = resp.Token
	}

	return resp.Token, resp.ObjectId
}

// Exchange for session token
func (this *QuickCloud) Token(requestCode string) string {

	Trace()

	req := httplib.Post(this.Endpoint + TOKEN)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_APP_SECRET, this.AppSecret)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data := map[string]interface{}{
		"code": requestCode,
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req.Body(jsonStr)

	var resp User
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
		panic(resp.Error)
	}
	this.SessionToken = resp.Token
	return resp.Token
}

// Create a new group / case
func (this *QuickCloud) CreateGroup(name string, desc string) string {

	Trace()
	log.Infoln("Creating group: " + name)

	req := httplib.Post(this.Endpoint + CORE + GROUPS)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data := map[string]interface{}{
		"_name":       name,
		"description": desc,
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req.Body(jsonStr)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	log.Warningln("GroupId: " + resp.ObjectId)
	return resp.ObjectId
}

// Create Role in given Group
func (this *QuickCloud) CreateRole(groupId string, role string) string {

	Trace()

	req := httplib.Post(this.Endpoint + CORE + ROLES)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data := map[string]interface{}{
		"group": groupId,
		"name":  role,
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req.Body(jsonStr)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}
	return resp.ObjectId
}

// Upload given Configuration
func (this *QuickCloud) CreateConfiguration(groupId string, config interface{}) string {

	Trace()

	req := httplib.Post(this.Endpoint + CORE + APPS + "/" + this.AppId + CONFIGURATION)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	req.Body(data)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}
	return resp.ObjectId
}

// Assign Role
// /groups/{objectId}/members/{userId}/roles
// Body
// roleName (required, string) ... Name of Role to assign (Admin, Supervisor)
func (this *QuickCloud) AssignRole(groupId string, userId string, role string) {

	Trace()

	if this.AdminToken == "" {
		log.Errorln("Admin Token is missing")
		panic("Admin Token is missing")
	}
	url := this.Endpoint + CORE + GROUPS + "/" + groupId + MEMBERS + "/" + userId + "/roles"

	req := httplib.Post(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.AdminToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data := map[string]interface{}{
		"roleName": role,
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req.Body(jsonStr)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}
}

func (this *QuickCloud) GetInvitationCode(groupId string, email string, sendMessage bool) string {

	Trace()
	log.Infoln("Get invitation code for " + email)

	if this.AdminToken == "" {
		log.Errorln("Admin Token is missing")
		panic("Admin Token is missing")
	}

	url := this.Endpoint + CORE + GROUPS + "/" + groupId + MEMBERS + "/invite"
	req := httplib.Post(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.AdminToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	if sendMessage {
		data := map[string]interface{}{
			"emailTo": email,
			"message": "New user want to join group " + groupId,
		}

		jsonStr, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		req.Body(jsonStr)
	} else {
		req.String()
	}

	type Response struct {
		Code  string `json:"code"`
		Error string `json:"error"`
	}

	var resp Response
	err := req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	return resp.Code
}

func (this *QuickCloud) JoinGroup(userToken string, code string) {

	Trace()

	url := this.Endpoint + CORE + GROUPS + "/join"
	req := httplib.Post(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, userToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data := map[string]interface{}{
		"code": code,
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req.Body(jsonStr)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}
}

func (this *QuickCloud) UploadFile(groupId string, file string, name string) File {

	url := this.Endpoint + CORE + FILES
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes("_file"), escapeQuotes(file)))
	h.Set("Content-Type", mime.TypeByExtension(filepath.Ext(file)))

	//fw, err := writer.CreateFormFile("_file", file)
	fw, err := writer.CreatePart(h)
	if err != nil {
		panic(err)
	}

	if _, err = io.Copy(fw, f); err != nil {
		panic(err)
	}

	// Add the other fields
	if fw, err = writer.CreateFormField("_name"); err != nil {
		panic(err)
	}
	if _, err = fw.Write([]byte(name)); err != nil {
		panic(err)
	}

	if fw, err = writer.CreateFormField("_folder"); err != nil {
		panic(err)
	}
	if _, err = fw.Write([]byte(groupId)); err != nil {
		panic(err)
	}

	if fw, err = writer.CreateFormField("groupId"); err != nil {
		panic(err)
	}
	if _, err = fw.Write([]byte(groupId)); err != nil {
		panic(err)
	}

	writer.Close()

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", url, body)

	req.Header.Add("X-QuickCloud-Session-Token", this.SessionToken)
	req.Header.Add("X-QuickCloud-App-Id", this.AppId)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// Fetch Request
	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(res.Body)

	var resp Response
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	fileUrl := this.Endpoint + CORE + FILES + "/" + resp.ObjectId
	log.Infoln("Uploaded file: " + fileUrl)

	this.SetPublicAcl(fileUrl)
	return File{
		ObjectId: resp.ObjectId,
		Url:      fileUrl,
		Name:     name,
		Folder:   groupId,
		GroupId:  groupId,
	}
}

// Register class to Search Index
// https://staging-api.quickcloud.io/search/apps/551d559764617400a4380000/Information/register
func (this *QuickCloud) RegisterSearchIndex(groupId string, className string) {

	req := httplib.Post(this.Endpoint + SEARCH + APPS + "/" + groupId + "/" + className + "/register")
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	_, err := req.String()
	if err != nil {
		panic(err)
	}

	log.Infoln("Registered search index for class:" + className)
}

func (this *QuickCloud) CreatePublicObject(groupId string, className string, object interface{}) string {
	Trace()

	url := this.Endpoint + CORE + "/classes/" + className
	req := httplib.Post(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	req.Body(data)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	this.SetRoleAcl(url+"/"+resp.ObjectId, groupId)

	log.Infoln(" -> Create " + className + "/" + resp.ObjectId)
	return resp.ObjectId
}

func (this *QuickCloud) CreateObject(groupId string, className string, object interface{}) string {

	Trace()

	url := this.Endpoint + CORE + "/classes/" + className
	req := httplib.Post(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	req.Body(data)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	this.SetPublicAcl(url + "/" + resp.ObjectId)

	return resp.ObjectId
}

func (this *QuickCloud) CreateAppObject(groupId string, className string, object interface{}) string {

	url := this.Endpoint + CORE + APPS + "/" + this.AppId + "/classes/" + className
	req := httplib.Post(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	req.Body(data)

	var resp Response
	err = req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	if resp.Error != "" {
		log.Errorln(resp.Error)
	}

	log.Infoln(" -> Create " + className + "/" + resp.ObjectId)

	this.SetPublicAcl(url + "/" + resp.ObjectId)
	return resp.ObjectId
}

func (this *QuickCloud) GetRole(groupId string) []Role {

	req := httplib.Get(this.Endpoint + CORE + ROLES)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	req.Param("where", "{\"_group\":\""+groupId+"\"}")

	type Response struct {
		Results []Role `json:"results"`
	}

	var resp Response
	err := req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%+v\n", resp)
	return resp.Results
}

func (this *QuickCloud) SetRoleAcl(url string, groupId string) {

	// Get Object
	req := httplib.Get(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	type Response struct {
		Name map[string]Acl `json:"_acl"`
	}

	var resp Response
	err := req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	roles := this.GetRole(groupId)

	adminAcl := Acl{true, true}
	normalAcl := Acl{true, false}
	publicAcl := Acl{true, true}

	resp.Name["*"] = publicAcl

	for _, value := range roles {
		if value.Name == "Admin" {
			resp.Name["role:"+value.ObjectId] = adminAcl
		} else {
			resp.Name["role:"+value.ObjectId] = normalAcl
		}
	}

	// fmt.Printf("%+v\n", resp)

	putReq := httplib.Put(url)
	putReq.Header(HEADER_APP_ID, this.AppId)
	putReq.Header(HEADER_TOKEN, this.SessionToken)
	putReq.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data, err1 := json.Marshal(resp)
	if err1 != nil {
		panic(err1)
	}

	putReq.Body(data)
	_, err = putReq.String()
	if err != nil {
		panic(err)
	}
}

func (this *QuickCloud) SetPublicAcl(url string) {

	// Trace()

	// Get Object
	req := httplib.Get(url)
	req.Header(HEADER_APP_ID, this.AppId)
	req.Header(HEADER_TOKEN, this.SessionToken)
	req.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	type Response struct {
		Name map[string]Acl `json:"_acl"`
	}

	var resp Response
	err := req.ToJson(&resp)
	if err != nil {
		panic(err)
	}

	publicAcl := Acl{true, true}
	resp.Name["*"] = publicAcl

	// fmt.Printf("%+v\n", resp)

	putReq := httplib.Put(url)
	putReq.Header(HEADER_APP_ID, this.AppId)
	putReq.Header(HEADER_TOKEN, this.SessionToken)
	putReq.Header(HEADER_CONTENT_TYPE, APPLICATION_JSON)

	data, err1 := json.Marshal(resp)
	if err1 != nil {
		panic(err1)
	}

	putReq.Body(data)
	_, err = putReq.String()
	if err != nil {
		panic(err)
	}
}

func escapeQuotes(s string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(s)
}
