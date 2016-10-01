package main

import (
	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	//mf "github.com/mixamarciv/gofncstd3000"

	"github.com/go-martini/martini"
)

//загружаем список сообщений
func http_get_messagelist(params martini.Params, session sessions.Session, r render.Render) {
	//page := params["page"]
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
	u := GetSessJson(session, "user", "{}")

	var js = map[string]interface{}{}

	js["user"] = u

	r.HTML(200, "messagelist", js)
}
