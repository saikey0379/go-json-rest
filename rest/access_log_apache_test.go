package rest

import (
	"bytes"
	"log"
	"regexp"
	"testing"

	"golang.org/x/net/context"

	"github.com/saikey0379/go-json-rest/rest/test"
)

func TestAccessLogApacheMiddleware(t *testing.T) {

	api := NewAPI()

	// the middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogApacheMiddleware{
		Logger:       log.New(buffer, "", 0),
		Format:       CommonLogFormat,
		textTemplate: nil,
	})
	api.Use(&TimerMiddleware{})
	api.Use(&RecorderMiddleware{})

	// a simple app
	api.SetApp(AppSimple(func(ctx context.Context, w ResponseWriter, r *Request) {
		w.WriteJSON(map[string]string{"Id": "123"})
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJSON()

	// log tests, eg: '127.0.0.1 - - 29/Nov/2014:22:28:34 +0000 "GET / HTTP/1.1" 200 12'
	apacheCommon := regexp.MustCompile(`127.0.0.1 - - \d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} [+\-]\d{4}\ "GET / HTTP/1.1" 200 12`)

	if !apacheCommon.Match(buffer.Bytes()) {
		t.Errorf("Got: %s", buffer.String())
	}
}

func TestAccessLogApacheMiddlewareMissingData(t *testing.T) {

	api := NewAPI()

	// the uncomplete middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogApacheMiddleware{
		Logger:       log.New(buffer, "", 0),
		Format:       CommonLogFormat,
		textTemplate: nil,
	})

	// a simple app
	api.SetApp(AppSimple(func(ctx context.Context, w ResponseWriter, r *Request) {
		w.WriteJSON(map[string]string{"Id": "123"})
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJSON()

	// not much to log when the Env data is missing, but this should still work
	apacheCommon := regexp.MustCompile(` - -  "GET / HTTP/1.1" 0 -`)

	if !apacheCommon.Match(buffer.Bytes()) {
		t.Errorf("Got: %s", buffer.String())
	}
}
