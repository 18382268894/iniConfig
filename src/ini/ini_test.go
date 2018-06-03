package ini

import (
	"testing"
	"fmt"
	"io/ioutil"
)


type server struct {
	MysqlServer `ini:"mysqlServer"`
	LoginServer `ini:"loginServer"`
}


type MysqlServer struct {
	IP string `ini:"ip"`
	Port int  `ini:"port"`
	UserName string `ini:"userName"`
	Passwd string `ini:"passwd"`
	Addtion string `ini:"addtion"`
}


type LoginServer struct {
	UserName string `ini:"userName"`
	Passwd string `ini:"passwd"`
	Email string `ini:"email"`
}



func TestIniConf(T *testing.T){

	data,err := ioutil.ReadFile("testIni.ini")
	if err != nil{
		fmt.Printf("cannot open the file,err:%v",err)
		return
	}
	var serv = new(server)
	err = Umarshal(data,serv)
	fmt.Printf("%#v",serv)
	if err != nil{
		fmt.Println(err)
		return
	}

}