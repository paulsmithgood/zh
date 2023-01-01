package model

import (
	"errors"
	"exercise/geektime/homework5/version1/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Option func(m *Model) error

// Registry 元数据注册中心的抽象
type Registry interface {
	// Get 查找元数据
	Get(val any) (*Model, error)
	// Register 注册一个模型
	Register(val any, opts ...Option) (*Model, error)
}

// registry 基于标签和接口的实现
// 目前来看，我们只有一个实现，所以暂时可以维持私有
type registry struct {
	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

// Get 查找元数据模型
func (r *registry) Get(val any) (*Model, error) {
	//fmt.Println(val)
	//if val.(*Model).TableName != "" {
	//	return val.(*Model), nil
	//}
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	return r.Register(val)
}

func (r *registry) Register(val any, opts ...Option) (*Model, error) {
	m, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err = opt(m)
		if err != nil {
			return nil, err
		}
	}
	typ := reflect.TypeOf(val)
	r.models.Store(typ, m)
	return m, nil
}

// parseModel 支持从标签中提取自定义设置
// 标签形式 orm:"key1=value1,key2=value2"
func (r *registry) parseModel(val any) (*Model, error) {
	//	先解析 一下model
	typ := reflect.TypeOf(val)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("ParseModel 只支持 解析 Struct")
	}
	num := typ.NumField() //结构体中 有 num 个 字段

	FieldMap := make(map[string]*Field, num)
	ColumnMap := make(map[string]*Field, num)

	for i := 0; i < num; i++ {
		fd := typ.Field(i)
		tag, err := r.parseTag(fd.Tag)
		//fmt.Println(tag, err, "s")
		if err != nil {
			return nil, err
		}
		tagname := tag["column"]
		if tagname == "" {
			tagname = underscoreName(fd.Name)
		}
		field := &Field{
			ColName: tagname,
			GoName:  fd.Name,
			Type:    fd.Type,
			Offset:  fd.Offset,
		}
		FieldMap[fd.Name] = field
		ColumnMap[tagname] = field
	}
	res := &Model{FieldMap: FieldMap, ColumnMap: ColumnMap, TableName: ""}

	var tableName string
	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}

	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}
	res.TableName = tableName
	return res, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		// 返回一个空的 map，这样调用者就不需要判断 nil 了
		return map[string]string{}, nil
	}
	// 这个初始化容量就是我们支持的 key 的数量，
	// 现在只有一个，所以我们初始化为 1
	res := make(map[string]string, 1)

	// 接下来就是字符串处理了
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}

func WithTableName(tableName string) Option {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

func WithColumnName(field string, columnName string) Option {
	return func(model *Model) error {
		fd, ok := model.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		// 注意，这里我们根本没有检测 ColName 会不会是空字符串
		// 因为正常情况下，用户都不会写错
		// 即便写错了，也很容易在测试中发现
		fd.ColName = columnName
		return nil
	}
}
