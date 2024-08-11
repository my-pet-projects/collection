package web

import (
	"net/http"

	"github.com/a-h/templ"
)

type HandlerFunc func(reqResp *ReqRespPair) error

func Handler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kit := &ReqRespPair{
			Response: w,
			Request:  r,
		}
		if err := h(kit); err != nil {
			kit.Text(http.StatusInternalServerError, err.Error())
			return
		}
	}
}

type ReqRespPair struct {
	Response http.ResponseWriter
	Request  *http.Request
}

func (rrp *ReqRespPair) Text(status int, msg string) error {
	rrp.Response.WriteHeader(status)
	rrp.Response.Header().Set("Content-Type", "text/plain")
	_, err := rrp.Response.Write([]byte(msg))
	return err
}

func (rrp *ReqRespPair) Render(c templ.Component) error {
	return c.Render(rrp.Request.Context(), rrp.Response)
}
