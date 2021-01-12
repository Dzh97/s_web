package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-session/session"
	"github.com/streadway/amqp"
)

//渲染上架商品页面(get)和实现上架商品功能(post)
func sell(w http.ResponseWriter, r *http.Request) {

	//session中获取用户名
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	username, ok := store.Get("username")

	//如果失败，则回到登陆界面
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method == "GET" {
		t, _ := template.ParseFiles("web/sell.html")
		t.Execute(w, nil)
	}

	if r.Method == "POST" {
		// 接受文件
		file, _, _ := r.FormFile("file")
		defer file.Close()

		//判断文件类型,如果不是JPG或PNG则失败
		file_t, _, _ := r.FormFile("file")
		defer file_t.Close()
		ispic := ispicture(file_t)
		if ispic == 0 {
			fmt.Fprintf(w, "图片格式错误")
			return
		}

		// 将文件拷贝到指定路径下
		picture := "pic/" + strconv.FormatInt(time.Now().UnixNano(), 10)
		if ispic == 1 {
			picture += ".jpg"
		} else if ispic == 2 {
			picture += ".png"
		}
		dst, _ := os.Create("./web/" + picture)
		io.Copy(dst, file)
		dst.Close()
		go repairfile(picture, ispic)

		r.ParseForm()
		name := r.Form["name"][0]
		time_b := []byte(r.Form["time"][0])
		price_s := r.Form["price"][0]
		number_s := r.Form["number"][0]
		introduce := r.Form["introduce"][0]

		price, _ := strconv.Atoi(price_s)
		number, _ := strconv.Atoi(number_s)
		time_b[10] = ' '
		time_t, _ := time.ParseInLocation("2006-01-02 15:04:05", string(time_b)+":00", time.Local)

		//看看参数

		fmt.Printf("%T:%v\n", username, username)
		fmt.Printf("%T:%v\n", name, name)
		fmt.Printf("%T:%v\n", time_t, time_t)
		fmt.Printf("%T:%v\n", price, price)
		fmt.Printf("%T:%v\n", number, number)
		fmt.Printf("%T:%v\n", introduce, introduce)

		//参数合法性检查
		if price <= 0 { //price错误
			fmt.Fprintf(w, "price错误")
			return
		}
		if number <= 0 { //number错误
			fmt.Fprintf(w, "number错误")
			return
		}

		//存入数据库
		if false == sell_db(username.(string), name, picture, introduce, price, number, time_t) {
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		} else {
			http.Redirect(w, r, "/success", http.StatusFound)
			return
		}
	}
}

//判断文件类型为jpg或png
func ispicture(file multipart.File) int {
	buf := make([]byte, 10)
	n, _ := file.Read(buf)
	encodedStr := hex.EncodeToString(buf[:n])
	if encodedStr == "ffd8ffe000104a464946" { //jpg
		return 1
	} else if encodedStr == "89504e470d0a1a0a0000" { //png
		return 2
	}
	return 0
}

//补回文件类型信息
func repairfile(pos string, filetype int) {
	file, _ := os.Open("web/" + pos)
	defer file.Close()
	if filetype == 1 {
		file.WriteAt([]byte("ffd8ffe000104a464946"), 0)
	} else if filetype == 2 {
		file.WriteAt([]byte("89504e470d0a1a0a0000"), 0)
	}
}

//查询商品结构
type Ret_listgoods struct {
	Count uint64
	Items []*Goods_info
}

//查询20个数量不为0的商品
func listgoods(w http.ResponseWriter, r *http.Request) {
	goods, count, _ := listgoods_db(0)

	ret := Ret_listgoods{Items: goods, Count: count}

	funcmap := template.FuncMap{"timeTrans_gi": timeTrans_gi}
	t, err := template.New("listgoods.html").Funcs(funcmap).ParseFiles("web/listgoods.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, ret)
}

//商品详细信息
func goodsinfo(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	id := r.FormValue("id")

	goods, _ := goodsbyid_db(id)

	funcmap := template.FuncMap{"timeTrans_gi": timeTrans_gi}
	t, err := template.New("goodsinfo.html").Funcs(funcmap).ParseFiles("web/goodsinfo.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, goods)
}

//购买请求在消息队列中的结构
type Buy_info struct {
	Buyer string //购买人
	Id    string //商品ID
	Time  int64  //购买时间戳
}

//发送购买商品信息，添加到消息队列中
func sentbuyinfo(w http.ResponseWriter, r *http.Request) {

	now := time.Now().UnixNano()

	store, _ := session.Start(context.Background(), w, r)

	//从session中查找上次buy，若距离当前不超过0.1s，则直接失败
	lastbuy, ok := store.Get("lastbuy")
	if ok && now-lastbuy.(int64) < 100000000 {
		fmt.Fprintf(w, "你有问题，购买失败")
		return
	} else {
		store.Set("lastbuy", now)
		store.Save()
	}

	//session中获取用户名
	username, _ := store.Get("username")

	r.ParseForm()
	id := r.FormValue("id")

	//构造消息队列信息
	buy_info := Buy_info{
		Buyer: username.(string),
		Id:    id,
		Time:  now}
	js, _ := json.Marshal(buy_info)

	fmt.Println(string(js))

	//连接rabbitmq

	conn, _ := amqp.Dial(mqpos)
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	q, _ := ch.QueueDeclare(
		"buy_infos", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	//发送
	ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        js,
		},
	)

	fmt.Fprintf(w, "申请成功！购买成功将生成购买订单")
}

//从消息队列中获得购买信息，进行处理
func buyinfoprocess() {
	//连接rabbitmq
	conn, _ := amqp.Dial("amqp://root:rw@120.27.241.209:5672")
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()
	q, _ := ch.QueueDeclare(
		"buy_infos", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	//接收
	msgs, _ := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			buy_info := &Buy_info{}
			json.Unmarshal(d.Body, &buy_info)
			result := buyinfo_db(buy_info)
			fmt.Println(result)
		}
	}()

	<-forever
}
