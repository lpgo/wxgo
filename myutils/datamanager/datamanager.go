package datamanager

import (
	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"encoding/xml"
	"fmt"
)

type Message struct {
	XMLName      xml.Name `datastore:"-" xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	Content      string            `xml:"Content,omitempty"`
	PicUrl       string            `xml:"PicUrl,omitempty"`
	MediaId      string            `xml:"MediaId,omitempty"`
	ThumbMediaId string            `xml:"ThumbMediaId,omitempty"`
	Format       string            `xml:"Format,omitempty"`
	Location_X   string            `xml:"Location_X,omitempty"`
	Location_Y   string            `xml:"Location_Y,omitempty"`
	Scale        string            `xml:"Scale,omitempty"`
	Label        string            `xml:"Label,omitempty"`
	Title        string            `xml:"Title,omitempty"`
	Description  string            `xml:"Description,omitempty"`
	Url          string            `xml:"Url,omitempty"`
	MsgId        int64             `xml:"MsgId,omitempty"`
	Key          appengine.BlobKey `xml:"-"`
}

/*func handleText(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	reader := blobstore.NewReader(c, key)
	buf, _ := ioutil.ReadAll(reader)
	fmt.Fprintln(w, string(buf))

}
*/
func SaveFile(c appengine.Context, buf []byte, mimeType string) (appengine.BlobKey, error) {
	writer, err := blobstore.Create(c, "image/jpeg")
	if err != nil {
		c.Errorf("SaveFile : %s", err.Error())
		return "", err
	}
	writer.Write(buf)
	writer.Close()
	return writer.Key()
}

func PutMessage(c appengine.Context, message Message) {
	key := datastore.NewIncompleteKey(c, "Message", nil)
	_, err := datastore.Put(c, key, &message)
	if err != nil {
		c.Errorf("PutMessage: %s", err.Error())
	}
}

func ReadFile(c appengine.Context, key appengine.BlobKey) blobstore.Reader {
	return blobstore.NewReader(c, key)
}

func ShowImage(c appengine.Context, page int64) string {
	var msgs []Message
	q := datastore.NewQuery("Message").Order("-CreateTime").Limit(5).Offset(5 * int(page))
	q.GetAll(c, &msgs)
	var data string = "["
	var s string
	for index, msg := range msgs {
		if index == (len(msgs) - 1) {
			s = fmt.Sprintf(`{"src":"/image?key=%s","title":"title"}`, msg.Key)
		} else {
			s = fmt.Sprintf(`{"src":"/image?key=%s","title":"title"},`, msg.Key)
		}
		data += s
	}
	data += "]"
	return data
}
