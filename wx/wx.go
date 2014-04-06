package wx

import (
	"appengine"
	"appengine/delay"
	"appengine/urlfetch"
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	sj "github.com/bitly/go-simplejson"
	"github.com/drone/routes"
	"io"
	"io/ioutil"
	"myutils/datamanager"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Message1 struct {
	MsgType string
	Content string
}

var accessToken string

func init() {
	mux := routes.New()
	mux.Post("/wx", wxCheck)
	mux.Get("/wx", wxCheck)
	mux.Get("/test", test)
	mux.Get("/image", image)
	http.Handle("/", mux)
	//accessToken, _ = getAccessToken()
	//fmt.Printf("token is %s\n", accessToken)
	//http.ListenAndServe(":80", nil)

}

func image(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	r.ParseForm()
	key := r.FormValue("key")
	io.Copy(w, datamanager.ReadFile(c, appengine.BlobKey(key)))
}

func test(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	r.ParseForm()
	page, _ := strconv.ParseInt(r.FormValue("page"), 0, 32)
	fmt.Fprint(w, datamanager.ShowImage(c, page))
}
func wxCheck(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("request incoming!")
	if ok, echostr := check(w, r); ok {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintln(w, echostr)
			return
		}
		var message datamanager.Message
		xml.Unmarshal(buf, &message)
		c.Infof("msgtype is %s,echostr is %s", message.MsgType, echostr)
		switch message.MsgType {
		case "text":
			replay(message.FromUserName, "呵呵", w)
		case "image":
			replay(message.FromUserName, "你的图片已上传。", w)
			var save = delay.Func("save", storeImage)
			save.Call(c, message)
		default:
			fmt.Fprintln(w, echostr)

		}
		c.Infof(string(buf))
	} else {
		c.Infof("error")
	}

}
func check(w http.ResponseWriter, r *http.Request) (bool, string) {
	token := "lp3385"
	params := r.URL.Query()
	timestamp := params.Get("timestamp")
	//fmt.Printf("timestamp = %s\n", timestamp)
	nonce := params.Get("nonce")
	//fmt.Printf("nonce = %s\n", nonce)
	echostr := params.Get("echostr")
	//fmt.Printf("echostr = %s\n", echostr)
	strs := []string{token, nonce, timestamp}
	sort.Strings(strs)
	result := strings.Join(strs, "")
	h := sha1.New()
	io.WriteString(h, result)
	result = fmt.Sprintf("%x", h.Sum(nil))
	//fmt.Printf("result = %s\n", result)
	signature := params.Get("signature")
	//fmt.Printf("signature = %s\n", signature)
	if result == signature {
		return true, echostr
	} else {
		return false, ""
	}
}
func replay(userName string, text string, w http.ResponseWriter) {
	var message datamanager.Message
	message.FromUserName = "gh_145b3f29c8e1"
	message.ToUserName = userName
	message.Content = text
	message.MsgType = "text"
	message.CreateTime = time.Now().Unix()
	buf, _ := xml.MarshalIndent(message, " ", "  ")
	fmt.Fprint(w, string(buf))
}

func getAccessToken(c appengine.Context) string {
	if "" == accessToken {
		client := urlfetch.Client(c)
		resp, _ := client.Get("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=wx06468735c6ca4e0e&secret=561a7b6cd3bf6cca7611f7cb3232ac2e")
		body, _ := ioutil.ReadAll(resp.Body)
		js, _ := sj.NewJson(body)
		return js.Get("access_token").MustString()
	} else {
		return accessToken
	}

}

func postMessage(c appengine.Context, msg string) {
	m := `{
    "touser":"o230Qt0p0YrMhrLzSLMLYu9zHTSE",
    "msgtype":"text",
    "text":
    {
         "content":"Hello World"
    }`
	client := urlfetch.Client(c)
	client.Post("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token="+getAccessToken(c), "application/json", strings.NewReader(m))
}

func storeImage(c appengine.Context, message datamanager.Message) {
	client := urlfetch.Client(c)
	resp, _ := client.Get("http://file.api.weixin.qq.com/cgi-bin/media/get?access_token=" + getAccessToken(c) + "&media_id=" + message.MediaId)
	buf, _ := ioutil.ReadAll(resp.Body)
	key, _ := datamanager.SaveFile(c, buf, "image/jpeg")
	message.Key = key
	datamanager.PutMessage(c, message)

	//ioutil.WriteFile("a.amr", buf, 0777)
}
