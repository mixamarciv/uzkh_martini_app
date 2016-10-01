package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	"path/filepath"

	mf "github.com/mixamarciv/gofncstd3000"

	"math/rand"

	"github.com/nfnt/resize"
)

var maxImageSize = 1280
var minImageSize = 160

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

func http_get_newmessage(r render.Render, session sessions.Session) {
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

func http_post_newmessagesavesession(req *http.Request, session sessions.Session) string {
	var m = map[string]interface{}{"cnt": 0}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		m["error"] = "ОШИБКА загрузки параметров: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}
	SetSessStr(session, "post", string(body))
	return "{\"success\":1}"
}

/************
//активация аккаунта пользователя
func http_get_activecode(params martini.Params, session sessions.Session, r render.Render) {
	activecode := params["activecode"]
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

	SetSessJson(session, "user", u)

	var js = map[string]interface{}{}
	js["user"] = u

	msg := "Активация учетной записи " + u["fam"].(string) + " " + u["name"].(string) + " " + u["pat"].(string) +
		" прошла успешно.\n Все ваши сообщения опубликованы."
	js["success"] = msg
	r.HTML(200, "user_activate", js)
}
**************/

type UploadForm struct {
	Uuid string                  `form:"uuid"`
	Time string                  `form:"time"`
	Path string                  `form:"path"`
	File []*multipart.FileHeader `form:"file"`
}

func http_post_uploadfile(uf UploadForm) string {

	retall := map[string]interface{}{"cnt": len(uf.File)}

	for i := 0; i < len(uf.File); i++ {

		file, err := uf.File[i].Open()

		/**************************
		log.Printf("ERR1: %#v", err)
		log.Printf("uuid: %#v", uf.Uuid)
		log.Printf("time: %#v", uf.Time)
		log.Printf("Path: %#v", uf.Path)
		log.Printf("File: %#v", file)
		***************************/

		t := uf.Time
		if len(t) < 15 || len(uf.Uuid) < 36 {
			return "{\"error\":\"ОШИБКА формата загрузки файлов (" + fmt.Sprintf("%d/15; %d/36", len(t), len(uf.Uuid)) + ")\"}"
		}
		apppath := mf.AppPath2() + "/public"
		crc32uuidstr := mf.StrCrc32([]byte(uf.Uuid))
		ipath := /*apppath +*/ "/upload/" + t[0:4] + "/" + t[4:6] + "/" + t[6:8] + "/" + t[9:15] + "_" + crc32uuidstr

		mf.MkdirAll(apppath + ipath)

		ifilename := mf.StrCrc32([]byte(mf.StrUuid()))
		{
			b, err := ioutil.ReadAll(file)
			LogPrintErrAndExit("ERROR ioutil.ReadAll", err)
			ifilename = mf.StrCrc32(b)
			_, err = file.Seek(0, 0)
			LogPrintErrAndExit("ERROR file.Seek0", err)
		}

		ifilepath := ipath + "/" + ifilename + filepath.Ext(uf.Path)
		ifilepath_min := ipath + "/" + ifilename + "_min" + filepath.Ext(uf.Path)
		//-------------------------------

		im, _, err := image.DecodeConfig(file)
		if err != nil {
			return "{\"error\":\"ОШИБКА к загрузке доступны только фото формата JPEG\"}"
		}
		LogPrintErrAndExit("ERROR image.DecodeConfig", err)

		_, err = file.Seek(0, 0)
		LogPrintErrAndExit("ERROR file.Seek1", err)

		iwidth := im.Width
		iheight := im.Height
		if iwidth >= iheight && iwidth > maxImageSize {
			iwidth = maxImageSize
			iheight = 0
		} else if iheight >= iwidth && iheight > maxImageSize {
			iheight = maxImageSize
			iwidth = 0
		}

		img, err := jpeg.Decode(file)
		LogPrintErrAndExit("ERROR jpeg.Decode", err)

		m := resize.Resize(uint(iwidth), uint(iheight), img, resize.Lanczos3)

		out, err := os.Create(apppath + ifilepath)
		LogPrintErrAndExit("ERROR os.Create", err)

		jpeg.Encode(out, m, nil)
		defer out.Close()

		//-- min image ---------------
		iwidth = im.Width
		iheight = im.Height
		minimize := 0
		if iwidth >= iheight && iwidth > minImageSize {
			iwidth = minImageSize
			iheight = 0
			minimize = 1
		} else if iheight >= iwidth && iheight > minImageSize {
			iheight = minImageSize
			iwidth = 0
			minimize = 1
		}

		log.Printf("minimize: %#v, %d / %d", minimize, iwidth, iheight)

		if minimize == 1 {
			m := resize.Resize(uint(iwidth), uint(iheight), img, resize.Lanczos3)
			out, err := os.Create(apppath + ifilepath_min)
			LogPrintErrAndExit("ERROR os.Create", err)
			defer out.Close()
			jpeg.Encode(out, m, nil)
		} else {
			ifilepath_min = ifilepath
		}

		ret := map[string]interface{}{"path": ifilepath, "pathmin": ifilepath_min}
		retall[strconv.Itoa(i)] = ret
	}

	retstr := mf.ToJsonStr(retall)

	return retstr
}

func http_post_newmessagesend(req *http.Request, session sessions.Session) string {
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

	check_and_register_user_in_sess_and_db(js, session)

	save_post_data(js)

	u := js["userdata"].(map[string]interface{})
	js["info"] = interface{}(string("Спасибо, ваше заявление успешно загружено, информация о рассмотрении будет направлена вам на " + js["email"].(string) + "<br>\n"))

	isactiveuser := 1
	if vi, okint := u["isactive"].(int); okint && vi == 0 {
		isactiveuser = 0
	} else if vf, okfloat64 := u["isactive"].(float64); okfloat64 && vf == 0 {
		isactiveuser = 0
	}
	if isactiveuser == 0 {
		js["info"] = interface{}(string("Спасибо, ваше заявление успешно загружено, для его публикации и рассмотрения подтвердите ваш email (пройдите по ссылке отправленной вам на " + js["email"].(string) + ")<br>\n"))
	}

	{
		msgi := js["emailsend_msg"]
		if msgi != nil {
			msg := msgi.(string)
			if msg != "" && len(msg) > 0 {
				mailto := js["email"].(string)
				LogPrint("Отправляем письмо на " + mailto)
				subject := js["emailsend_sbj"].(string)
				go SendMail(mailto, subject, msg)
			}
		}
	}

	SetSessStr(session, "post", string("")) //затираем данные сессии, что бы пользователь дважды не создал один и тот же пост
	retstr := mf.ToJsonStr(js)
	return retstr
}

func save_post_data(js map[string]interface{}) {
	u := js["userdata"].(map[string]interface{})
	posttime := mf.CurTimeStrShort()
	query := "INSERT INTO tpost(uuid_user,uuid,userdata,text,upddate,postdate,postdatet,isactive) "
	query += "VALUES(?,?,?,?,?,?,CURRENT_TIMESTAMP,?)"
	_, err := db.Exec(query, u["uuid"], js["uuid"], mf.ToJsonStr(u), js["posttext"], mf.CurTimeStrShort(), js["time"], u["isactive"])
	LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)

	imgs := js["imagesuploaded"].([]interface{})
	for _, imgi := range imgs {
		img := imgi.(map[string]interface{})
		//log.Printf("%#v\n", img)
		query := "INSERT INTO timage(uuid_post,uuid,hash,title,path,pathmin,imgdate,imgdatet) "
		query += "VALUES(?,?,?,?,?,?,?,CURRENT_TIMESTAMP)"
		_, err := db.Exec(query, js["uuid"], mf.StrUuid(), "nohash", img["text"], img["path"], img["pathmin"], posttime)
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
	}

}

