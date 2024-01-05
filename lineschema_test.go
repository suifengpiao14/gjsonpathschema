package lineschema_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/lineschema"
)

func TestResolveRef(t *testing.T) {
	ls, err := lineschema.ParseLineschema(packschema)
	require.NoError(t, err)
	fs := ls.ResolveRef()
	s := fs.String()
	fmt.Println(s)

}

var packschema = `version=http://json-schema.org/draft-07/schema#,id=out
fullname=domain,required,allowEmptyValue,title=领域,comment=领域
fullname=scene,required,allowEmptyValue,title=应用场景,comment=应用场景
fullname=name,required,allowEmptyValue,title=名称,comment=名称
fullname=title,required,allowEmptyValue,title=标题,comment=标题
fullname=path,required,allowEmptyValue,title=路径,comment=路径
fullname=method,required,allowEmptyValue,title=请求方法,comment=请求方法
fullname=summary,required,allowEmptyValue,title=摘要,comment=摘要
fullname=description,required,allowEmptyValue,title=介绍,comment=介绍
fullname=service.name,required,allowEmptyValue,title=服务标识,comment=服务标识
fullname=service.servers[].name,required,allowEmptyValue,title=服务器名称,comment=服务器名称
fullname=service.servers[].title,required,allowEmptyValue
fullname=service.servers[].url,required,allowEmptyValue,title=url地址,comment=url地址
fullname=service.servers[].ip,required,allowEmptyValue,title=服务器IP,comment=服务器IP
fullname=service.servers[].description,required,allowEmptyValue,title=介绍,comment=介绍
fullname=service.servers[].proxy,required,allowEmptyValue,title=代理地址,comment=代理地址
fullname=service.servers[].extensionIds,required,allowEmptyValue,title=扩展字段,comment=扩展字段
fullname=service.title,required,allowEmptyValue,title=标题,comment=标题
fullname=service.description,required,allowEmptyValue,title=介绍,comment=介绍
fullname=service.version,required,allowEmptyValue,title=版本,comment=版本
fullname=service.contacts[].name,required,allowEmptyValue,title=联系人名称,comment=联系人名称
fullname=service.contacts[].phone,required,allowEmptyValue,title=联系人手机号,comment=联系人手机号
fullname=service.contacts[].email,required,allowEmptyValue,title=联系人邮箱,comment=联系人邮箱
fullname=service.license,required,allowEmptyValue,title=协议,comment=协议
fullname=service.security,required,allowEmptyValue,title=鉴权,comment=鉴权
fullname=service.requestPreScript,type=Scripts,required,allowEmptyValue,title=前置脚本,comment=前置脚本
fullname=service.requestPostScript,type=Scripts,required,allowEmptyValue,title=后置请求脚本,comment=后置请求脚本
fullname=service.variables,required,allowEmptyValue,title=变量,comment=变量
fullname=service.navigates,required,allowEmptyValue,title=导航,comment=导航
fullname=service.document,required,allowEmptyValue,title=文档,comment=文档
fullname=requestHeader,type=Parameters,required,allowEmptyValue
fullname=responseHeader,type=Parameters,required,allowEmptyValue
fullname=query,type=Parameters,required,allowEmptyValue
fullname=requestBody,type=Parameters,required,allowEmptyValue
fullname=responseBody,type=Parameters,required,allowEmptyValue
fullname=examples[].tag,required,allowEmptyValue,title=标签,description=标签,mock数据时不同接口案例优先返回相同tag案例,comment=标签,mock数据时不同接口案例优先返回相同tag案例
fullname=examples[].method,required,allowEmptyValue,title=请求方法,comment=请求方法
fullname=examples[].title,required,allowEmptyValue,title=案例名称,comment=案例名称
fullname=examples[].summary,required,allowEmptyValue,title=简介,comment=简介
fullname=examples[].url,required,allowEmptyValue,title=请求地址,comment=请求地址
fullname=examples[].proxy,required,allowEmptyValue,title=代理地址,comment=代理地址
fullname=examples[].auth,required,allowEmptyValue,title=鉴权,comment=鉴权
fullname=examples[].headers,type=object,required,allowEmptyValue,title=请求头,comment=请求头
fullname=examples[].contentType,required,allowEmptyValue,title=请求格式,comment=请求格式
fullname=examples[].requestPreScript,type=Scripts,required,allowEmptyValue
fullname=examples[].requestPostScript,type=Scripts,required,allowEmptyValue
fullname=examples[].requestBody,required,allowEmptyValue,title=请求体,comment=请求体
fullname=examples[].testScript,required,allowEmptyValue,title=返回体测试脚本,comment=返回体测试脚本
fullname=examples[].response,required,allowEmptyValue,title=请求体,comment=请求体
fullname=links[].name,required,allowEmptyValue,title=名称,comment=名称
fullname=links[].value,required,allowEmptyValue,title=值,comment=值
fullname=links[].description,required,allowEmptyValue,title=描述,comment=描述
fullname=code,format=int,required,title=0,default=0,comment=0
fullname=message,required,title=ok,default=ok,comment=ok
fullname=Parameters,type=[]Parameter,required,allowEmptyValue,title=参数集合,comment=参数集合
fullname=Parameter.title,required,allowEmptyValue,title=验证规则标识,comment=验证规则标识
fullname=Parameter.fullname,required,allowEmptyValue,title=名称,description=名称(冗余local.en),comment=名称(冗余local.en)
fullname=Parameter.name,required,allowEmptyValue,title=参数类型,description=参数类型(string-字符,int-整型,number-数字,array-数组,object-对象),comment=参数类型(string-字符,int-整型,number-数字,array-数组,object-对象)
fullname=Parameter.type,required,allowEmptyValue,title=参数所在的位置,description=参数所在的位置(body-BODY,head-HEAD,comment=参数所在的位置(body-BODY,head-HEAD
fullname=Parameter.position,required,allowEmptyValue
fullname=Parameter.format,required,allowEmptyValue,title=案例,comment=案例
fullname=Parameter.example,required,allowEmptyValue
fullname=Parameter.default,required,allowEmptyValue,title=是否弃用,description=是否弃用(true-是,false-否),comment=是否弃用(true-是,false-否)
fullname=Parameter.deprecated,required,allowEmptyValue,title=是否必须,description=是否必须(true-是,false-否),comment=是否必须(true-是,false-否)
fullname=Parameter.required,required,allowEmptyValue,title=对数组,description=对数组、对象序列化方法,参照openapiparameters.style,comment=对数组、对象序列化方法,参照openapiparameters.style
fullname=Parameter.serialize,required,allowEmptyValue,title=对象的key,description=对象的key,是否单独成参数方式,参照openapiparameters.explode(true-是,false-否),comment=对象的key,是否单独成参数方式,参照openapiparameters.explode(true-是,false-否)
fullname=Parameter.explode,required,allowEmptyValue,title=是否容许空值,description=是否容许空值(true-是,false-否),comment=是否容许空值(true-是,false-否)
fullname=Parameter.allowEmptyValue,required,allowEmptyValue,title=特殊字符是否容许出现在uri参数中,description=特殊字符是否容许出现在uri参数中(true-是,false-否),comment=特殊字符是否容许出现在uri参数中(true-是,false-否)
fullname=Parameter.allowReserved,required,allowEmptyValue,title=简介,comment=简介
fullname=Parameter.description,required,allowEmptyValue
fullname=Parameter.enum,required,allowEmptyValue
fullname=Parameter.enumNames,required,allowEmptyValue
fullname=Parameter.scene,required,allowEmptyValue
fullname=Scripts,type=[]Script,required,allowEmptyValue,title=code脚本集合,comment=code脚本集合
fullname=Script.language,required,allowEmptyValue,title=前置脚本语言,comment=前置脚本语言
fullname=Script.script,required,allowEmptyValue,title=前置脚本,comment=前置脚本`
