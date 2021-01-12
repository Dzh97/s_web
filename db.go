package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-session/session"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//用户信息表
type User_info struct {
	Username string `gorm:"username"` //用户名
	Password string `gorm:"password"` //密码MD5
}

//操作记录表
type Log_info struct {
	Username string    `gorm:"username"` //操作用户
	Time     time.Time `gorm:"time"`     //操作时间
	Action   string    `gorm:"action"`   //操作
}

//商品信息表
type Goods_info struct {
	Id        string    `gorm:"id" json:"id"`               //商品id
	Seller    string    `gorm:"seller" json:"seller"`       //卖家
	Price     int       `gorm:"price" json:"price"`         //售价
	Name      string    `gorm:"name" json:"name"`           //商品名
	Starttime time.Time `gorm:"starttime" json:"starttime"` //开卖时间
	Number    int       `gorm:"number" json:"number"`       //数量
	Picture   string    `gorm:"picture" json:"picture"`     //图片
	Introduce string    `gorm:"introduce" json:"introduce"` //文字介绍
}

//订单信息表
type Order_info struct {
	Id       string    `gorm:"id"`       //订单ID
	Name     string    `gorm:"name"`     //商品名称
	Goods_id string    `gorm:"goods_id"` //商品ID
	Seller   string    `gorm:"seller"`   //卖家
	Buyer    string    `gorm:"buyer"`    //买家
	Address  string    `gorm:"address"`  //地址信息
	Express  int       `gorm:"express"`  //快递信息,0未发货,1已发货,2已收货
	Pay      int       `gorm:"pay"`      //付款信息,0未付款,1已付款
	Time     time.Time `gorm:"time"`     //下单时间
}

//用户角色表
type User_role struct {
	Username string `gorm:"username"` //用户名
	Role     string `gorm:"role"`     //角色
}

//角色权限表
type Role_ac struct {
	Role        string `gorm:"role"`        //角色
	Access      string `gorm:"access"`      //权限
	Description string `gorm:"description"` //权限描述
}

//从offset开始查询20个log信息
func listlog_db(offset int) ([]*Log_info, uint64, error) {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	log := make([]*Log_info, 0)
	var count uint64

	if err := db.Offset(offset).Limit(20).Order("time desc").Find(&log).Count(&count).Error; err != nil {
		return log, count, err
	}
	return log, count, nil
}

//根据用户名查询role
func findrolebyname_db(name string) string {
	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	var user_role User_role
	if err := db.Where("username = ? ", name).Find(&user_role).Error; err != nil {
		fmt.Println(user_role)
		fmt.Println(err)
		return ""
	}
	return user_role.Role
}

//从offset开始查询20个用户信息
func listuser_db(offset int) ([]*User_info, uint64, error) {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	user := make([]*User_info, 0)
	var count uint64

	if err := db.Offset(offset).Limit(20).Find(&user).Count(&count).Error; err != nil {
		return user, count, err
	}
	return user, count, nil
}

//根据用户名和行为进行权限判断
func islegal_db(ac string, w http.ResponseWriter, r *http.Request) bool {
	store, _ := session.Start(context.Background(), w, r)
	username, ok := store.Get("username")
	if !ok {
		return false
	}

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	var user_role User_role
	if err := db.Where("username = ?", username.(string)).First(&user_role).Error; err != nil {
		fmt.Println("根据username从user_role表中查找失败,username = ", username)
		fmt.Println(err)
		return false
	}
	if user_role.Username != username.(string) {
		return false
	}

	var role_ac Role_ac
	if err := db.Where("role = ? and access = ?", user_role.Role, ac).First(&role_ac).Error; err != nil {
		fmt.Println("根据role和ac从role_ac表中查找失败, role =", user_role.Role, " |ac = ", ac)
		fmt.Println(err)
		return false
	}
	if role_ac.Access != ac {
		return false
	}
	return true

}

