package main

import (
	"html"
	"strconv"
	"time"

	"github.com/go-gomail/gomail"
	mf "github.com/mixamarciv/gofncstd3000"
)

var opts_sendmail map[string]string

func initSendMail() {
	opts_sendmail = make(map[string]string, 1)
	opts_sendmail["from"] = "uzkhinta@gmail.com"
	opts_sendmail["host"] = "smtp.gmail.com"
	opts_sendmail["port"] = "25"

	opts_sendmail["host"] = "smtp.gmail.com"
	opts_sendmail["port"] = "465"

	opts_sendmail["login"] = "uzkhinta"
	opts_sendmail["pass"] = secret_email_pass
	opts_sendmail["bodytype"] = "text/html"
}

func SendMail(mailto, subject, msg string, files []string) {
	m := gomail.NewMessage()
	m.SetHeader("From", opts_sendmail["from"])
	m.SetHeader("To", mailto)
	m.SetHeader("Subject", subject)
	m.SetBody(opts_sendmail["bodytype"], msg)

	for i := 0; i < len(files); i++ {
		m.Attach(files[i])
	}

	smtpport, _ := strconv.Atoi(opts_sendmail["port"])

	d := gomail.NewPlainDialer(
		//"smtp.gmail.com", 25, "", "")
		opts_sendmail["host"], smtpport, opts_sendmail["login"], opts_sendmail["pass"])

	err := d.DialAndSend(m)

	LogPrintErrAndExit("ОШИБКА отправки сообщения: ", err)
	LogPrint("письмо успешно отправлено")
}

//отправляем уведомления о новых сообщениях на рабочий емэйл адрес
func SendMailNewPostsToWork() {
	go func() {
		time.Sleep(3000 * time.Millisecond)

		posts := make([]map[string]interface{}, 0)
		{ //загружаем список сообщений которые ещё не отправляли:
			query := "SELECT uuid,type,userdata,text,postdatet,"
			query += "(SELECT COUNT(*) FROM timage t WHERE t.uuid_post=p.uuid) "
			query += " FROM tpost p WHERE isactive=1 AND isstartwork=0 ORDER BY postdatet"
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
				m := map[string]interface{}{"uuid": uuid.get("-"), "ptype": ptype.get("-")}
				m["userdata"] = mf.FromJsonStr([]byte(userdata.get("{}")))
				m["text"] = html.EscapeString(text.get("-"))
				m["text"] = mf.StrRegexpReplace(m["text"].(string), "\\n", "<br>")
				m["postdatet"] = postdatet
				m["postdatefmt"] = postdatet.Format("02.01.2006 15:04")
				m["postdates"] = postdates
				m["imgcnt"] = imgcnt
				if imgcnt > 0 {
					m["images"] = load_posts_or_comment_images_arr(m["uuid"].(string), "post", 0)
				}
				posts = append(posts, m)
			}
			rows.Close()
		}

		apppath, _ := mf.AppPath()
		for _, m := range posts {
			u := m["userdata"].(map[string]interface{})
			sbj := "сообщение на " + sitedomain + " от " + (m["postdatefmt"]).(string) + " разместил " + (u["name"]).(string) + " " + (u["pat"]).(string) + " " + (u["fam"]).(string)

			siteurl := "http://" + sitedomain + "/messagelist"
			msg := (m["text"]).(string) + "<br>\n<br>\n"
			msg += "----<br>\n"
			msg += "посмотреть на сайте <a href=\"" + siteurl + "\">" + siteurl + "</a> <br>\n"
			msg += "сообщение загружено " + (m["postdatefmt"]).(string) + "<br>\n"
			msg += "от имени " + (u["fam"]).(string) + " " + (u["name"]).(string) + " " + (u["pat"]).(string) + "<br>\n"
			msg += "email: " + (u["email"]).(string) + "<br>\n"
			msg += "тел.: " + (u["phone"]).(string) + "<br>\n"
			msg += "адрес: ул. " + (u["street"]).(string) + ", дом " + (u["house"]).(string) + ", кв. " + (u["flat"]).(string) + "<br>\n"

			files := make([]string, 0)
			if m["imgcnt"].(int) > 0 {
				//imgs := m["images"].{[]map[string]string}
				for _, img := range m["images"].([]map[string]string) {
					path := apppath + "/public" + img["pathmin"]
					files = append(files, path)
				}
			}

			for _, iemail := range work_emails {
				SendMail(iemail, sbj, msg, files)
			}

			{ //обновляем информацию о том что сообщение отправлено на рабочий емейл
				query := "UPDATE tpost SET isstartwork=1 WHERE uuid=? "
				_, err := db.Exec(query, m["uuid"])
				LogPrintErrAndExit("ERROR db.Exec: \n"+query+"\n\n", err)
			}
		}
	}()
}
