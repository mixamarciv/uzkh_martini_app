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
	query := "SELECT uuid,type,userdata,text,postdatet,"
	query += "(SELECT COUNT(*) FROM timage t WHERE t.uuid_post=p.uuid) "
	query += " FROM tpost p WHERE isactive=1 ORDER BY postdatet DESC"
	rows, err := db.Query(query)
	if err != nil {
		LogPrintErrAndExit("ERROR db.Query(query): \n"+query+"\n\n", err)
	}
	for rows.Next() {
		var uuid, ptype, userdata, text, postdates NullString
		var postdatet time.Time
		var imgcnt int
		if err := rows.Scan(&uuid, &ptype, &userdata, &text, &postdatet, &imgcnt); err != nil {
			LogPrintErrAndExit("ERROR rows.Scan: \n"+query+"\n\n", err)
		}
		m := map[string]interface{}{"uuid": uuid.get("-"), "ptype": ptype.get("-"), "text": post_text_to_html(text.get("-"))}
		m["userdata"] = mf.FromJsonStr([]byte(userdata.get("{}")))
		m["postdatet"] = postdatet
		m["postdatefmt"] = postdatet.Format("02.01.2006 15:04")
		m["postdates"] = postdates
		m["imgcnt"] = imgcnt
		if imgcnt > 0 {
			m["images"] = load_posts_images_arr(m["uuid"].(string))
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

func load_posts_images_arr(uuid_post string) []map[string]string {
	ret := make([]map[string]string, 0)
	query := "SELECT uuid,title,path,pathmin "
	query += " FROM timage WHERE uuid_post=? ORDER BY uuid"
	rows, err := db.Query(query, uuid_post)
	if err != nil {
		LogPrintErrAndExit("ERROR db.Query(query): \n"+query+"\n\n", err)
	}
	for rows.Next() {
		var uuid, title, path, pathmin NullString
		if err := rows.Scan(&uuid, &title, &path, &pathmin); err != nil {
			LogPrintErrAndExit("ERROR rows.Scan: \n"+query+"\n\n", err)
		}
		m := map[string]string{"uuid": uuid.get("-"), "title": title.get(""), "path": path.get(""), "pathmin": pathmin.get("")}
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
	/***
		LogPrint("activecode: " + activecode)
		var u = map[string]interface{}{}
		{
			var uuid, fam, name, pat, email, phone, street, house, flat, info string
			var utype, isactive, istemp int

			query := "SELECT uuid,type,fam,name,pat,email,phone,street,house,flat,info,isactive,istemp FROM tuser WHERE activecode=?"
			stmt, err := db.Prepare(query)
			LogPrintErrAndExit("ERROR db.Prepare: \n"+query+"\n\n", err)
			err = stmt.QueryRow(activecode).Scan(&uuid, &utype, &fam, &name, &pat, &email, &phone, &street, &house, &flat, &info, &isactive, &istemp)
			if err == sql.ErrNoRows {
				var js = map[string]interface{}{}
				js["error"] = string("не верная ссылка для активации учетной записи")
				r.HTML(200, "user_activate", js)
				return
			}
			LogPrintErrAndExit("ERROR stmt.QueryRow(activecode).Scan(...): \n"+query+"\n\n", err)

			u["uuid"] = uuid
			u["type"] = utype
			u["email"] = email
			u["fam"] = fam
			u["name"] = name
			u["pat"] = pat
			u["phone"] = phone
			u["street"] = street
			u["house"] = house
			u["flat"] = flat
			u["isactive"] = isactive
			u["istemp"] = istemp
		}

		//if u["isactive"].(int) > 0
		{ //обновляем данные пользователя на те что он актвирует
			query := "UPDATE tuser SET fam=?,name=?,pat=?,phone=?,street=?,house=?,flat=?,upddate=?,isactive=1 "
			query += "WHERE email=LOWER(?) AND istemp=0"
			_, err := db.Exec(query, u["fam"], u["name"], u["pat"], u["phone"], u["street"], u["house"], u["flat"], mf.CurTimeStrShort(), u["email"])
			LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
			u["isactive"] = 1
		}

		{ //обновляем сообщения которые он отправлял
			query := "UPDATE tpost SET isactive=1 "
			query += "WHERE uuid_user=? AND isactive=0"
			_, err := db.Exec(query, u["uuid"])
			LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
		}
	****/

	js["page"] = page

	r.HTML(200, "messagelist", js)
}
