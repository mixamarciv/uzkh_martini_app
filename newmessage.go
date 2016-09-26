package main

import (
	"image"
	"image/jpeg"
	"log"
	"mime/multipart"
	"os"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	"path/filepath"

	mf "github.com/mixamarciv/gofncstd3000"

	"github.com/nfnt/resize"
)

var maxImageSize = 1280
var minImageSize = 100

func newmessage(r render.Render, session sessions.Session) {
	var m = map[string]interface{}{"cnt": 0}
	m["uuid"] = mf.StrUuid()
	m["time"] = mf.CurTimeStrShort()
	r.HTML(200, "newmessage", m)
}

type UploadForm struct {
	Uuid string                `form:"uuid"`
	Time string                `form:"time"`
	Path string                `form:"path"`
	File *multipart.FileHeader `form:"file"`
}

func uploadfile(uf UploadForm) string {
	file, err := uf.File.Open()
	log.Printf("ERR1: %#v", err)
	log.Printf("uuid: %#v", uf.Uuid)
	log.Printf("time: %#v", uf.Time)
	log.Printf("Path: %#v", uf.Path)
	log.Printf("File: %#v", file)

	//-------------------------------
	t := uf.Time
	ipath := mf.AppPath2()
	ipath = ipath + "/public/upload/" + t[0:4] + "/" + t[4:6] + "/" + t[6:8] + "/" + t + "_" + uf.Uuid

	mf.MkdirAll(ipath)

	ifilename := mf.StrUuid()
	ifilepath := ipath + "/" + ifilename + filepath.Ext(uf.Path)
	ifilepath_min := ipath + "/" + ifilename + "_min" + filepath.Ext(uf.Path)
	//-------------------------------

	im, _, err := image.DecodeConfig(file)
	LogPrintErrAndExit("ERROR image.DecodeConfig", err)

	_, err = file.Seek(0, 0)
	LogPrintErrAndExit("ERROR file.Seek1", err)

	iwidth := im.Width
	iheight := im.Height
	if iwidth > iheight && iwidth > maxImageSize {
		iwidth = maxImageSize
		iheight = 0
	} else if iheight > iwidth && iheight > maxImageSize {
		iheight = maxImageSize
		iwidth = 0
	}

	img, err := jpeg.Decode(file)
	LogPrintErrAndExit("ERROR jpeg.Decode", err)

	m := resize.Resize(uint(iwidth), uint(iheight), img, resize.Lanczos3)

	out, err := os.Create(ifilepath)
	LogPrintErrAndExit("ERROR os.Create", err)

	jpeg.Encode(out, m, nil)
	defer out.Close()

	//-- min image ---------------
	iwidth = im.Width
	iheight = im.Height
	minimize := 0
	if iwidth > iheight && iwidth > minImageSize {
		iwidth = minImageSize
		iheight = 0
		minimize = 1
	} else if iheight > iwidth && iheight > minImageSize {
		iheight = minImageSize
		iwidth = 0
		minimize = 1
	}

	log.Printf("minimize: %#v", minimize)

	if minimize == 1 {
		//_, err = file.Seek(0, 0)
		//LogPrintErrAndExit("ERROR file.Seek1", err)

		//img, err := jpeg.Decode(file)
		//LogPrintErrAndExit("ERROR jpeg.Decode", err)
		m := resize.Resize(uint(iwidth), uint(iheight), img, resize.Lanczos3)

		out, err := os.Create(ifilepath_min)
		LogPrintErrAndExit("ERROR os.Create", err)
		defer out.Close()

		// write new image to file
		jpeg.Encode(out, m, nil)
	}

	/**************
	var d []byte = make([]byte, 1024*1024)
	size, err := file.Read(d)

	log.Printf("ERR2: %#v", err)
	log.Printf("size: %#v", size)

	t := uf.Time
	ipath := mf.AppPath2()
	log.Printf("ERR3: %#v", err)
	ipath = ipath + "/public/upload/" + t[0:4] + "/" + t[4:6] + "/" + t[6:8] + "/" + t + "_" + uf.Uuid

	mf.MkdirAll(ipath)

	ifilename := mf.StrUuid()
	ifilepath := ipath + "/" + ifilename + filepath.Ext(uf.Path)
	ifilepath_min := ipath + "/" + ifilename + "_min" + filepath.Ext(uf.Path)
	mf.FileWrite(ifilepath, d)

	//--- resize -------------------------------------------------------------
	{
		file, err := os.Open(ifilepath)
		if err != nil {
			log.Fatal(err)
		}
		img, err := jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		// resize to width 320 using Lanczos resampling and preserve aspect ratio
		m := resize.Resize(320, 0, img, resize.Lanczos3)

		out, err := os.Create(ifilepath_min)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		// write new image to file
		jpeg.Encode(out, m, nil)
	}
	//--- /resize ------------------------------------------------------------
	****************/
	return ipath
}
