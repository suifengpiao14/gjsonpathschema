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

func TestJson2Lineschema(t *testing.T) {
	jsonStr := `[{"DatabaseConfig":{"databaseName":"ad","tablePrefix":"","columnPrefix":"","deletedAtColumn":"deleted_at","logLevel":"","version":"","extaConfigs":null},"TableName":"creative","PrimaryKey":"id","DeleteColumn":"deleted_at","Columns":[{"Prefix":"","CamelName":"Id","ColumnName":"id","Name":"id","Type":"int","Comment":"主键","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":true,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"PlanId","ColumnName":"plan_id","Name":"plan_id","Type":"string","Comment":"广告计划Id","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Name","ColumnName":"name","Name":"name","Type":"string","Comment":"名称","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Content","ColumnName":"content","Name":"content","Type":"string","Comment":"广告内容","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"CreatedAt","ColumnName":"created_at","Name":"created_at","Type":"string","Comment":"创建时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"current_timestamp()","OnCreate":true,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"UpdatedAt","ColumnName":"updated_at","Name":"updated_at","Type":"string","Comment":"修改时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"current_timestamp()","OnCreate":false,"OnUpdate":true,"OnDelete":false},{"Prefix":"","CamelName":"DeletedAt","ColumnName":"deleted_at","Name":"deleted_at","Type":"string","Comment":"删除时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"NULL","OnCreate":false,"OnUpdate":false,"OnDelete":true}],"EnumsConst":[],"Comment":"广告物料","TableDef":null},{"DatabaseConfig":{"databaseName":"ad","tablePrefix":"","columnPrefix":"","deletedAtColumn":"deleted_at","logLevel":"","version":"","extaConfigs":null},"TableName":"plan","PrimaryKey":"id","DeleteColumn":"deleted_at","Columns":[{"Prefix":"","CamelName":"Id","ColumnName":"id","Name":"id","Type":"int","Comment":"主键","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":true,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"AdvertiserId","ColumnName":"advertiser_id","Name":"advertiser_id","Type":"string","Comment":"广告主","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Name","ColumnName":"name","Name":"name","Type":"string","Comment":"名称","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Position","ColumnName":"position","Name":"position","Type":"string","Comment":"位置编码","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"BeginAt","ColumnName":"begin_at","Name":"begin_at","Type":"string","Comment":"投放开始时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"NULL","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"EndAt","ColumnName":"end_at","Name":"end_at","Type":"string","Comment":"投放结束时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"NULL","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Did","ColumnName":"did","Name":"did","Type":"int","Comment":"出价","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"0","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"LandingPage","ColumnName":"landing_page","Name":"landing_page","Type":"string","Comment":"落地页","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"CreatedAt","ColumnName":"created_at","Name":"created_at","Type":"string","Comment":"创建时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"current_timestamp()","OnCreate":true,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"UpdatedAt","ColumnName":"updated_at","Name":"updated_at","Type":"string","Comment":"修改时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"current_timestamp()","OnCreate":false,"OnUpdate":true,"OnDelete":false},{"Prefix":"","CamelName":"DeletedAt","ColumnName":"deleted_at","Name":"deleted_at","Type":"string","Comment":"删除时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"NULL","OnCreate":false,"OnUpdate":false,"OnDelete":true}],"EnumsConst":[],"Comment":"广告计划","TableDef":null},{"DatabaseConfig":{"databaseName":"ad","tablePrefix":"","columnPrefix":"","deletedAtColumn":"deleted_at","logLevel":"","version":"","extaConfigs":null},"TableName":"window","PrimaryKey":"id","DeleteColumn":"deleted_at","Columns":[{"Prefix":"","CamelName":"Id","ColumnName":"id","Name":"id","Type":"int","Comment":"主键","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":true,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"MediaId","ColumnName":"media_id","Name":"media_id","Type":"string","Comment":"媒体Id","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Position","ColumnName":"position","Name":"position","Type":"string","Comment":"位置编码","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Name","ColumnName":"name","Name":"name","Type":"string","Comment":"位置名称","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Remark","ColumnName":"remark","Name":"remark","Type":"string","Comment":"广告位描述(建议记录位置、app名称等)","Tag":"","Nullable":false,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"Scheme","ColumnName":"scheme","Name":"scheme","Type":"string","Comment":"广告素材的格式规范","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"","OnCreate":false,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"CreatedAt","ColumnName":"created_at","Name":"created_at","Type":"string","Comment":"创建时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"current_timestamp()","OnCreate":true,"OnUpdate":false,"OnDelete":false},{"Prefix":"","CamelName":"UpdatedAt","ColumnName":"updated_at","Name":"updated_at","Type":"string","Comment":"修改时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"current_timestamp()","OnCreate":false,"OnUpdate":true,"OnDelete":false},{"Prefix":"","CamelName":"DeletedAt","ColumnName":"deleted_at","Name":"deleted_at","Type":"string","Comment":"删除时间","Tag":"","Nullable":true,"Enums":[],"AutoIncrement":false,"DefaultValue":"NULL","OnCreate":false,"OnUpdate":false,"OnDelete":true}],"EnumsConst":[],"Comment":"广告位表","TableDef":null}]`
	lineschema, err := lineschema.Json2lineSchema(jsonStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(lineschema.String())
}
