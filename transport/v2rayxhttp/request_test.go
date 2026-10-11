package xhttp

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sagernet/sing-box/option"
)

// Request construction only: these tests never invoke OpenStream, a transport,
// a listener, or a network request.
func TestNewStreamRequest(t *testing.T) {
	for _, tc := range []struct {
		name      string
		host      string
		method    string
		body      string
		session   string
		noGRPC    bool
		wantError bool
	}{
		{name: "download", host: "example.invalid", method: "GET"},
		{name: "upload", host: "example.invalid", method: "POST", body: "synthetic payload", session: "synthetic-session"},
		{name: "upload-no-grpc", host: "example.invalid", method: "POST", body: "synthetic payload", noGRPC: true},
		{name: "issue62-host", host: "/?marker@example.invalid", method: "GET", wantError: true},
		{name: "issue62-upload-host", host: "/?marker@example.invalid", method: "POST", body: "synthetic payload", wantError: true},
		{name: "invalid-uplink-method", host: "example.invalid", method: "INVALID METHOD", body: "synthetic payload", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			u := url.URL{Scheme: "https", Host: tc.host, Path: "/"}
			options := &option.V2RayXHTTPBaseOptions{Host: tc.host, Path: "/", NoGRPCHeader: tc.noGRPC}
			req, err := newStreamRequest(ctx, tc.method, u.String(), body, tc.session, options)
			if tc.wantError {
				if err == nil || req != nil {
					t.Fatalf("want construction error and nil request, got req=%v err=%v", req, err)
				}
				return
			}
			if err != nil || req == nil {
				t.Fatalf("construction failed: req=%v err=%v", req, err)
			}
			if req.Context() != ctx || req.Context().Err() != nil {
				t.Fatal("helper changed or canceled its caller-owned context")
			}
			if req.Method != tc.method || req.Host != tc.host {
				t.Errorf("method/host changed: %s %s", req.Method, req.Host)
			}
			if req.Header.Get("Referer") == "" {
				t.Error("missing padding header")
			}
			if tc.body != "" {
				actual, err := io.ReadAll(req.Body)
				if err != nil || string(actual) != tc.body {
					t.Fatalf("body changed: %q, %v", actual, err)
				}
				wantType := "application/grpc"
				if tc.noGRPC {
					wantType = ""
				}
				if req.Header.Get("Content-Type") != wantType {
					t.Errorf("Content-Type=%q, want %q", req.Header.Get("Content-Type"), wantType)
				}
			} else if req.Body != nil && req.Body != http.NoBody {
				t.Error("unexpected download body")
			}
			if tc.session != "" && req.URL.Path != "/"+tc.session {
				t.Errorf("session path=%q", req.URL.Path)
			}
			cancel()
			if req.Context().Err() != context.Canceled {
				t.Error("explicit owner cancellation did not reach request")
			}
		})
	}
}
