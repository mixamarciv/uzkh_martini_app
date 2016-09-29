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

func newmessage(r render.Render, session sessions.Session) {
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
	r.HTML(200, "newmessage", m)
}

func newmessagesavesession(req *http.Request, session sessions.Session) string {
	var m = map[string]interface{}{"cnt": 0}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		m["error"] = "ОШИБКА загрузки параметров: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}
	SetSessStr(session, "post", string(body))
	return "{\"success\":1}"
}

type UploadForm struct {
	Uuid string                  `form:"uuid"`
	Time string                  `form:"time"`
	Path string                  `form:"path"`
	File []*multipart.FileHeader `form:"file"`
}

func uploadfile(uf UploadForm) string {

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

func newmessagesend(req *http.Request, session sessions.Session) string {
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
	js["info"] = interface{}(string(""))

	check_and_register_user_in_sess_and_db(js, session)

	//SetSessStr(session, "post", string("")) //затираем данные сессии, что бы пользователь дважды не создал один и тот же пост
	js["info"] = interface{}(string("Спасибо, ваше заявление успешно отправлено, информация о рассмотрении прийдет вам на " + js["email"].(string) + "<br>\n" + js["info"].(string)))
	retstr := mf.ToJsonStr(js)
	return retstr
	/********
	err = req.ParseMultipartForm(15485760)
	if err != nil {
		m["error"] = "ОШИБКА разбора параметров1: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}

	//err = req.ParseForm()
	if err != nil {
		m["error"] = "ОШИБКА разбора параметров2: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}

	m["fam"] = req.PostFormValue("fam")
	m["name"] = req.PostFormValue("name")
	m["pat"] = req.PostFormValue("pat")
	m["email"] = req.PostFormValue("email")
	m["phone"] = req.PostFormValue("phone")
	m["street"] = req.PostFormValue("street")
	m["house"] = req.PostFormValue("house")
	m["flat"] = req.PostFormValue("flat")
	m["posttext"] = req.PostFormValue("posttext")

	log.Println(m["fam"])

	m["info"] = interface{}(string("ваше заявление успешно отправлено, информация о рассмотрении прийдет вам на " + m["email"].(string)))

	retstr := mf.ToJsonStr(m)
	log.Println(retstr)
	********/
}

// проверяем наличие и если надо регистрируем нового пользователя в бд
func check_and_register_user_in_sess_and_db(js map[string]interface{}, session sessions.Session) {
	user := GetSessJson(session, "user", "{}")
	if _, ok := user["uuid"]; !ok { //если пользователь не авторизован
		register_new_user_in_sess_and_db(js, session)
	} else if sess_email, ok := user["email"]; !ok || sess_email.(string) != js["email"].(string) {
		SetSessJson(session, "user", map[string]interface{}{})
		register_new_user_in_sess_and_db(js, session)
	} else {
		update_user_in_sess_and_db(js, session)
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

	var n int
	{
		query := "SELECT COUNT(*) FROM tuser WHERE email=?"
		stmt, err := db.Prepare(query)
		LogPrintErrAndExit("ERROR db.Prepare: \n"+query+"\n\n", err)
		email := u["email"].(string)
		err = stmt.QueryRow(email).Scan(&n)
		LogPrintErrAndExit("ERROR stmt.QueryRow(email).Scan(&n): \n"+query+"\n\n", err)
	}
	if n == 0 { //если такой email не существует, то создаем нового пользователя
		u["pass"] = genPassword(6, 10)
		u["regdate"] = mf.CurTimeStrShort()

		activecode := mf.StrUuid()

		query := "INSERT INTO tuser(uuid,type,fam,name,pat,email,phone,pass,street,house,flat,info,regdate,regdatet,isactive,activecode) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,0,?)"
		_, err := db.Exec(query, u["uuid"], 0, u["fam"], u["name"], u["pat"], u["email"], u["phone"], u["pass"], u["street"], u["house"], u["flat"], "{}", u["regdate"], activecode)
		LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)

		js["emailsend"] = interface{}(string("Для активации вашего аккаунта пройдите по ссылке /actusercode/" + activecode))

		SetSessJson(session, "user", u)
		return
	}
	update_user_in_sess_and_db(js, session)
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

func update_user_in_sess_and_db(js map[string]interface{}, session sessions.Session) {

}
