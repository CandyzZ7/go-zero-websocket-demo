package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"gorm.io/gen"

	"go-zero-websocket-demo/infrastructure/pkg/ormengine"
	"go-zero-websocket-demo/internal/config"
)

var configFile = flag.String("f", "../../etc/app.yaml", "the config file")

//go:generate go run .

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	db, err := ormengine.NewGormEngine(c.MysqlConf)
	if err != nil {
		panic(fmt.Errorf("cannot establish db connection: %w", err))
	}
	g := gen.NewGenerator(gen.Config{
		// 相对执行`go run`时的路径, 会自动创建目录
		OutPath:      "../../internal/repository/generate/dal",
		ModelPkgPath: "../../model/entity",

		// WithDefaultQuery 生成默认查询结构体(作为全局变量使用), 即`Q`结构体和其字段(各表模型)
		// WithoutContext 生成没有context调用限制的代码供查询
		// WithQueryInterface 生成interface形式的查询代码(可导出), 如`Where()`方法返回的就是一个可导出的接口类型
		Mode: gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,

		// 表字段可为 null 值时, 对应结体字段使用指针类型
		FieldNullable: true, // generate pointer when field is nullable

		// 表字段默认值与模型结构体字段零值不一致的字段, 在插入数据时需要赋值该字段值为零值的, 结构体字段须是指针类型才能成功, 即`FieldCoverable:true`配置下生成的结构体字段.
		// 因为在插入时遇到字段为零值的会被GORM赋予默认值. 如字段`age`表默认值为10, 即使你显式设置为0最后也会被GORM设为10提交.
		// 如果该字段没有上面提到的插入时赋零值的特殊需要, 则字段为非指针类型使用起来会比较方便.
		FieldCoverable: false, // generate pointer when field has default value, to fix problem zero value cannot be assign: https://gorm.io/docs/create.html#Default-Values

		// 模型结构体字段的数字类型的符号表示是否与表字段的一致, `false`指示都用有符号类型
		FieldSignable: false, // detect integer field's unsigned type, adjust generated data type
		// 生成 gorm 标签的字段索引属性
		FieldWithIndexTag: false, // generate with gorm index tag
		// 生成 gorm 标签的字段类型属性
		FieldWithTypeTag: true, // generate with gorm column type tag

	})
	g.UseDB(db)
	// 自定义字段的数据类型
	// 统一数字类型为int64,兼容protobuf和thriftma
	dataMap := map[string]func(detailType gorm.ColumnType) (dataType string){
		"tinyint":   func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"smallint":  func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"mediumint": func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"bigint":    func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"int":       func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"decimal":   func(detailType gorm.ColumnType) (dataType string) { return "decimal.Decimal" }, // 金额类型全部转换为第三方库,github.com/shopspring/decimal
	}
	// 要先于`ApplyBasic`执行
	g.WithDataTypeMap(dataMap)

	// 自定义模型结体字段的标签
	// 将特定字段名的 json 标签加上`string`属性,即 MarshalJSON 时该字段由数字类型转成字符串类型
	// jsonField := gen.FieldJSONTagWithNS(func(columnName string) (tagContent string) {
	// 	if strings.Contains(`deleted_at`, columnName) {
	// 		return "-"
	// 	}
	// 	return xtransform.LowerCamelCase(columnName) // 下划线转小驼峰
	// })

	// 将非默认字段名的字段定义为自动时间戳和软删除字段;
	// 自动时间戳默认字段名为:`updated_at`、`created_at, 表字段数据类型为: INT 或 DATETIME
	// 软删除默认字段名为:`deleted_at`, 表字段数据类型为: DATETIME
	autoUpdateTimeField := gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
		return field.GormTag{
			"column":  []string{"updated_at"},
			"comment": []string{"更新时间"},
		}
	})
	autoCreateTimeField := gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
		return field.GormTag{
			"column":  []string{"created_at"},
			"comment": []string{"创建时间"},
		}
	})
	softDeleteField := gen.FieldType("deleted_at", "gorm.DeletedAt")
	// 模型自定义选项组
	fieldOpts := []gen.ModelOpt{autoCreateTimeField, autoUpdateTimeField, softDeleteField}
	// fieldOpts := []gen.ModelOpt{jsonField, softDeleteField}

	// 创建模型的结构体,生成文件在 model 目录; 先创建的结果会被后面创建的覆盖
	// 这里创建个别模型仅仅是为了拿到`*generate.QueryStructMeta`类型对象用于后面的模型关联操作中
	order := g.GenerateModel("order", fieldOpts...)
	// allModel := g.GenerateAllTable(fieldOpts...)
	g.ApplyBasic(order)
	g.Execute()
}
