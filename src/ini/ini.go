package ini

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"strconv"
	"io/ioutil"
	"os"
)



func MarshalFile(fileName string,result interface{})(err error){
	f,err := os.OpenFile(fileName,os.O_CREATE |os.O_TRUNC |os.O_WRONLY,0766)
	if err != nil{
		err  = fmt.Errorf("cannot open the file,err:",err)
		return
	}
	data,err := Marshal(result)
	if err != nil{
		return
	}
	
	_,err = f.Write(data)
	if err != nil{
		err = fmt.Errorf("Write ini failed,err:",err)
		return 
	}
	return nil
}

func Marshal(result interface{})(data []byte,err error){
	var dataString = ""
	v := reflect.ValueOf(result)
	t := v.Type()
	if t.Kind() != reflect.Struct {
		err = fmt.Errorf("The container must be struct")
		return
	}
	for i:=0;i<v.NumField();i++ {

		if v.Field(i).Type().Kind() == reflect.Struct{
			writeSection(&dataString,v.Field(i),t.Field(i))

		}else{
			writeItem(&dataString,v.Field(i),t.Field(i))
		}

	}
	return []byte(dataString),nil

}

func writeSection(dataString *string, vfield reflect.Value, tfield reflect.StructField){
	sectionName := fmt.Sprintf("%v",tfield.Name)
	*dataString = *dataString + "[" + sectionName + "]" +"\n"
	for i:= 0;i<vfield.NumField();i++  {
		writeItem(dataString,vfield.Field(i),tfield.Type.Field(i))
	}
}

func writeItem(dataString *string,vfield reflect.Value, tfield reflect.StructField){
	tagStr := tfield.Tag.Get("ini")
	if tagStr == ""{
		tagStr = tfield.Name
	}
	val := fmt.Sprintf("%v",vfield)
	*dataString = *dataString + tagStr + "=" + val + "\n"
}

func UnMarshalFile(fileName string,result interface{})(err error){
	t := reflect.ValueOf(result).Type()
	if t.Kind() != reflect.Ptr {
		err = fmt.Errorf("The container must be address")
		return
	}
	if t.Elem().Kind() != reflect.Struct {
		err = fmt.Errorf("The container must be struct")
		return
	}

	data,err := ioutil.ReadFile(fileName)
	if err != nil{
		fmt.Errorf("Read data from file failed,err:%v",err)
		return
	}

	err = Unmarshal(data,result)
	if err != nil{
		fmt.Errorf("UnMarshalFile failed,err:%v",err)
		return
	}
	return
}

func Unmarshal(data []byte, result interface{}) (err error) {
	lineArr := strings.Split(string(data), "\n")
	val := reflect.ValueOf(result)
	t := val.Type()
	if t.Kind() != reflect.Ptr {
		err = fmt.Errorf("The container must be address")
		return err
	}

	if t.Elem().Kind() != reflect.Struct {
		err = fmt.Errorf("The container must be struct")
		return err
	}
	err = parseLines(lineArr,result)
	if err != nil{
		return err
	}

	return nil
}

func parseLines(lines []string,result interface{}) (err error) {
	v := reflect.ValueOf(result)
	structElem := v.Elem()
	lastSectionName := ""
	for index, lineStr := range lines {

		//空行直接忽略
		if len(lineStr) == 0 {
			continue
		}

		//处理以空格开始的行，如果有section和item则报错
		if lineStr[0] == ' ' {
			lineStr = strings.Trim(lineStr, " ")
			if lineStr[0] != '#' && lineStr[0] !=  ';' {
				err = fmt.Errorf("syntactic error,lineNo:%v", index+1) //section和item不能以空格开头
				return err
			}
			continue
		}

		//注释直接忽略
		if lineStr[0] == '#' || lineStr[0] == ';' {
			continue
		}

		// [ 开头的,通过parseSection来解析，如果错误会返回报错
		if lineStr[0] == '[' {
			err,lastSectionName = parseSection(lineStr)
			if err != nil{
				err = fmt.Errorf("%v,line%v:",err,index+1)
				return
			}
		} else {   //处理item
			if lastSectionName == "" {
				err = parseItem(lineStr,structElem)
				if err != nil{
					err = fmt.Errorf("%v,lineNo:%v",err,index+1)
					return
				}
			}else {     //section下的item处理

				var structStructElem reflect.Value
				ok := false
				//在结构体中查找section匹配的嵌套结构体
				for i:= 0;i < structElem.NumField();i++ {

					tagStr := structElem.Type().Field(i).Tag.Get("ini")
					if tagStr == lastSectionName{
						ok =true
						structStructElem = structElem.Field(i).Addr().Elem()  //获取嵌套结构的reflect.value
						break
					}

				}

				if ok == false{
					err = fmt.Errorf("No matched section,lineNO:%v",index+1)
				}

				err = parseItem(lineStr,structStructElem)
				if err != nil{
					err = fmt.Errorf("%v,lineNo:%v",err,index+1)
					return
				}
			}
			
		}

	}
	return nil
}

func parseSection(lineStr string) (err error, sectionName string) {
	if strings.IndexByte(lineStr, ']') == -1 || strings.IndexByte(lineStr, ']') != len(lineStr)-1 { //section必须包含[]，并且]只能在末尾
		err = errors.New("section syntactic error")
		return err, ""
	}
	sectionStr := lineStr[1:len(lineStr)-1]
	sectionName = strings.Trim(sectionStr, " ")  //获取[]中的内容，并且去掉空字符
	if len(sectionName) == 0 {
		err = errors.New("section context is empty")
		return err, ""
	}
	return nil, sectionName


}

func parseItem(lineStr string,structElem reflect.Value) (err error) {

	if strings.IndexByte(lineStr,'=') == -1{
		err = errors.New("item context must have \"=\"")
		return
	}

	//item不能直接以=开头
	if lineStr[0] == '=' {
		err = errors.New("item context is empty")
		return
	}
	keyAndVal := strings.Split(lineStr, "=")
	key := strings.Trim(keyAndVal[0]," ")
	val := strings.Trim(keyAndVal[1]," ")

	//忽略掉=号后面注释
	if val[0] == ';' || val[0] == '#'{
		val = ""
	}

	//在字段中查找tag
	ok := false
	var fieldElem reflect.Value
	for i:= 0;i < structElem.NumField();i++ {
		tagStr := structElem.Type().Field(i).Tag.Get("ini")
		if key == tagStr{
			ok =true
			fieldElem = structElem.Field(i)
			break
		}
	}
	if ok == false {

		err = errors.New("No mathed item")
		return
	}


	switch fieldElem.Type().Kind() {
	case reflect.String:
		fieldElem.SetString(val)
	case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
		num,_ := strconv.ParseInt(val,10,64)
		fieldElem.SetInt(num)
	case reflect.Float32,reflect.Float64:
		num,_ := strconv.ParseFloat(val,64)
		fieldElem.SetFloat(num)
	default:
		fieldElem.SetString(val)
	}


	return nil
}
