package main

import (
	"database/sql"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	mf "github.com/mixamarciv/gofncstd3000"

	"github.com/go-martini/martini"

	"net/http"
)

//активация аккаунта пользователя
func http_get_useractivatecode(params martini.Params, session sessions.Session, r render.Render) {
	acttype := params["acttype"]
	activecode := params["activecode"]
	LogPrint("activecode: " + activecode)
	var u = map[string]interface{}{}
	{
		var uuid, fam, name, pat, email, phone, street, house, flat, info, utype string
		var isactive, istemp int

		query := "SELECT uuid,(SELECT name FROM tuser_type t WHERE t.type=u.type),"
		query += "fam,name,pat,email,phone,street,house,flat,info,isactive,istemp FROM tuser u WHERE activecode=?"
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
		u["activecodepass"] = user_update_activecodepass(u["uuid"].(string))
	}

	{ //обновляем сообщения которые он отправлял
		query := "UPDATE t" + acttype + " SET isactive=1 "
		query += "WHERE uuid_user=? AND isactive=0 AND activecode=?"
		_, err := db.Exec(query, u["uuid"], activecode)
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
	}

	SendMailNewPostsToWork()

	SetSessJson(session, "user", u)

	var js = map[string]interface{}{}
	js["user"] = u

	msg := "Активация учетной записи " + u["fam"].(string) + " " + u["name"].(string) + " " + u["pat"].(string) +
		" прошла успешно.\n Все ваши сообщения опубликованы."
	js["success"] = msg
	r.HTML(200, "user_activate", js)
}

func error503(session sessions.Session, r render.Render) {
	var js = map[string]interface{}{}
	u := GetSessJson(session, "user", "{}")
	js["user"] = u
	r.HTML(200, "error503", js)
}

/********
func http_get_userlogin(session sessions.Session, r render.Render) {
	error503(session, r)
}

func http_post_userlogin(session sessions.Session, r render.Render) {
	error503(session, r)
}
********/
func http_get_userform(session sessions.Session, r render.Render) {
	var js = map[string]interface{}{}
	u := GetSessJson(session, "user", "{}")
	js["user"] = u
	r.HTML(200, "user_form", js)
}

func http_post_userform(req *http.Request, session sessions.Session) string {
	p := ParseBodyParams(req)

	if _, ok := p["error"]; ok {
		return mf.ToJsonStr(p)
	}

	if _, ok := p["type"]; !ok {
		p["error"] = string("ОШИБКА: не верно заданы параметры запроса")
		return mf.ToJsonStr(p)
	}

	reqtype := p["type"].(string)

	if reqtype == "auth" {
		return user_req_auth(p, req, session)
	}

	if reqtype == "edit" {
		return user_req_edit(p, req, session)
	}

	if reqtype == "logout" {
		return user_req_logout(p, req, session)
	}

	//в случае если ни один из вариантов обработки не прошел возвращаем ошибку:
	p["error"] = string("ОШИБКА3000: не верно заданы параметры запроса")
	return mf.ToJsonStr(p)
}

//logout пользователя
func user_req_logout(p map[string]interface{}, req *http.Request, session sessions.Session) string {
	u := GetSessJson(session, "user", "{}")
	session.Delete("user")

	if p["uuid"].(string) != u["uuid"].(string) {
		p["error"] = string("ошибка обновления, возможно данные вашей сессии устарели, выйдите и заново войдите под вашей учетной записью")
		//SetSessJson(session, "user", map[string]interface{}{})

		return mf.ToJsonStr(p)
	}

	user_update_activecodepass(p["uuid"].(string))

	return "{\"success\":\"Выход успешно выполнен\"}"
}

