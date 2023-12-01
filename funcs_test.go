package lineschema_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/lineschema"
)

func getDefaultJson() (defaul string, err error) {
	packschema := `version=http://json-schema.org/draft-07/schema#,id=out
	fullname=code,format=int,required,title=业务状态码,default=0,comment=业务状态码,example=0
	fullname=message,required,title=业务提示,default=ok,comment=业务提示,example=ok
	fullname=services[].id,format=int,required,title=主键,comment=主键,example=1
	fullname=services[].name,required,title=项目标识,comment=项目标识,example=advertise
	fullname=services[].title,required,title=名称,comment=名称
	fullname=services[].createdAt,format=datetime,required,title=创建时间,comment=创建时间,example=2023-01-1200:00:00
	fullname=services[].updatedAt,format=datetime,required,title=修改时间,comment=修改时间,example=2023-01-3000:00:00
	fullname=services[].servers[].name,required,title=服务标识,comment=服务标识,example=dev
	fullname=services[].servers[].title,required,title=服务名称,comment=服务名称,example=dev
	fullname=pagination.index,format=int,required,title=页索引,0开始,default=0,comment=页索引,0开始,example=0
	fullname=pagination.size,format=int,required,title=每页数量,default=10,comment=每页数量,example=10
	fullname=pagination.total,format=int,required,title=总数,comment=总数,example=60`
	lschema, err := lineschema.ParseLineschema(packschema)
	if err != nil {
		return "", err
	}
	jsonSchma, err := lschema.JsonSchema()
	if err != nil {
		return "", err
	}
	def, err := lineschema.GenerateDefaultJSON(jsonSchma)
	if err != nil {
		return "", err
	}
	s := string(def)
	return s, nil
}

func TestGenerateDefaultJSON(t *testing.T) {
	def, err := getDefaultJson()
	require.NoError(t, err)
	fmt.Println(def)
}

func TestMergeDefault(t *testing.T) {
	data := `{"code":0,"message":"","services":[{"id":1,"name":"advertise","title":"广告服务","createdAt":"2023-11-25 22:32:16","updatedAt":"2023-11-25 22:32:16","servers":[]}],"pagination":{"index":0,"size":10,"total":1}}`
	def, err := getDefaultJson()
	require.NoError(t, err)
	merge, err := lineschema.MergeDefault([]byte(data), []byte(def))
	require.NoError(t, err)
	m := string(merge)
	fmt.Println(m)

}
