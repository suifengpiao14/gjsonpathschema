package lineschema_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/lineschema"
	"github.com/tidwall/gjson"
)

func TestTransfer(t *testing.T) {
	lschemaRaw := `
	version=http://json-schema.org/draft-07/schema#,id=out,direction=out
	fullname=code,format=int,description=业务状态码,comment=业务状态码,example=0
	fullname=message,description=业务提示,comment=业务提示,example=ok
	fullname=items,type=array,description=数组,comment=数组,example=-
	fullname=items[].id,format=int,description=主键,comment=主键,example=0
	fullname=items[].title,description=广告标题,comment=广告标题,example=新年豪礼
	fullname=items[].windowIds[],format=int,description=窗口Id集合,comment=窗口Id集合,example=[1,23,4]
	fullname=items[].windowIds1[],type=int,format=int,description=窗口Id集合,comment=窗口Id集合,example=[1,23,4]
	fullname=items[].windowIds2,type=array,format=string,description=窗口Id集合,comment=窗口Id集合,example=[1,23,4]
	fullname=pagination,type=object,description=对象,comment=对象
	fullname=pagination.index,format=int,description=页索引,0开始,comment=页索引,0开始,example=0
	fullname=pagination.size,format=int,description=每页数量,comment=每页数量,example=10
	fullname=pagination.total,format=int,description=总数,comment=总数,example=60
	`
	lschema, err := lineschema.ParseLineschema(lschemaRaw)
	require.NoError(t, err)
	//input := `{"code":200,"message":"ok","items":[{"id":1,"title":"test1","windowIds":[1,2,3],"windowIds1":[1,2,3],"windowIds2":[1,2,3]},{"id":2,"title":"test2","windowIds":[4,5,6],"windowIds1":[4,5,6]},"windowIds2":[4,5,6]}],"pagination":{"index":0,"size":10,"total":100}}`
	//input := `{"code":"200","message":"ok"}`
	pathMap := lschema.TransferToFormat().String()
	excepted := `{code:code.@tonum,message:message.@tostring,items:{title:items.#.title.@tostring,windowIds:items.#.windowIds.#.@tostring,windowIds1:items.#.windowIds1.#.@tonum,windowIds2:items.#.windowIds2.#.@tostring,id:items.#.id.@tostring}|@group,pagination:{index:pagination.index.@tostring,size:pagination.size.@tostring,total:pagination.total.@tostring}}`
	//assert.Equal(t, excepted, pathMap)
	_ = excepted
	fmt.Println(pathMap)

}

type user struct {
	Name   string `json:"name"`
	UserId int    `json:"userId"`
}

func TestToGoTypeTransfer(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		lineSchema := lineschema.ToGoTypeTransfer(new(user)).String()
		expected := `{name:@this.name.@tostring,userId:@this.userId.@tonum}`
		assert.Equal(t, expected, lineSchema)
	})

	t.Run("slice[struct]", func(t *testing.T) {
		users := make([]user, 0)
		lineSchema := lineschema.ToGoTypeTransfer(users).String()
		expected := `{name:@this.#.name.@tostring,userId:@this.#.userId.@tonum}|@group`
		assert.Equal(t, expected, lineSchema)
	})
	t.Run("array[struct]", func(t *testing.T) {
		users := [2]user{}
		lineSchema := lineschema.ToGoTypeTransfer(users).String()
		expected := `{name:@this.#.name.@tostring,userId:@this.#.userId.@tonum}|@group`
		assert.Equal(t, expected, lineSchema)
	})

	t.Run("array[int]", func(t *testing.T) {
		ids := [2]string{}
		lineSchema := lineschema.ToGoTypeTransfer(ids).String()
		expected := `@this.#.@tostring`
		assert.Equal(t, expected, lineSchema)
		fmt.Println(lineSchema)
	})

	t.Run("int", func(t *testing.T) {
		id := 2
		lineSchema := lineschema.ToGoTypeTransfer(id).String()
		expected := `@this.@tonum`
		assert.Equal(t, expected, lineSchema)
		fmt.Println(lineSchema)
	})

}

func TestStructArrayPath(t *testing.T) {
	jsonStr := `[{"name":"张三","userId":"1"},{"name":"李四","userId":"2"}]`
	path := `[{name:@this.#.name.@tostring,userId:@this.#.userId.@tonum}|@group]`
	newJson := gjson.Get(jsonStr, path).String()
	fmt.Println(newJson)
}

func TestSimpleArrayPath(t *testing.T) {
	jsonStr := `[1,2,3]`
	path := `@this.#.@tostring`
	newJson := gjson.Get(jsonStr, path).String()
	fmt.Println(newJson)
}
func TestValuePath(t *testing.T) {
	jsonStr := `"1"`
	path := `@this.@tonum`
	newJson := gjson.Get(jsonStr, path).String()
	fmt.Println(newJson)
}
