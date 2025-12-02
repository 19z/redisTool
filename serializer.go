package redisTool

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
)

// defaultSerializer 默认序列化器
type defaultSerializer struct{}

// DefaultSerializer 默认序列化器实例
var DefaultSerializer SerializerFunc = &defaultSerializer{}

// Serialize 序列化
func (s *defaultSerializer) Serialize(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	// 如果实现了 Serializer 接口
	if serializer, ok := v.(Serializer); ok {
		return serializer.Serialize()
	}

	// 处理基本类型
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return json.Marshal(v)
	}

	// 使用 gob 序列化其他类型
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("gob encode error: %w", err)
	}
	return buf.Bytes(), nil
}

// Deserialize 反序列化
func (s *defaultSerializer) Deserialize(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// 如果实现了 Serializer 接口
	if serializer, ok := v.(Serializer); ok {
		return serializer.Deserialize(data)
	}

	// 获取指针指向的值
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	elem := val.Elem()
	
	// 处理 interface{} 类型：先尝试作为字符串，如果失败再尝试 gob
	if elem.Kind() == reflect.Interface {
		// 先尝试作为字符串
		elem.Set(reflect.ValueOf(string(data)))
		return nil
	}
	
	// 处理基本类型
	switch elem.Kind() {
	case reflect.String:
		elem.SetString(string(data))
		return nil
	case reflect.Slice:
		if elem.Type().Elem().Kind() == reflect.Uint8 { // []byte
			elem.SetBytes(data)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return json.Unmarshal(data, v)
	}

	// 使用 gob 反序列化其他类型
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("gob decode error: %w", err)
	}
	return nil
}
