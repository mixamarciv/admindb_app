package main

//отладка
var gcfg_debug = map[string]int{"render_template": 1}

//глобальные переменные на весь проект
var gcfg_db_pass = "masterkey"
var gcfg_webserver_port = "8091"

var gcfg_secret_email_pass string = "AsPeefW2m42i03yqVB9f123"

var gcfg_secret_cookie_name = "s"
var gcfg_secret_cookie_key = "qwer1234"

//емеил на который будут уходить сообщения с сайта
var gcfg_work_emails = []string{"mixamarciv@gmail.com"}

var gcfg_default_session_data = map[string]interface{}{"style": "dark"}

//количество сообщений на одной странице
var gcfg_cnt_messages_on_page = 8
