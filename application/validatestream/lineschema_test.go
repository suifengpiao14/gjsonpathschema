package validatestream_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/lineschema"
	"github.com/suifengpiao14/lineschema/application/validatestream"
	"github.com/tidwall/gjson"
)

func TestConvert(t *testing.T) {
	lschemaRaw := `
	version=http://json-schema.org/draft-07/schema#,id=out,direction=out
	fullname=code,description=业务状态码,comment=业务状态码,example=0
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
	input := `{"code":200,"message":"ok","items":[{"id":1,"title":"test1","windowIds":[1,2,3],"windowIds1":[1,2,3],"windowIds2":[1,2,3]},{"id":2,"title":"test2","windowIds":[4,5,6],"windowIds1":[4,5,6]},"windowIds2":[4,5,6]}],"pagination":{"index":0,"size":10,"total":100}}`
	//input := `{"code":"200","message":"ok"}`
	pathMap := lschema.TransferToTypeGjsonPath()
	fmt.Println(pathMap)
	out := validatestream.ConvertFomat([]byte(input), pathMap)
	fmt.Println(string(out))
}

/*
*
{pageInfo:{pageIndex:input.pageIndex,pageSize:input.pageSize,total:PaginateTotalOut},items:{key:PaginateOut.#.key,label:PaginateOut.#.label,title:PaginateOut.#.title,deletedAt:PaginateOut.#.deleted_at,description:PaginateOut.#.description,id:PaginateOut.#.id,thumb:PaginateOut.#.thumb,updatedAt:PaginateOut.#.updated_at,content:PaginateOut.#.content,createdAt:PaginateOut.#.created_at,icon:PaginateOut.#.icon}|@group}
*
*/
func TestPath(t *testing.T) {
	input := `{"code":200,"message":"ok","items":[{"id":1,"title":"test1"},{"id":2,"title":"test2"}],"pagination":{"index":0,"size":10,"total":100}}`
	pathMap := "{code:code.@tostring,message:message.@tostring,items.#.id:items.#.id.@tostring,items.#.title:items.#.title.@tostring,}"
	//pathMap = `{code:code.@tostring,message:message.@tostring,pagination:{index:pagination.index.@tostring,size:pagination.size.@tostring,total:pagination.total.@tostring},items:{id:items.#.id.@tostring,title:items.#.title.@tostring}|@group}`
	outputStr := gjson.GetBytes([]byte(input), pathMap).String()
	fmt.Println(outputStr)
}