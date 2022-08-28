package main

import (
	"everest/everest"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
)

const (
	Interval = 200
	Port     = ":8080"
)

func main() {
	service := everest.NewService()

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/request":
			service.RequestHandler(ctx)
		case "/admin/requests":
			service.AdminHandler(ctx)
		default:
			pprofhandler.PprofHandler(ctx)
			// ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	service.Populate()
	go service.Ticker(Interval * time.Millisecond)

	fasthttp.ListenAndServe(Port, m)
}
