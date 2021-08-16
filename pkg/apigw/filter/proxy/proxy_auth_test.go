package proxy

import (
	"net/http"
	"testing"

	"github.com/cortezaproject/corteza-server/pkg/apigw/types"
	"github.com/stretchr/testify/require"
)

func Test_authDo(t *testing.T) {
	type (
		tf struct {
			name   string
			err    string
			errv   string
			params proxyAuthParams
			exp    http.Header
		}
	)

	var (
		tcc = []tf{
			{
				name: "auth header match headers",
				params: proxyAuthParams{
					Type: proxyAuthTypeHeader,
					Params: map[string]interface{}{
						"Client-Id":          "123455",
						"Client_credentials": "pass1234",
					},
				},
				exp: http.Header{
					"Client-Id":          []string{"123455"},
					"Client_credentials": []string{"pass1234"},
				},
			},
			{
				name: "auth header match canonicalized headers",
				params: proxyAuthParams{
					Type: proxyAuthTypeHeader,
					Params: map[string]interface{}{
						"camelCaseHeader": "123455",
					},
				},
				exp: http.Header{
					"Camelcaseheader": []string{"123455"},
				},
			},
			{
				name: "auth basic match headers",
				params: proxyAuthParams{
					Type: proxyAuthTypeBasic,
					Params: map[string]interface{}{
						"username": "user",
						"password": "pass1234",
					},
				},
				exp: http.Header{"Authorization": []string{"Basic dXNlcjpwYXNzMTIzNA=="}},
			},
			{
				name: "auth basic match headers fail user validation",
				params: proxyAuthParams{
					Type:   proxyAuthTypeBasic,
					Params: map[string]interface{}{"password": "pass1234"},
				},
				exp:  http.Header{},
				errv: "invalid param username",
			},
			{
				name: "auth basic match headers fail pass validation",
				params: proxyAuthParams{
					Type:   proxyAuthTypeBasic,
					Params: map[string]interface{}{"username": "user"},
				},
				exp:  http.Header{},
				errv: "invalid param password",
			},
			{
				name:   "noop default fallback",
				params: proxyAuthParams{},
				exp:    http.Header{},
			},
		}
	)

	for _, tc := range tcc {
		t.Run(tc.name, func(t *testing.T) {
			var (
				req = require.New(t)
				c   = http.DefaultClient
			)

			c.Transport = types.MockRoundTripper(func(r *http.Request) (rs *http.Response, err error) { return })

			rq, _ := http.NewRequest("POST", "/foo", http.NoBody)

			authServicer, err := NewProxyAuthServicer(c, tc.params, struct{}{})

			if tc.errv != "" {
				req.EqualError(err, tc.errv)
				return
			}

			err = authServicer.Do(rq)

			if tc.err != "" {
				req.EqualError(err, tc.err)
			} else {
				req.Equal(tc.exp, rq.Header)
			}
		})
	}
}