package main

import (
	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"

	"net/http"
	"time"

	mf "github.com/mixamarciv/gofncstd3000"
)

func http_get_messageview(params martini.Params, session sessions.Session, r render.Render) {
	var js = map[string]interface{}{}
	u := GetSessJson(session, "user", "{}")
	js["user"] = u

	c := GetSessJson(session, "comment", "{}") //данные комментария который набивал пользователь
	js["comment"] = c
	if _, ok := c["uuid"]; !ok {
		c["uuid"] = mf.StrUuid()
		c["time"] = mf.CurTimeStrShort()
		SetSessJson(session, "comment", c)
	}
	if imgs, ok := c["imagesuploaded"]; ok {
		c["imagesuploaded_jsonstr"] = mf.ToJsonStr(imgs)
	}

	uuid_post := params["uuid"]

	js["post"] = load_post(uuid_post)
	//js["comments"] = load_post(pagen)

	r.HTML(200, "messageview", js)
}

func load_post(uuid_post string) map[string]interface{} {
	m := map[string]interface{}{}
	query := "SELECT uuid,type,userdata,text,postdatet,postdate,"
	query += "(SELECT COUNT(*) FROM timage t WHERE t.uuid_post=p.uuid), 0 "
	query += " FROM tpost p WHERE isactive=1 AND uuid=? "
	rows, err := db.Query(query, uuid_post)
	if err != nil {
		LogPrintErrAndExit("ERROR db.Query(query): \n"+query+"\n\n", err)
	}
	for rows.Next() {
		var uuid, ptype, userdata, text, postdates NullString
		var postdatet time.Time
		var imgcnt, commentcnt int
		if err := rows.Scan(&uuid, &ptype, &userdata, &text, &postdatet, &postdates, &imgcnt, &commentcnt); err != nil {
			LogPrintErrAndExit("ERROR rows.Scan: \n"+query+"\n\n", err)
		}
		m = map[string]interface{}{"uuid": uuid.get("-"), "ptype": ptype.get("-"), "text": post_text_to_html(text.get("-"))}
		m["userdata"] = mf.FromJsonStr([]byte(userdata.get("{}")))
		m["postdatet"] = postdatet
		m["postdatefmt"] = postdatet.Format("02.01.2006 15:04")
		m["postdates"] = postdates
		m["imgcnt"] = imgcnt
		if imgcnt > 0 {
			m["images"] = load_posts_or_comment_images_arr(m["uuid"].(string), "post", 0)
		}
	}
	rows.Close()
	return m
}

//обработка запросов списка комментариев
func http_post_commentsview(req *http.Request, session sessions.Session) string {
	p := ParseBodyParams(req)

	if _, ok := p["error"]; ok {
		return mf.ToJsonStr(p)
	}

	reqtype := p["type"].(string)

	if reqtype == "postallcomments" {
		return user_req_postallcomments(p, req, session)
	}

	//в случае если ни один из вариантов обработки не прошел возвращаем ошибку:
	p["error"] = string("ОШИБКА3000: не верно заданы параметры запроса")
	return mf.ToJsonStr(p)
}

//возвращаем список всех комментариев для заданного поста
func user_req_postallcomments(p map[string]interface{}, req *http.Request, session sessions.Session) string {

	ret := make([]map[string]interface{}, 0)

	uuid_post, ok := p["uuid_post"]
	if !ok {
		return "{\"error\":\"uuid_post не задан\"}"
	}

	query := "SELECT uuid,uuid_parent,iif(ishideuser=1,'{}',userdata) AS userdata,text,commentdatet, "
	query += "(SELECT COUNT(*) FROM timage t WHERE t.uuid_comment=p.uuid) "
	query += "FROM tcomment p "
	query += "WHERE uuid_post=? AND isactive=1 AND ishide=0 "
	query += "ORDER BY commentdatet "
	rows, err := db.Query(query, uuid_post)
	if err != nil {
		LogPrintErrAndExit("ERROR db.Query(query): \n"+query+"\n\n", err)
	}

	cnt_rows := 0
	for rows.Next() {
		cnt_rows++
		var uuid, uuid_parent, userdata, text NullString
		var commentdatet time.Time
		var imgcnt int
		if err := rows.Scan(&uuid, &uuid_parent, &userdata, &text, &commentdatet, &imgcnt); err != nil {
			LogPrintErrAndExit("ERROR rows.Scan: \n"+query+"\n\n", err)
		}
		m := map[string]interface{}{"uuid": uuid.get(""), "uuid_parent": uuid_parent.get(""), "text": post_text_to_html(text.get("-"))}
		m["userdata"] = mf.FromJsonStr([]byte(userdata.get("{}")))
		for k, _ := range m["userdata"].(map[string]interface{}) {
			switch k {
			case "phone", "email":
				delete(m["userdata"].(map[string]interface{}), k)
			}
		}
		m["datefmt"] = commentdatet.Format("02.01.2006 15:04")
		m["imgcnt"] = imgcnt
		if imgcnt > 0 {
			m["images"] = load_posts_or_comment_images_arr(m["uuid"].(string), "comment", cnt_rows)
		}
		ret = append(ret, m)
	}
	rows.Close()

	test := map[string]interface{}{"uuid_post": uuid_post, "cnt_rows": cnt_rows, "query": query}
	ret = append(ret, test)

	return mf.ToJsonStr(ret)
}
