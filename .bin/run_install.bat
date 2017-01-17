CALL "%~dp0/set_path.bat"



::@CLS
::@pause

@echo === install ===================================================================
go get "github.com/gorilla/sessions"
go get "github.com/gorilla/mux"
go get "github.com/mixamarciv/gofncstd3000"
go get "github.com/go-gomail/gomail"
go get "github.com/nakagami/firebirdsql"
go get "github.com/nfnt/resize"
go get "github.com/tdewolff/minify"

go get "github.com/davecgh/go-spew/spew"

go install

@echo ==== end ======================================================================
@PAUSE
