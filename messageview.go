package main

import (
	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	//mf "github.com/mixamarciv/gofncstd3000"
)

func http_get_messageview(session sessions.Session, r render.Render) {
	error503(session, r)
}
