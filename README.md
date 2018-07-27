# bzmj
开发交流群：731507716
紧急公告：由于代码克隆过慢，即日起代码有src目录下代码，迁移到src工程下。
棋牌游戏服务器
1:开发过程讨论模块
1.1：游戏服务器断线重连大厅服务器机制
1.2：同时在线连接数量
1.3: 如果使用redis+mysql或者redis+mongodb，数据一致性怎么保证

2:数据库环境
uname:root
password:root
sudo apt-get install mysql-server
sudo apt-get install mysql-client
sudo apt install libmysqlclient-dev

mysql -uroot -p你的密码

现在设置mysql允许远程访问，首先编辑文件/etc/mysql/mysql.conf.d/mysqld.cnf：

sudo vi /etc/mysql/mysql.conf.d/mysqld.cnf

注释掉bind-address = 127.0.0.1：
保存退出，然后进入mysql服务，执行授权命令：

grant all on *.* to root@'%' identified by '你的密码' with grant option;

flush privileges;

然后执行quit命令退出mysql服务，执行如下命令重启mysql：

service mysql restart

现在在Windows下可以使用navicat远程连接Ubuntu下的MySQL服务：
CREATE DATABASE IF NOT EXISTS dbtest1 DEFAULT CHARSET utf8 COLLATE utf8_general_ci;
CREATE DATABASE IF NOT EXISTS dbtest2 DEFAULT CHARSET utf8 COLLATE utf8_general_ci;
CREATE DATABASE IF NOT EXISTS dpsglog DEFAULT CHARSET latin1;

*******************************************************************************
3:redis环境
在 Ubuntu 系统安装 Redi 可以使用以下命令:
$sudo apt-get update
$sudo apt-get install redis-server

启动 Redis
$ redis-server

查看 redis 是否启动？
$ redis-cli

以上命令将打开以下终端：
redis 127.0.0.1:6379>

127.0.0.1 是本机 IP ，6379 是 redis 服务端口。现在我们输入 PING 命令。
redis 127.0.0.1:6379> ping
PONG

以上说明我们已经成功安装了redis。

