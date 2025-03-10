package rest

import (
	"testing"

	"github.com/saikey0379/go-json-rest/rest/test"
	"golang.org/x/net/context"
)

func TestJSONIndentMiddleware(t *testing.T) {

	api := NewAPI()

	// the middleware to test
	api.Use(&JSONIndentMiddleware{})

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
	recorded.BodyIs("{\n  \"Id\": \"123\"\n}")
}
