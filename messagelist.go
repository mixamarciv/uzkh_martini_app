package main

import (
	"strconv"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	"time"

	mf "github.com/mixamarciv/gofncstd3000"

	"database/sql"

	"html"

	"github.com/go-martini/martini"
)

type NullString struct {
	sql.NullString
}

func (p *NullString) get(defaultval string) string {
	if p.Valid {
		return p.String
	}
	return defaultval
}

func load_posts_arr(page int) []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	query := "SELECT uuid,type,userdata,text,postdatet,postdate,"
	query += "(SELECT COUNT(*) FROM timage t WHERE t.uuid_post=p.uuid), "
	query += "(SELECT COUNT(*) FROM tcomment t WHERE t.uuid_post=p.uuid) "
	query += " FROM tpost p WHERE isactive=1 ORDER BY postdatet DESC"
	rows, err := db.Query(query)
	if err != nil {
		LogPrintErrAndExit("ERROR db.Query(query): \n"+query+"\n\n", err)
	}

	row_cnt := 0
	for rows.Next() {
		row_cnt++
		var uuid, ptype, userdata, text, postdates NullString
		var postdatet time.Time
		var imgcnt, commentcnt int
		if err := rows.Scan(&uuid, &ptype, &userdata, &text, &postdatet, &postdates, &imgcnt, &commentcnt); err != nil {
			LogPrintErrAndExit("ERROR rows.Scan: \n"+query+"\n\n", err)
		}
		m := map[string]interface{}{"uuid": uuid.get("-"), "ptype": ptype.get("-"), "text": post_text_to_html(text.get("-"))}
		m["userdata"] = mf.FromJsonStr([]byte(userdata.get("{}")))
		m["postdatet"] = postdatet
		m["postdatefmt"] = postdatet.Format("02.01.2006 15:04")
		m["postdates"] = postdates
		m["commentcnt"] = commentcnt
		m["imgcnt"] = imgcnt
		if imgcnt > 0 {
			m["images"] = load_posts_or_comment_images_arr(m["uuid"].(string), "post", row_cnt)
		}
		ret = append(ret, m)
	}
	rows.Close()
	return ret
}

func post_text_to_html(text string) string {
	text = html.EscapeString(text)
	//text = mf.StrRegexpReplace(text, "\\n", "<br>")
	return text
}

func load_posts_or_comment_images_arr(uuid_post string, post_or_comment string, i_post int) []map[string]string {
	ret := make([]map[string]string, 0)
	query := "SELECT uuid,title,path,pathmin "
	query += " FROM timage WHERE uuid_" + post_or_comment + "=? ORDER BY uuid"
	rows, err := db.Query(query, uuid_post)
	if err != nil {
		LogPrintErrAndExit("ERROR db.Query(query): \n"+query+"\n\n", err)
	}
	for rows.Next() {
		var uuid, title, path, pathmin NullString
		if err := rows.Scan(&uuid, &title, &path, &pathmin); err != nil {
			LogPrintErrAndExit("ERROR rows.Scan: \n"+query+"\n\n", err)
		}
		m := map[string]string{"uuid": uuid.get("-"), "title": title.get(""), "path": path.get(""), "pathmin": pathmin.get(""), "ipost": strconv.Itoa(i_post)}
		ret = append(ret, m)
	}
	rows.Close()
	return ret
}

//загружаем список сообщений
func http_get_messagelist(params martini.Params, session sessions.Session, r render.Render) {
	var js = map[string]interface{}{}
	u := GetSessJson(session, "user", "{}")
	js["user"] = u

	page := params["page"]
	js["page"] = page

	pagen, err := strconv.Atoi(page)
	if err != nil {
		//js["error"] = "не верный формат страницы " + page
		//r.HTML(200, "messagelist", js)
		//return
		pagen = 1
	}

	js["posts"] = load_posts_arr(pagen)

	js["page"] = page

	r.HTML(200, "messagelist", js)
}