//обновление данных пользователя
func user_req_edit(p map[string]interface{}, req *http.Request, session sessions.Session) string {
	u := GetSessJson(session, "user", "{}")
	error_text := "ошибка обновления, возможно данные вашей сессии устарели, выйдите и заново войдите под вашей учетной записью"

	if p["uuid"].(string) != u["uuid"].(string) {
		p["error"] = error_text
		return mf.ToJsonStr(p)
	}

	{
		var uuid, fam, name, pat, email, phone, street, house, flat, info, pass string
		var utype string

		p_uuid := p["uuid"].(string)
		p_pass := p["pass"].(string)
		p_newpass := p["newpass"].(string)
		p_email := p["email"].(string)
		p_activecodepass := p["activecodepass"].(string)

		query := "SELECT uuid,(SELECT name FROM tuser_type t WHERE t.type=u.type),"
		query += "fam,name,pat,email,phone,street,house,flat,info,pass FROM tuser u WHERE uuid=? AND email=LOWER(?) AND activecodepass=? AND istemp=0"
		stmt, err := db.Prepare(query)
		LogPrintErrAndExit("ERROR db.Prepare: \n"+query+"\n\n", err)
		err = stmt.QueryRow(p_uuid, p_email, p_activecodepass).Scan(&uuid, &utype, &fam, &name, &pat, &email, &phone, &street, &house, &flat, &info, &pass)
		if err == sql.ErrNoRows {
			p["error"] = error_text
			return mf.ToJsonStr(p)
		}
		LogPrintErrAndExit("ERROR stmt.QueryRow(p_email, p_pass).Scan(...): \n"+query+"\n\n", err)

		if p_pass != "" && p_pass != pass {
			p["error"] = error_text
			return mf.ToJsonStr(p)
		} else if p_pass == "" { //если пароль не меняется
			p_newpass = pass
		}

		//обновляем данные пользователя
		curtime := mf.CurTimeStrShort()
		query = "UPDATE tuser SET lastvisit=?,fam=?,name=?,pat=?,phone=?,street=?,house=?,flat=?,pass=? WHERE uuid=? AND email=? AND istemp=0 "
		_, err = db.Exec(query, curtime, p["fam"], p["name"], p["pat"], p["phone"], p["street"], p["house"], p["flat"], p_newpass, p_uuid, p_email)
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)

		u["uuid"] = uuid
		u["type"] = utype
		u["fam"] = p["fam"]
		u["name"] = p["name"]
		u["pat"] = p["pat"]
		u["phone"] = p["phone"]
		u["street"] = p["street"]
		u["house"] = p["house"]
		u["flat"] = p["flat"]
	}

	SetSessJson(session, "user", u)

	return "{\"success\":\"данные успешно сохранены\"}"
}

//авторизация юзера
func user_req_auth(p map[string]interface{}, req *http.Request, session sessions.Session) string {
	var u = map[string]interface{}{}
	{
		var uuid, fam, name, pat, email, phone, street, house, flat, info, utype string
		var isactive, istemp int

		p_pass := p["pass"].(string)
		p_email := p["email"].(string)

		query := "SELECT uuid,(SELECT name FROM tuser_type t WHERE t.type=u.type),"
		query += "fam,name,pat,email,phone,street,house,flat,info,isactive,istemp FROM tuser u WHERE email=LOWER(?) AND pass=?"
		stmt, err := db.Prepare(query)
		LogPrintErrAndExit("ERROR db.Prepare: \n"+query+"\n\n", err)
		err = stmt.QueryRow(p_email, p_pass).Scan(&uuid, &utype, &fam, &name, &pat, &email, &phone, &street, &house, &flat, &info, &isactive, &istemp)
		if err == sql.ErrNoRows {
			p["error"] = string("пользователи с таким логином и паролем не найдены")
			return mf.ToJsonStr(p)
		}
		LogPrintErrAndExit("ERROR stmt.QueryRow(p_email, p_pass).Scan(...): \n"+query+"\n\n", err)

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
		u["activecodepass"] = user_update_activecodepass(u["uuid"].(string))
	}

	SetSessJson(session, "user", u)

	return "{\"success\":\"авторизация пройдена успешно. добро пожаловать )\"}"
}

func user_update_activecodepass(uuid_user string) string {
	//сохраняем в бд временный код для задания нового пароля пользователя
	new_uuid_activecodepass := mf.StrUuid()
	query := "UPDATE tuser SET activecodepass=?, lastvisit=? "
	query += "WHERE uuid=? AND istemp=0"
	_, err := db.Exec(query, new_uuid_activecodepass, mf.CurTimeStrShort(), uuid_user)
	LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
	return new_uuid_activecodepass
}
