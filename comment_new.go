package main

import (
	//"fmt"
	//"image"
	//"image/jpeg"
	"io/ioutil"
	"log"
	//"mime/multipart"
	"net/http"
	//"os"
	//"strconv"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	//"path/filepath"

	mf "github.com/mixamarciv/gofncstd3000"

	//"math/rand"

	//"github.com/nfnt/resize"
)

func http_get_newmessage2(r render.Render, session sessions.Session) {
	var m = map[string]interface{}{"cnt": 0}

	post := GetSessJson(session, "post", "{}")
	if _, ok := post["uuid"]; !ok {
		post["uuid"] = mf.StrUuid()
		post["time"] = mf.CurTimeStrShort()
		SetSessJson(session, "post", post)
	}
	m["user"] = GetSessJson(session, "user", "{}")
	m["post"] = post

	if imgs, ok := post["imagesuploaded"]; ok {
		post["imagesuploaded_jsonstr"] = mf.ToJsonStr(imgs)
	}
	r.HTML(200, "messagenew", m)
}

func http_post_comment_new_savesession(req *http.Request, session sessions.Session) string {
	var m = map[string]interface{}{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		m["error"] = "ОШИБКА загрузки параметров: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}
	SetSessStr(session, "comment", string(body))
	return "{\"success\":1}"
}

func http_post_comment_new(req *http.Request, session sessions.Session) string {
	var m = map[string]interface{}{"cnt": 0}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		m["error"] = "ОШИБКА загрузки параметров: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}
	log.Println(string(body))

	js, err := mf.FromJson(body)
	if err != nil {
		m["error"] = "ОШИБКА разбора параметров: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}

	js["acttype"] = "comment"
	check_and_register_user_in_sess_and_db(js, session)

	log.Println(js["userdata"])

	save_comment_data(js)

	u := js["userdata"].(map[string]interface{})
	js["info"] = interface{}(string("Спасибо, ваш комментарий успешно загружен<br>\n"))

	isactiveuser := 1
	if vi, okint := u["isactive"].(int); okint && vi == 0 {
		isactiveuser = 0
	} else if vf, okfloat64 := u["isactive"].(float64); okfloat64 && vf == 0 {
		isactiveuser = 0
	}
	if isactiveuser == 0 {
		js["info"] = interface{}(string("Спасибо, ваш комментарий успешно загружен, для его публикации подтвердите ваш email " +
			" (пройдите по ссылке отправленной вам на " + js["email"].(string) + ")<br>\n"))
		js["warning"] = interface{}(string("Комментарий будет опубликован <b>только после подтверждения</b> вашего email " + js["email"].(string)))
	}

	{
		msgi := js["emailsend_msg"]
		if msgi != nil {
			msg := msgi.(string)
			if msg != "" && len(msg) > 0 {
				mailto := js["email"].(string)
				LogPrint("Отправляем письмо на " + mailto)
				subject := js["emailsend_sbj"].(string)
				go SendMail(mailto, subject, msg, nil)
			}
		}
	}

	SetSessStr(session, "comment", string("")) //затираем данные сессии, что бы пользователь дважды не создал один и тот же пост
	{
		//создаем новый uuid для нового комментария и сохраняем его в текущей сессии
		c := map[string]interface{}{}
		c["uuid"] = mf.StrUuid()
		c["time"] = mf.CurTimeStrShort()
		SetSessJson(session, "comment", c)
		js["comment"] = c
	}
	retstr := mf.ToJsonStr(js)
	return retstr
}

func save_comment_data(js map[string]interface{}) {
	u := js["userdata"].(map[string]interface{})

	//сохраняем данные пользователя без activecodepass
	u2 := make(map[string]interface{})
	for k2, v2 := range u {
		switch k2 {
		case "activecodepass":
			continue
		}
		u2[k2] = v2
	}

	posttime := mf.CurTimeStrShort()
	query := "INSERT INTO tcomment(uuid_post,uuid_user,uuid,userdata,text,upddate,commentdate,commentdatet,isactive,activecode) "
	query += "VALUES(?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?,?)"
	_, err := db.Exec(query, js["uuid_post"], u["uuid"], js["uuid"], mf.ToJsonStr(u2), js["posttext"], mf.CurTimeStrShort(), js["time"], u["isactive"], u["activecode"])
	LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)

	imgs := js["imagesuploaded"].([]interface{})
	for _, imgi := range imgs {
		img, ok := imgi.(map[string]interface{})
		if !ok {
			continue
		}
		//log.Printf("%#v\n", img)
		query := "INSERT INTO timage(uuid_comment,uuid,hash,title,path,pathmin,imgdate,imgdatet) "
		query += "VALUES(?,?,?,?,?,?,?,CURRENT_TIMESTAMP)"
		_, err := db.Exec(query, js["uuid"], mf.StrUuid(), "nohash", img["text"], img["path"], img["pathmin"], posttime)
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
	}

	//SendMailNewPostsToWork() //новые комментарии пока не отправляем в работу..
}
