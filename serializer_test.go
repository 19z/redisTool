package redisTool

import (
	"bytes"
	"encoding/json"
	"testing"
)

type TestStruct struct {
	Name string
	Age  int
}

type TestStructWithSerializer struct {
	Name string
	Age  int
}

func (t *TestStructWithSerializer) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TestStructWithSerializer) Deserialize(data []byte) error {
	return json.Unmarshal(data, t)
}

func TestDefaultSerializer_Serialize(t *testing.T) {
	s := &defaultSerializer{}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil value",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "string value",
			input:   "hello",
			wantErr: false,
		},
		{
			name:    "byte slice",
			input:   []byte("test"),
			wantErr: false,
		},
		{
			name:    "int value",
			input:   42,
			wantErr: false,
		},
		{
			name:    "int64 value",
			input:   int64(42),
			wantErr: false,
		},
		{
			name:    "float64 value",
			input:   3.14,
			wantErr: false,
		},
		{
			name:    "bool value",
			input:   true,
			wantErr: false,
		},
		{
			name:    "struct value",
			input:   TestStruct{Name: "Alice", Age: 30},
			wantErr: false,
		},
		{
			name:    "struct with serializer",
			input:   &TestStructWithSerializer{Name: "Bob", Age: 25},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := s.Serialize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.input != nil && len(data) == 0 {
				t.Errorf("Serialize() returned empty data for non-nil input")
			}
		})
	}
}

func TestDefaultSerializer_Deserialize(t *testing.T) {
	s := &defaultSerializer{}

	t.Run("empty data", func(t *testing.T) {
		var result string
		err := s.Deserialize([]byte{}, &result)
		if err != nil {
			t.Errorf("Deserialize() empty data error = %v", err)
		}
	})

	t.Run("string value", func(t *testing.T) {
		data, _ := s.Serialize("hello")
		var result string
		err := s.Deserialize(data, &result)
		if err != nil {
			t.Errorf("Deserialize() string error = %v", err)
		}
		if result != "hello" {
			t.Errorf("Deserialize() string = %v, want hello", result)
		}
	})

	t.Run("byte slice", func(t *testing.T) {
		original := []byte("test")
		data, _ := s.Serialize(original)
		var result []byte
		err := s.Deserialize(data, &result)
		if err != nil {
			t.Errorf("Deserialize() bytes error = %v", err)
		}
		if !bytes.Equal(result, original) {
			t.Errorf("Deserialize() bytes = %v, want %v", result, original)
		}
	})

	t.Run("int value", func(t *testing.T) {
		data, _ := s.Serialize(42)
		var result int
		err := s.Deserialize(data, &result)
		if err != nil {
			t.Errorf("Deserialize() int error = %v", err)
		}
		if result != 42 {
			t.Errorf("Deserialize() int = %v, want 42", result)
		}
	})

	t.Run("struct value", func(t *testing.T) {
		original := TestStruct{Name: "Alice", Age: 30}
		data, _ := s.Serialize(original)
		var result TestStruct
		err := s.Deserialize(data, &result)
		if err != nil {
			t.Errorf("Deserialize() struct error = %v", err)
		}
		if result.Name != original.Name || result.Age != original.Age {
			t.Errorf("Deserialize() struct = %v, want %v", result, original)
		}
	})

	t.Run("struct with serializer", func(t *testing.T) {
		original := &TestStructWithSerializer{Name: "Bob", Age: 25}
		data, _ := s.Serialize(original)
		var result TestStructWithSerializer
		err := s.Deserialize(data, &result)
		if err != nil {
			t.Errorf("Deserialize() struct with serializer error = %v", err)
		}
		if result.Name != original.Name || result.Age != original.Age {
			t.Errorf("Deserialize() struct = %v, want %v", result, *original)
		}
	})

	t.Run("non-pointer value", func(t *testing.T) {
		data := []byte("test")
		var result string
		err := s.Deserialize(data, result)
		if err == nil {
			t.Errorf("Deserialize() non-pointer should error")
		}
	})

	t.Run("invalid gob data", func(t *testing.T) {
		data := []byte{0xFF, 0xFF, 0xFF}
		var result TestStruct
		err := s.Deserialize(data, &result)
		if err == nil {
			t.Errorf("Deserialize() invalid gob data should error")
		}
	})
}

func TestRedisType_String(t *testing.T) {
	tests := []struct {
		redisType RedisType
		want      string
	}{
		{RedisTypeString, "string"},
		{RedisTypeList_, "list"},
		{RedisTypeSet_, "set"},
		{RedisTypeZSet_, "zset"},
		{RedisTypeHash_, "hash"},
		{RedisTypeQueue_, "queue"},
		{RedisTypeCache_, "cache"},
		{RedisTypeLock_, "lock"},
		{RedisTypeSafeTypeMap_, "safetypemap"},
		{RedisType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.redisType.String(); got != tt.want {
				t.Errorf("RedisType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
