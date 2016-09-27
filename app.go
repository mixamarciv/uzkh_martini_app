package main

import (
	"log"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/sessions"

	"html/template"
	"strconv"

	//mf "github.com/mixamarciv/gofncstd3000"
)

func init() {
	InitLog()
	InitDb()
}

func main() {
	var m *martini.ClassicMartini = martini.Classic()
	//m := martini.Classic() // martini.ClassicMartini

	//--- log  ---------------------------------------------------------
	m.Use(func(c martini.Context, log *log.Logger) {
		//log.Println("before a request")
		c.Next()
		//log.Println("after a request")
	})
	//--- /log ---------------------------------------------------------

	//m.Use(auth.Basic("test", "123"))

	//--- static  ---------------------------------------------------------
	m.Use(martini.Static("public"))
	//--- /static ---------------------------------------------------------

	//--- session ---------------------------------------------------------
	store := sessions.NewCookieStore([]byte("secret1234"))
	m.Use(sessions.Sessions("uzkhsess", store))
	//--- /session --------------------------------------------------------

	//--- render  ---------------------------------------------------------
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",                // Specify what path to load the templates from.
		Layout:     "maintemplate",             // Specify a layout template. Layouts can call {{ yield }} to render the current template.
		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
		Funcs:      []template.FuncMap{},       // Specify helper function maps for templates to access.
		Delims:     render.Delims{"{{", "}}"},  // Sets delimiters to the specified strings.
		Charset:    "UTF-8",                    // Sets encoding for json and html content-types. Default is "UTF-8".
		IndentJSON: true,                       // Output human readable JSON
	}))
	//--- /render --------------------------------------------------------

	m.Get("/", func(r render.Render, session sessions.Session) {
		v := session.Get("cnt")
		var a string
		if v == nil {
			a = "0"
		} else {
			a = v.(string)
		}
		cnt, err := strconv.Atoi(a)
		LogPrintErrAndExit("strconv.Atoi(v) error: \n"+a+"\n\n", err)
		session.Set("cnt", strconv.Itoa(cnt+1))
		r.HTML(200, "main", map[string]interface{}{"hello": "world", "cnt": a})
	})

	m.Get("/about", func(r render.Render, session sessions.Session) {
		r.HTML(200, "about", map[string]interface{}{"cnt": 0})
	})

	m.Get("/newmessage", newmessage)
	m.Post("/newmessagesend", newmessagesend)
	m.Post("/uploadfile", binding.MultipartForm(UploadForm{}), uploadfile)
	//--- /fileupload -----------------------------------------------------

	m.RunOnAddr(":8091")
}
