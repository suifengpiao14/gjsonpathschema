package lineschema_test

import (
	"encoding/json"
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

	t.Run("complex", func(t *testing.T) {
		packschema := `version=http://json-schema.org/draft-07/schema#,id=out
fullname=code,format=int,required,title=业务状态码,default=0,comment=业务状态码,example=0
fullname=message,required,title=业务提示,default=ok,comment=业务提示,example=ok
fullname=items[].id,format=int,required,title=主键,comment=主键,example=1
fullname=items,type=array,title=-,comment=-
fullname=items[].name,required,title=项目标识,comment=项目标识,example=advertise
fullname=items[].title,required,title=名称,comment=名称
fullname=items[].config,required,title=项目curd配置内容,comment=项目curd配置内容
fullname=items[].createdAt,format=datetime,required,title=创建时间,comment=创建时间,example=2023-01-1200:00:00
fullname=items[].updatedAt,format=datetime,required,title=修改时间,comment=修改时间,example=2023-01-3000:00:00
fullname=pagination.index,format=int,required,title=页索引,0开始,default=0,comment=页索引,0开始,example=0
fullname=pagination.size,format=int,required,title=每页数量,default=10,comment=每页数量,example=10
fullname=pagination.total,format=int,required,title=总数,comment=总数,example=60`
		lschema, err := lineschema.ParseLineschema(packschema)
		require.NoError(t, err)
		gjsonPath := lschema.TransferToFormat().Reverse().String()
		//gjsonPath = `{code:code.@tostring,message:message.@tostring,items:{config:items.#.config.@tostring,createdAt:items.#.createdAt.@tostring,updatedAt:items.#.updatedAt.@tostring,id:items.#.id.@tostring,name:items.#.name.@tostring,title:items.#.title.@tostring}|@group,pagination:{index:pagination.index.@tostring,size:pagination.size.@tostring,total:pagination.total.@tostring}}`
		fmt.Println(gjsonPath)
		data := `{"code":0,"message":"","items":[{"id":2,"name":"advertise1","title":"广aa告","config":"{\"navs\":[1]}","createdAt":"","updatedAt":""}],"pagination":{"index":0,"size":10,"total":1}}`
		out := gjson.Get(data, gjsonPath).String()
		fmt.Println(out)
	})

	t.Run("complex2", func(t *testing.T) {
		packschema := `version=http://json-schema.org/draft-07/schema#,id=out
fullname=code,format=int,required,title=业务状态码,default=0,comment=业务状态码,example=0
fullname=message,required,title=业务提示,default=ok,comment=业务提示,example=ok
fullname=navs[].id,format=int,required,title=主键,comment=主键
fullname=navs[].name,required,title=名称,comment=名称
fullname=navs[].title,required,title=标题,comment=标题
fullname=navs[].route,required,title=路由,comment=路由
fullname=navs[].sort,format=int,required,title=排序,comment=排序`
		lschema, err := lineschema.ParseLineschema(packschema)
		require.NoError(t, err)
		gjsonPath := lschema.TransferToFormat().Reverse().String()
		data := `{"code":0,"message":"","navs":[{"id":1,"name":"creative","title":"广告创意","route":"creativeList","sort":99},{"id":2,"name":"plan","title":"广告计划","route":"planList","sort":98},{"id":3,"name":"window","title":"橱窗","route":"windowList","sort":97}]}`
		fmt.Println(gjsonPath)
		out := gjson.Get(data, gjsonPath).String()
		fmt.Println(out)

	})
	t.Run("object with no children", func(t *testing.T) {
		packschema := `version=http://json-schema.org/draft-07/schema#,id=out
fullname=code,format=int,required,title=业务状态码,default=0,comment=业务状态码,example=0
fullname=message,required,title=业务提示,default=ok,comment=业务提示,example=ok
fullname=uiSchema,type=object,required,title=uiSchema对象,comment=uiSchema对象`
		lschema, err := lineschema.ParseLineschema(packschema)
		require.NoError(t, err)
		gjsonPath := lschema.TransferToFormat().Reverse().String()
		data := `{"code":0,"message":"","uiSchema":""}`
		fmt.Println(gjsonPath)
		out := gjson.Get(data, gjsonPath).String()
		fmt.Println(out)

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

type UerNoJsonTag struct {
	Name      string
	ID        int
	CreatedAt string
	Update_at string
}

func TestJsonUmarsh(t *testing.T) {
	u := UerNoJsonTag{}
	data := `{"name":"张三","id":2,"createdAt":"2023-11-24 16:10:00","Update_at":"2023-11-24 16:10:00"}`
	json.Unmarshal([]byte(data), &u)
	b, _ := json.Marshal(u)
	s := string(b)
	fmt.Println(s)
}
