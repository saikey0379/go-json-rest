package rest

import (
	"encoding/base64"
	"testing"

	"github.com/saikey0379/go-json-rest/rest/test"
	"golang.org/x/net/context"
)

func TestAuthBasic(t *testing.T) {

	// the middleware to test
	authMiddleware := &AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
		Authorizator: func(userId string, request *Request) bool {
			if request.Method == "GET" {
				return true
			}
			return false
		},
	}

	// api for testing failure
	apiFailure := NewAPI()
	apiFailure.Use(authMiddleware)
	apiFailure.SetApp(AppSimple(func(ctx context.Context, w ResponseWriter, r *Request) {
		t.Error("Should never be executed")
	}))
	handler := apiFailure.MakeHandler()

	// simple request fails
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(401)
	recorded.ContentTypeIsJSON()

	// auth with wrong cred and right method fails
	wrongCredReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:AdmIn"))
	wrongCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, handler, wrongCredReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJSON()

	// auth with right cred and wrong method fails
	rightCredReq := test.MakeSimpleRequest("POST", "http://localhost/", nil)
	encoded = base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	rightCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, handler, rightCredReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJSON()

	// api for testing success
	apiSuccess := NewAPI()
	apiSuccess.Use(authMiddleware)
	apiSuccess.SetApp(AppSimple(func(ctx context.Context, w ResponseWriter, r *Request) {
		env := EnvFromContext(ctx)
		if env["REMOTE_USER"] == nil {
			t.Error("REMOTE_USER is nil")
		}
		user := env["REMOTE_USER"].(string)
		if user != "admin" {
			t.Error("REMOTE_USER is expected to be 'admin'")
		}
		w.WriteJSON(map[string]string{"Id": "123"})
	}))

	// auth with right cred and right method succeeds
	rightCredReq = test.MakeSimpleRequest("GET", "http://localhost/", nil)
	encoded = base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	rightCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, apiSuccess.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJSON()
}