//修改订单地址
func setaddress_db(id, address, username string) bool {
	order, err := orderbyid_db(id)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if order.Buyer != username {
		return false
	}

	order.Address = address

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	if err = db.Save(&order).Error; err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

//根据ID获取订单信息
func orderbyid_db(id string) (*Order_info, error) {
	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	var order Order_info
	if err := db.Where("id = ?", id).First(&order).Error; err != nil {
		return &order, err
	}
	return &order, nil
}

//订单处理
func buyinfo_db(buy_info *Buy_info) int {
	//查询商品信息
	goods_info, err := goodsbyid_db(buy_info.Id)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	//合法性检查
	if goods_info.Number == 0 {
		return 2
	}
	st := goods_info.Starttime.UnixNano()
	if st > buy_info.Time {
		return 3
	}

	//事务：商品信息更新，订单生成
	goods_info.Number--
	order := Order_info{
		Id:       getUniqueTimeStamp(),
		Name:     goods_info.Name,
		Goods_id: goods_info.Id,
		Seller:   goods_info.Seller,
		Buyer:    buy_info.Buyer,
		Address:  "",
		Express:  0,
		Pay:      0,
		Time:     time.Now()}

	sw := db.Begin()
	if err = db.Save(&goods_info).Error; err != nil {
		fmt.Println(err)
		sw.Rollback()
		return 4
	}
	if err = db.Save(&order).Error; err != nil {
		fmt.Println(err)
		sw.Rollback()
		return 5
	}
	sw.Commit()

	//更新log_info表信息
	log := Log_info{Username: buy_info.Buyer, Time: time.Now(), Action: "buy " + buy_info.Id}
	err = db.Save(&log).Error
	if err != nil {
		fmt.Println(err)
	}

	return 0
}

//根据商品id查询商品信息
func goodsbyid_db(id string) (*Goods_info, error) {
	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	var goods Goods_info
	if err := db.Where("id = ?", id).First(&goods).Error; err != nil {
		return &goods, err
	}
	return &goods, nil
}

//从offset开始查询20个商品信息
func listgoods_db(offset int) ([]*Goods_info, uint64, error) {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	goods := make([]*Goods_info, 0)
	var count uint64

	if err := db.Where("number > ?", "0").Offset(offset).Limit(20).Find(&goods).Count(&count).Error; err != nil {
		return goods, count, err
	}
	return goods, count, nil
}

//从offset开始查询20个指定用户为seller的订单信息
func sellorder_db(username string, offset int) ([]*Order_info, uint64, error) {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	order := make([]*Order_info, 0)
	var count uint64

	if err := db.Where("seller = ?", username).Offset(offset).Limit(20).Find(&order).Count(&count).Error; err != nil {
		return order, count, err
	}
	return order, count, nil
}

//从offset开始查询20个指定用户为buyer的订单信息
func buyorder_db(username string, offset int) ([]*Order_info, uint64, error) {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	order := make([]*Order_info, 0)
	var count uint64

	if err := db.Where("buyer = ?", username).Offset(offset).Limit(20).Find(&order).Count(&count).Error; err != nil {
		return order, count, err
	}
	return order, count, nil
}

//将商品信息存入数据库
func sell_db(seller, name, picture, introduce string, price, number int, starttime time.Time) bool {

	//获取唯一id号
	id := getUniqueTimeStamp()

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	goods := Goods_info{Seller: seller,
		Price:     price,
		Name:      name,
		Starttime: starttime,
		Number:    number,
		Picture:   picture,
		Introduce: introduce,
		Id:        id}

	err := db.Save(&goods).Error
	if err != nil {
		fmt.Println(err)
		return false
	}

	//更新log_info表信息
	log := Log_info{Username: seller, Time: time.Now(), Action: "sell " + id}
	err = db.Save(&log).Error
	if err != nil {
		fmt.Println(err)
	}

	return true
}

//根据用户名和密码登陆，返回true登陆成功，返回false登录失败
func login_db(w http.ResponseWriter, r *http.Request, un string, pw string) bool {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	var user User_info
	db.Where("username = ? AND password = ?", un, pw).Find(&user)

	if user.Username == un { //登陆成功

		//将username存入session
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Println(err)
			return false
		}

		username := user.Username

		store.Set("username", username)
		err = store.Save()
		if err != nil {
			fmt.Println(err)
			return false
		}

		//更新log_info表信息
		log := Log_info{Username: user.Username, Time: time.Now(), Action: "login"}
		err = db.Save(&log).Error
		if err != nil {
			fmt.Println(err)
		}

		return true
	}
	//登录失败
	return false
}

//将注册信息写入数据库
func register_db(un, pw string) bool {

	//连接数据库
	db, _ := gorm.Open("mysql", dsn)
	defer db.Close()
	db.SingularTable(true)

	user := User_info{Username: un, Password: pw}

	//将注册信息写入数据库
	err := db.Save(&user).Error

	//若错误则输出
	if err != nil {
		fmt.Println(err)
		return false
	} else { //更新log_info表信息
		log := Log_info{Username: un, Time: time.Now(), Action: "register"}
		err = db.Save(&log).Error
		if err != nil {
			fmt.Println(err)
		}
		return true
	}
}
