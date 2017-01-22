package main

import (
	//"log"
	"net/http"
	"time"

	//"github.com/gorilla/context"
	"github.com/gorilla/mux"

	//"github.com/gorilla/sessions"
	"io"
	"os"
)

func init() {
	InitApp()
	InitLog()
	InitDb()
	InitAppCfg()
	//InitSendMail()
	InitMinify()
}

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	r.HandleFunc("/favicon.ico", LogReq(http_static_favicon_ico))

	//r.HandleFunc("/main", http_sess_handler)
	r.HandleFunc("/", LogReq(http_main))
	r.HandleFunc("/main", LogReq(http_main))

	r.HandleFunc("/s", LogReq(http_search))
	r.HandleFunc("/p", LogReq(http_post_view))

	r.HandleFunc("/e", LogReq(http_post_edit))
	r.HandleFunc("/e_ajax", LogReq(http_post_edit_ajax))

	r.HandleFunc("/login", LogReq(http_login))
	r.HandleFunc("/logout", LogReq(http_logout))
	r.HandleFunc("/auth_vk", LogReq(http_auth_vk))

	//вывод
	r.NotFoundHandler = MakeHttpHandler(LogReq(http_404))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + gcfg_webserver_port,
		WriteTimeout: 400 * time.Second,
		ReadTimeout:  400 * time.Second,
	}

	LogPrint("start listening port: " + gcfg_webserver_port)
	err := srv.ListenAndServe()
	LogPrintErrAndExit("ERROR start listening port: "+gcfg_webserver_port+" ", err)
}

func http_static_favicon_ico(w http.ResponseWriter, r *http.Request) {
	filename := apppath + "/public/favicon.ico"
	f, err := os.OpenFile(filename, os.O_RDONLY, 0000)
	if err != nil {
		ShowError("http_static_favicon_ico: OpenFile error", err, w, r)
		return
	}
	io.Copy(w, f)
}

func http_main(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	RenderTemplate(w, r, d, "maintemplate.html", "main.html")
	//w.Write([]byte("привет ворлд"))
	//context.Get(r, "nextfunc").(func(http.ResponseWriter, *http.Request))(w, r)
}

func http_404(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("404\npage not found\n\n\n"))
	d := map[string]interface{}{}
	RenderTemplate(w, r, d, "maintemplate.html", "error404.html")
}

func user_check_access(w http.ResponseWriter, r *http.Request, d map[string]interface{}) error {
	return nil
}
