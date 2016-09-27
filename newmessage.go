package main

import (
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

	"github.com/nfnt/resize"
)

var maxImageSize = 1280
var minImageSize = 160

func newmessage(r render.Render, session sessions.Session) {
	var m = map[string]interface{}{"cnt": 0}
	m["uuid"] = mf.StrUuid()
	m["time"] = mf.CurTimeStrShort()
	r.HTML(200, "newmessage", m)
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
			//_, err = file.Seek(0, 0)
			//LogPrintErrAndExit("ERROR file.Seek1", err)

			//img, err := jpeg.Decode(file)
			//LogPrintErrAndExit("ERROR jpeg.Decode", err)
			m := resize.Resize(uint(iwidth), uint(iheight), img, resize.Lanczos3)

			out, err := os.Create(apppath + ifilepath_min)
			LogPrintErrAndExit("ERROR os.Create", err)
			defer out.Close()

			// write new image to file
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

type test_struct struct {
	Test string
}

func newmessagesend(req *http.Request /*, session sessions.Session*/) string {
	var m = map[string]interface{}{"cnt": 0}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		m["error"] = "ОШИБКА разбора параметров0: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}
	log.Println(string(body))
	return string(body)

	/********/
	err = req.ParseMultipartForm(15485760)
	if err != nil {
		m["error"] = "ОШИБКА разбора параметров1: " + mf.ErrStr(err)
		return mf.ToJsonStr(m)
	}
	/********/
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
	return retstr
}
