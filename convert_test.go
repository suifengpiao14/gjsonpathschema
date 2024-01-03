package lineschema_test

import (
	"fmt"
	"testing"

	"github.com/suifengpiao14/lineschema"
)

func TestJsonExample(t *testing.T) {

	l := NewLineSchema()
	l.JsonSchema()
	jsonStr, err := l.JsonExample()
	if err != nil {
		panic(err)
	}
	fmt.Println(jsonStr)
}

func NewLineSchema() (l *lineschema.Lineschema) {
	var jsonStr = `
		[{
			"config":{
				"id":"1",
				"status":"2",
				"identify":"abcde",
				"merchantId":"123",
				"merchantName":"测试商户",
				"operateName":"彭政",
				"storeId":"1",
				"storeName":"门店名称",
				"ids":["1","2"],
				"array":[
					{"id":"2","name":"ok"},
					{"id":"3","name":"ok"}
					]
			}
		}]
	`
	lineschema, err := lineschema.Json2lineSchema(jsonStr)
	if err != nil {
		panic(err)
	}
	return lineschema
}
