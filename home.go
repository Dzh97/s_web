package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-session/session"
)

type Login_info struct {
	Username string
	Admin    string
}

//渲染home页面
func home(w http.ResponseWriter, r *http.Request) {

	//session中获取用户名
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	username, ok := store.Get("username")

	if ok {

		logininfo := Login_info{Username: username.(string)}

		isadmin := islegal_db("admin_home", w, r)
		if isadmin {
			logininfo.Admin = `<a href="admin_home"><button type="button" class="btn btn-default">后台管理</button></a>`
		}

		funcmap := template.FuncMap{
			"unescaped": unescaped,
		}
		t, err := template.New("home.html").Funcs(funcmap).ParseFiles("web/home.html")
		if err != nil {
			fmt.Println(err)
		}
		t.Execute(w, logininfo)
	} else { //若失败则未登录，跳转到登陆页面
		//转到login页面
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
}
