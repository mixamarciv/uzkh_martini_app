CALL "%~dp0/set_path.bat"



::@CLS
::@pause

@echo === install ===================================================================

go get -u "github.com/go-martini/martini"
go get -u "github.com/martini-contrib/auth"
go get -u "github.com/martini-contrib/binding"
go get -u "github.com/martini-contrib/sessions"
go get -u "github.com/codegangsta/martini-contrib/render"

go install

@echo ==== end ======================================================================
@PAUSE