// проверяем наличие и если надо регистрируем нового пользователя в бд
func check_and_register_user_in_sess_and_db(js map[string]interface{}, session sessions.Session) {
	user := GetSessJson(session, "user", "{}")
	if _, ok := user["uuid"]; !ok { //если пользователь не авторизован
		register_new_user_in_sess_and_db(js, session)
	} else if sess_email, ok := user["email"]; !ok || sess_email.(string) != js["email"].(string) {
		register_new_user_in_sess_and_db(js, session)
	} else {
		update_user_in_sess_and_db(js, user, session)
	}
}

func register_new_user_in_sess_and_db(js map[string]interface{}, session sessions.Session) {
	var u = map[string]interface{}{}
	u["uuid"] = mf.StrUuid()
	u["fam"] = js["fam"]
	u["name"] = js["name"]
	u["pat"] = js["pat"]
	u["email"] = js["email"]
	u["phone"] = js["phone"]
	u["street"] = js["street"]
	u["house"] = js["house"]
	u["flat"] = js["flat"]
	u["istemp"] = int(0)
	u["isactive"] = int(0)

	var n, istemp int
	var db_uuid_user, db_pass string
	{
		query := "SELECT COUNT(*),COALESCE(MAX(uuid),'-'),COALESCE(MAX(pass),'-'),COALESCE(MAX(istemp)+1,1) FROM tuser WHERE email=?"
		stmt, err := db.Prepare(query)
		LogPrintErrAndExit("ERROR db.Prepare: \n"+query+"\n\n", err)
		email := u["email"].(string)
		err = stmt.QueryRow(email).Scan(&n, &db_uuid_user, &db_pass, &istemp)
		LogPrintErrAndExit("ERROR stmt.QueryRow(email).Scan(&n): \n"+query+"\n\n", err)
	}
	if n == 0 { //если такой email не существует, то создаем нового пользователя
		pass := genPassword(6, 10)
		u["regdate"] = mf.CurTimeStrShort()

		activecode := mf.StrUuid()

		query := "INSERT INTO tuser(upddate,uuid,type,fam,name,pat,email,phone,pass,street,house,flat,info,regdate,regdatet,isactive,activecode,istemp) "
		query += "VALUES(?,?,0,?,?,?,LOWER(?),?,?,?,?,?,?,?,CURRENT_TIMESTAMP,0,?,0)"
		_, err := db.Exec(query, u["regdate"], u["uuid"], u["fam"], u["name"], u["pat"], u["email"], u["phone"], pass, u["street"], u["house"], u["flat"], "{}", u["regdate"], activecode)
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)

		urlactiv := "http://" + sitedomain + "/useractivecode/" + activecode
		msg := "Для публикации и отправки вашего сообщения: <br>\n"
		msg += "\"" + js["posttext"].(string) + "\"<br>\n"
		msg += "от имени " + u["fam"].(string) + " " + u["name"].(string) + " " + u["pat"].(string) + " <br>\n"
		msg += "а так же для подтверждения этого email и активации вашей учетной записи<br>\n" //на сайте " + sitedomain + "
		msg += "пройдите по ссылке <a href=\"" + urlactiv + "\">" + urlactiv + "</a><br><br>\n\n"
		msg += "В дальнейшем для входа на сайт " + sitedomain + " вы можете использовать следующий<br>\n"
		msg += "логин: " + u["email"].(string) + "<br>\n"
		msg += "пароль: " + pass + "<br>\n"
		//msg += "<br>\nесли вы не писали никаких сообщений то удалите это письмо<br>\n"
		msg += "<br><br>\n--<br>\nС Уважением Администрация сайта " + sitedomain + "<br>\n"

		js["emailsend_msg"] = interface{}(msg)
		js["emailsend_sbj"] = interface{}(sitedomain + " запрос подтверждения email и отправки сообщения")
		js["userdata"] = u

		SetSessJson(session, "user", u)
		return
	}

	//если email существует но пользователь не авторизован
	//то сохраняем его новые данные во временную запись до того как он подтвердит свою учетную запись
	u["uuid"] = db_uuid_user
	u["istemp"] = istemp
	activecode := mf.StrUuid()
	query := "INSERT INTO tuser(upddate,uuid,type,"
	query += "fam,name,pat,"
	query += "email,phone,pass,"
	query += "street,house,flat,"
	query += "info,activecode,istemp) "
	query += "VALUES(?,?,0," //upddate,uuid,type
	query += "?,?,?,"        //fam,name,pat
	query += "LOWER(?),?,?," //email,phone,pass
	query += "?,?,?,"        //street,house,flat
	query += "?,?,?)"        //info,activecode,istemp
	_, err := db.Exec(query,
		mf.CurTimeStrShort(), u["uuid"], //upddate,uuid,type
		u["fam"], u["name"], u["pat"], //fam,name,pat
		u["email"], u["phone"], db_pass, //email,phone,pass
		u["street"], u["house"], u["flat"], //street,house,flat
		"{}", activecode, u["istemp"]) //info,activecode,istemp
	LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)

	urlactiv := "http://" + sitedomain + "/useractivecode/" + activecode
	msg := "Для публикации и отправки вашего сообщения: <br>\n"
	msg += "\"" + js["posttext"].(string) + "\"<br>\n"
	msg += "от имени " + u["fam"].(string) + " " + u["name"].(string) + " " + u["pat"].(string) + " <br>\n"
	msg += "пройдите по ссылке <a href=\"" + urlactiv + "\">" + urlactiv + "</a><br><br>\n\n"
	msg += "Для входа на сайт " + sitedomain + " вы можете использовать ваши<br>\n"
	msg += "логин: " + u["email"].(string) + "<br>\n"
	msg += "пароль: " + db_pass + "<br>\n"
	msg += "<br>\nесли вы не писали никаких сообщений то удалите это письмо<br>\n"
	msg += "<br><br>\n\n--<br>\nС Уважением Администрация сайта " + sitedomain + "<br>\n"

	js["emailsend_msg"] = interface{}(msg)
	js["emailsend_sbj"] = interface{}(sitedomain + " запрос подтверждения отправки сообщения")
	js["userdata"] = u

	SetSessJson(session, "user", u)
}

//обновляем данные пользователя в текущей сессии и в базе данных
func update_user_in_sess_and_db(js, u map[string]interface{}, session sessions.Session) {
	//var u = map[string]interface{}{}
	u["fam"] = js["fam"]
	u["name"] = js["name"]
	u["pat"] = js["pat"]
	u["phone"] = js["phone"]
	u["street"] = js["street"]
	u["house"] = js["house"]
	u["flat"] = js["flat"]

	//if u["isactive"].(int) > 0
	{
		query := "UPDATE tuser SET fam=?,name=?,pat=?,phone=?,street=?,house=?,flat=?,info=?,upddate=? "
		query += "WHERE email=LOWER(?) AND istemp=?"
		_, err := db.Exec(query, u["fam"], u["name"], u["pat"], u["phone"], u["street"], u["house"], u["flat"], "{}", mf.CurTimeStrShort(), u["email"], u["istemp"])
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
	}

	js["userdata"] = u

	SetSessJson(session, "user", u)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func genPassword(minlen, maxlen int) string {
	n := minlen + rand.Intn(maxlen-minlen)
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
