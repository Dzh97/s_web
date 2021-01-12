# s_web介绍

用go语言写的一个简单的物品买卖系统。

# 配置说明：

1、配置好mysql数据库和RabbitMq消息队列。

2、在listene.go中修改数据库连接信息和RabbitMq连接信息。

3、默认本地访问地址为127.0.0.1:12345

#### 数据库表的创建如下：

create table user_info(

username varchar(20),

password varchar(50),

primary key(username));



create table goods_info(

id varchar(19),

seller varchar(20),

price int(11),

name varchar(50),

starttime datetime,

picture varchar(100),

introduce varchar(200),

primary key(id));



create table order_info(

id varchar(19),

name varchar(50),

goods_id varchar(19),

seller varchar(20),

buyer varchar(20),

address varchar(200),

express int(11),

pay int(11),

time datetime

primary key(id));



create table log_info(

username varchar(20),

time datetime,

action varchar(200));



create table user_role(

username varchar(20),

role varchar(20));



create table role_ac(

role varchar(20),

access varchar(50),

description varchar(50));





