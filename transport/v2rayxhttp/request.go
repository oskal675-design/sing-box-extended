package xhttp

import (
	"context"
	"io"
	"net/http"

	"github.com/sagernet/sing-box/option"
)

// newStreamRequest constructs a request without starting transport work.
// The caller owns the context and must cancel it if construction fails.
func newStreamRequest(ctx context.Context, method, url string, body io.Reader, sessionId string, options *option.V2RayXHTTPBaseOptions) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	FillStreamRequest(req, sessionId, "", options)
	return req, nil
}
