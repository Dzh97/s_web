package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-session/session"
)

type Ret_orders struct {
	Count uint64
	Items []*Order_info
}

//查询用户的销售订单
func sellorder(w http.ResponseWriter, r *http.Request) {
	store, _ := session.Start(context.Background(), w, r)
	username, _ := store.Get("username")

	order, count, err := sellorder_db(username.(string), 0)
	if err != nil {
		fmt.Println(err)
	}

	ret := Ret_orders{Items: order, Count: count}

	funcmap := template.FuncMap{
		"timeTrans_gi": timeTrans_gi,
		"address_show": address_show,
		"pay_show":     pay_show,
		"express_show": express_show}
	t, err := template.New("sellorder.html").Funcs(funcmap).ParseFiles("web/sellorder.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, ret)
}

//查询用户的购买订单
func buyorder(w http.ResponseWriter, r *http.Request) {
	store, _ := session.Start(context.Background(), w, r)
	username, _ := store.Get("username")

	order, count, _ := buyorder_db(username.(string), 0)

	ret := Ret_orders{Items: order, Count: count}

	funcmap := template.FuncMap{
		"timeTrans_gi": timeTrans_gi,
		"address_show": address_show,
		"pay_show":     pay_show,
		"express_show": express_show}
	t, err := template.New("buyorder.html").Funcs(funcmap).ParseFiles("web/buyorder.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, ret)
}

//填写/修改地址信息
func setaddress(w http.ResponseWriter, r *http.Request) {
	store, _ := session.Start(context.Background(), w, r)
	username, _ := store.Get("username")

	r.ParseForm()
	id := r.FormValue("id")
	address := r.FormValue("address")

	setaddress_db(id, address, username.(string))
}
