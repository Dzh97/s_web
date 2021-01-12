package main

import (
	"fmt"
	"html/template"
	"net/http"
)

//渲染管理主页
func admin_home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("web/admin_home.html")
	t.Execute(w, nil)
}

//渲染用户角色信息
func admin_userrole(w http.ResponseWriter, r *http.Request) {
	user, count, _ := listuser_db(0)
	u_r := make([]User_role, count)
	for i, u := range user {
		u_r[i].Username = u.Username
		u_r[i].Role = findrolebyname_db(u.Username)
	}

	funcmap := template.FuncMap{
		"empty2user": empty2user,
	}
	t, _ := template.New("admin_userrole.html").Funcs(funcmap).ParseFiles("web/admin_userrole.html")
	t.Execute(w, u_r)
}

//若role为空则表示普通user
func empty2user(role string) string {
	if role == "" {
		return "user"
	} else {
		return role
	}
}

//渲染log信息
func admin_log(w http.ResponseWriter, r *http.Request) {
	log, _, _ := listlog_db(0)
	for _, l := range log {
		fmt.Println(l.Username, l.Action, l.Time)
	}

	funcmap := template.FuncMap{
		"timeTrans_gi": timeTrans_gi,
	}
	t, _ := template.New("admin_log.html").Funcs(funcmap).ParseFiles("web/admin_log.html")
	t.Execute(w, log)
}
