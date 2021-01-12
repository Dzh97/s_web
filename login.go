package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-session/session"
)

//登陆功能，如果存在session则自动登陆，否则需要渲染登陆页面
func login(w http.ResponseWriter, r *http.Request) {

	//如果是GET方式
	if r.Method == "GET" {

		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		_, ok := store.Get("username")

		//session存在，自动登陆
		if ok {
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		} else { //不存在，需要登陆
			//渲染页面
			t, err := template.ParseFiles("web/login.html")
			if err != nil {
				fmt.Println(err)
			}
			t.Execute(w, nil)
		}
	}

	if r.Method == "POST" { //登陆

		//获取登录信息
		r.ParseForm()
		un := r.Form["username"][0]
		pw := r.Form["password"][0]
		pw = mymd5(pw)

		//查询数据库登陆
		status := login_db(w, r, un, pw)
		if status == true {
			http.Redirect(w, r, "/home", http.StatusFound)
		} else {
			fmt.Fprintf(w, "登录失败")
			return
		}
	}
}

//注册
func register(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, err := template.ParseFiles("web/register.html")
		if err != nil {
			fmt.Println(err)
		}
		t.Execute(w, nil)
	}

	if r.Method == "POST" {

		//获取登录信息
		r.ParseForm()
		un := r.Form["username"][0]
		pw := r.Form["password"][0]
		pw = mymd5(pw)

		//将注册信息写入数据库
		status := register_db(un, pw)
		if status == false {
			fmt.Fprintf(w, "注册失败")
		} else {
			//将登录信息存入session
			store, err := session.Start(context.Background(), w, r)
			if err != nil {
				fmt.Println(err)
				return
			}

			store.Set("username", un)
			err = store.Save()
			if err != nil {
				fmt.Println(err)
				return
			}

			//转到home页面
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session.Destroy(context.Background(), w, r)
	//转到login页面
	http.Redirect(w, r, "/login", http.StatusFound)
}
