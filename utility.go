package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//转MD5
func mymd5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

//获得指定长度随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//获得全局唯一时间戳,19位长
var uniqueTimeMutex sync.Mutex

func getUniqueTimeStamp() string {
	uniqueTimeMutex.Lock()
	time.Sleep(time.Microsecond)
	result := strconv.FormatInt(time.Now().UnixNano(), 10)
	uniqueTimeMutex.Unlock()
	return result
}

//成功页面
func successhtml(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("web/success.html")
	t.Execute(w, nil)
}

//失败页面
func errorhtml(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("web/error.html")
	t.Execute(w, nil)
}

//时间格式转换(Goods_info)
func timeTrans_gi(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

//收货地址转换
func address_show(add string) string {
	if add == "" {
		return "地址未设置"
	} else {
		return add
	}
}

//付款信息转换
func pay_show(pay int) string {
	if pay == 0 {
		return "未付款"
	} else if pay == 1 {
		return "已付款"
	} else {
		return "付款信息异常"
	}
}

//快递信息转换
func express_show(exp int) string {
	if exp == 0 {
		return "未发货"
	} else if exp == 1 {
		return "已发货"
	} else if exp == 2 {
		return "已收获"
	} else {
		return "付款信息异常"
	}
}

//功能测试
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "here is test")
}

//避免渲染html转义
func unescaped(x string) interface{} {
	return template.HTML(x)
}
