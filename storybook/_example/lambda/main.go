package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/a-h/templ/storybook/example"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var s = example.Storybook()

func build() {
	if err := s.Build(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Embed the build output into the Lambda.
// The build output is only 4MB, so there's plenty of space.
//
//go:embed storybook-server/storybook-static
var storybookStatic embed.FS

func run() {
	// Replace the filesystem handler with the embedded data.
	rooted, _ := fs.Sub(storybookStatic, "storybook-server/storybook-static")
	s.StaticHandler = http.FileServer(http.FS(rooted))
	// Start a Lambda handler.
	lambda.Start(handler)
}

func handler(ctx context.Context, e events.APIGatewayV2HTTPRequest) (resp events.APIGatewayV2HTTPResponse, err error) {
	// Record the result.
	w := httptest.NewRecorder()
	u := e.RawPath
	if len(e.RawQueryString) > 0 {
		u += "?" + e.RawQueryString
	}
	r := httptest.NewRequest(e.RequestContext.HTTP.Method, u, nil)
	s.ServeHTTP(w, r)

	// Convert it to an API Gateway response.
	result := w.Result()
	resp.StatusCode = result.StatusCode
	bdy, err := io.ReadAll(w.Result().Body)
	if err != nil {
		return
	}
	resp.Body = string(bdy)
	if len(result.Header) > 0 {
		resp.Headers = make(map[string]string, len(result.Header))
		for k := range result.Header {
			v := result.Header.Get(k)
			resp.Headers[k] = v
		}
	}
	cookies := result.Cookies()
	if len(cookies) > 0 {
		resp.Cookies = make([]string, len(cookies))
		for i := 0; i < len(cookies); i++ {
			resp.Cookies[i] = cookies[i].String()
		}
	}
	return
}

func main() {
	if len(os.Args) < 2 {
		run()
	}
	switch os.Args[1] {
	case "build":
		build()
	case "run":
		run()
	default:
		fmt.Printf("unexpected command %q\n", os.Args[1])
		os.Exit(1)
	}
}
