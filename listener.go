package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-session/session"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//数据库连接信息，"数据库名:数据库密码@连接方法(ip地址:端口)/数据库名?其他参数"
var dsn = "xxxxxxxx:xxxxxxxx@tcp(127.0.0.1:3306)/xxxxxxxx?charset=utf8mb4&parseTime=True&loc=Local"

//RabbitMq连接信息，"amqp://账号:密码@IP地址:5672"
var mqpos = "amqp://xxxxxxxx:xxxxxxxx@127.0.0.1:5672"

func main() {
	go buyinfoprocess() //购买信息处理

	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/register", register)
	http.HandleFunc("/home", home)
	http.HandleFunc("/sell", sell)
	http.HandleFunc("/success", successhtml)
	http.HandleFunc("/error", errorhtml)
	http.Handle("/listgoods", islogin(http.HandlerFunc(listgoods)))
	http.Handle("/goodsinfo", islogin(http.HandlerFunc(goodsinfo)))
	http.Handle("/sentbuyinfo", islogin(http.HandlerFunc(sentbuyinfo)))
	http.Handle("/sellorder", islogin(http.HandlerFunc(sellorder)))
	http.Handle("/buyorder", islogin(http.HandlerFunc(buyorder)))
	http.Handle("/setaddress", islogin(http.HandlerFunc(setaddress)))
	http.Handle("/admin_home", isadmin(http.HandlerFunc(admin_home)))
	http.Handle("/admin_userrole", isadmin(http.HandlerFunc(admin_userrole)))
	http.Handle("/admin_log", isadmin(http.HandlerFunc(admin_log)))
	http.HandleFunc("/test", home)

	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img"))))
	http.Handle("/pic/", http.StripPrefix("/pic/", http.FileServer(http.Dir("web/pic"))))
	http.Handle("/js-css/", http.StripPrefix("/js-css/", http.FileServer(http.Dir("web/js-css"))))

	port := "12345"
	fmt.Println("listen in port : ", port)
	http.ListenAndServe(":"+port, nil)
}

//登陆判断
func islogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//session中获取用户名
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		//当前仅进行登陆与否的权限控制
		_, ok := store.Get("username")

		if ok {
			next.ServeHTTP(w, r)
		} else {
			login(w, r)
		}
	})
}

//后台登陆权限验证
func isadmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if result := islegal_db("admin_home", w, r); result == false {
			errorhtml(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
