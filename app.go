package main

import (
	"log"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/sessions"

	"html/template"
	"io/ioutil"
	"net/http"

	mf "github.com/mixamarciv/gofncstd3000"
)

//var sitedomain string = "192.168.1.120:8091"

func init() {
	InitLog()
	InitDb()
	initSendMail()
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
	store := sessions.NewCookieStore([]byte(secret_cookie_store))
	m.Use(sessions.Sessions(cookie_store_name, store))
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
		var js = map[string]interface{}{}
		u := GetSessJson(session, "user", "{}")
		js["user"] = u
		r.HTML(200, "main", js)
	})

	m.Get("/about", func(r render.Render, session sessions.Session) {
		var js = map[string]interface{}{}
		u := GetSessJson(session, "user", "{}")
		js["user"] = u
		r.HTML(200, "about", js)
	})

	m.Get("/messagenew", http_get_newmessage)
	m.Post("/newmessagesavesession", http_post_newmessagesavesession)
	m.Post("/newmessagesend", http_post_newmessagesend)
	m.Post("/uploadfile", binding.MultipartForm(UploadForm{}), http_post_uploadfile)

	m.Get("/useractivatecode/:acttype/:activecode", http_get_useractivatecode)
	m.Get("/user", http_get_userform)
	m.Post("/user", http_post_userform)

	m.Get("/messagelist", http_get_messagelist)
	m.Get("/messagelist/:page", http_get_messagelist)
	m.Get("/messageview/:uuid", http_get_messageview)
	m.Post("/commentsview", http_post_commentsview)

	m.Post("/comment_new_savesession", http_post_comment_new_savesession)
	m.Post("/comment_new", http_post_comment_new)
	//--- /fileupload -----------------------------------------------------

	m.RunOnAddr(run_on_addr)
}

//разбор параметров пост запроса в map[string]interface{}
func ParseBodyParams(req *http.Request) map[string]interface{} {
	var m = map[string]interface{}{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		m["error"] = "ОШИБКА загрузки параметров: " + mf.ErrStr(err)
		return m
	}
	//log.Println(string(body))

	js, err := mf.FromJson(body)
	if err != nil {
		m["error"] = "ОШИБКА разбора параметров: " + mf.ErrStr(err)
		return m
	}

	return js
}

func GetSessStr(session sessions.Session, varname, defaultval string) string {
	v := session.Get(varname)
	if v == nil {
		return defaultval
	}
	return v.(string)
}

func SetSessStr(session sessions.Session, varname, val string) {
	session.Set(varname, val)
}

func GetSessJson(session sessions.Session, varname, defaultval string) map[string]interface{} {
	v := session.Get(varname)
	if v == nil {
		j, err := mf.FromJson([]byte(defaultval))
		if err == nil {
			return j
		}
		m := map[string]interface{}{"error": mf.ErrStr(err)}
		return m
	}
	j, err := mf.FromJson([]byte(v.(string)))
	if err == nil {
		return j
	}
	m := map[string]interface{}{"error": mf.ErrStr(err)}
	return m
}

func SetSessJson(session sessions.Session, varname string, val map[string]interface{}) {
	session.Set(varname, mf.ToJsonStr(val))
}
