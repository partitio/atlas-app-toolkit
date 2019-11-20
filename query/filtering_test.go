package query

import (
	"regexp/syntax"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

type TestObject struct {
	Str   string  `json:"str"`
	Float float64 `json:"float"`
	Uint  uint    `json:"uint"`
	Ptr   *struct{}
}

type TestProtoMessage struct {
	StringValue *wrappers.StringValue `protobuf:"bytes,1,opt,name=string_value,proto3" json:"id,omitempty"`
	IntValue    *wrappers.Int64Value  `protobuf:"bytes,1,opt,name=int_value,proto3" json:"id,omitempty"`
	Str         string                `protobuf:"bytes,1,opt,name=str"`
	Int         int32                 `protobuf:"varint,2,opt,name=int"`
	Bool        bool                  `protobuf:"bytes,3,opt,name=bool,proto3" json:"id,omitempty"`
	Nested      *NestedMessage        `protobuf:"bytes,3,opt,name=nested,json=nestedJSON"`
	Enum        Enum                  `protobuf:"varint,6,opt,name=enum,proto3,enum=query.Enum" json:"enum,omitempty"`
}

func (m *TestProtoMessage) Reset()         { *m = TestProtoMessage{} }
func (m *TestProtoMessage) String() string { return proto.CompactTextString(m) }
func (*TestProtoMessage) ProtoMessage()    {}

type NestedMessage struct {
	Str string `protobuf:"bytes,1,opt,name=str"`
}

func (m *NestedMessage) Reset()         { *m = NestedMessage{} }
func (m *NestedMessage) String() string { return proto.CompactTextString(m) }
func (*NestedMessage) ProtoMessage()    {}

type Enum int32

const (
	ENUM_ONE Enum = 0
	ENUM_TwO Enum = 1
)

var Enum_name = map[int32]string{
	0: "ONE",
	1: "TW0",
}

var Enum_value = map[string]int32{
	"ONE": 0,
	"TW0": 1,
}

func (x Enum) String() string {
	return proto.EnumName(Enum_name, int32(x))
}

func TestFiltering(t *testing.T) {

	tests := []struct {
		obj    interface{}
		filter string
		res    bool
	}{
		{
			obj: &TestProtoMessage{
				Str:         "111",
				Int:         111,
				StringValue: &wrappers.StringValue{Value: "111"},
				IntValue:    &wrappers.Int64Value{Value: 111},
				Enum:        ENUM_TwO,
			},
			filter: "enum == 1 and string_value == '111' and int_value == 111 and str == '111' and int == 111 and nestedJSON == null",
			res:    true,
		},
		{
			obj:    &TestObject{Str: "111", Float: 11.11, Uint: 11},
			filter: "str == '111' and float == 11.11 and uint == 11 and Ptr == null",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Str: "111", Int: 111},
			filter: "str == '111' and int == 111 and nestedJSON != null",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "111", Int: 111},
			filter: "str == '222' and int == 111 and nestedJSON == null",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "111", Int: 111},
			filter: "str == '111' or int == 222 or nestedJSON != null",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Str: "111", Int: 111},
			filter: "str == '222' or not int == 222",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Str: "111"},
			filter: "str == '111'",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Str: "111"},
			filter: "not str == '111'",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "111"},
			filter: "str == '222'",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "111"},
			filter: "str ~ '1*'",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Str: "111"},
			filter: "str !~ '1112?'",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "111"},
			filter: "str ~ '[23]1*'",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "aaa"},
			filter: "str > 'aaa'",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "aaa"},
			filter: "str >= 'aaa'",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Str: "aaa"},
			filter: "str < 'aaa'",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Str: "aaa"},
			filter: "str <= 'aaa'",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "int == 111",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "not int == 111",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "int == 222",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "int > 110",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "int >= 111",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "int < 112",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{Int: 111},
			filter: "int <= 111",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{},
			filter: "nestedJSON == null",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{},
			filter: "not nestedJSON == null",
			res:    false,
		},
		{
			obj:    &TestProtoMessage{Nested: &NestedMessage{}},
			filter: "nestedJSON == null",
			res:    false,
		},
		{
			obj: &TestProtoMessage{
				Bool: true,
			},
			filter: "bool == true",
			res:    true,
		},
		{
			obj: &TestProtoMessage{
				Bool: true,
			},
			filter: "bool == false",
			res:    false,
		},
		{
			obj: &TestProtoMessage{
				Bool: false,
			},
			filter: "bool != true",
			res:    true,
		},
		{
			obj:    &TestProtoMessage{},
			filter: "",
			res:    true,
		},
	}
	for _, test := range tests {
		res, err := Filter(test.obj, test.filter)
		assert.Equal(t, res, test.res)
		assert.Nil(t, err)
	}
}

func TestFilteringNegative(t *testing.T) {

	tests := []struct {
		obj    interface{}
		filter string
		err    error
	}{
		{
			obj:    &TestObject{Str: "111"},
			filter: "str == 111",
			err:    &TypeMismatchError{},
		},
		{
			obj:    &TestObject{Float: 11.11},
			filter: "float == '11.11'",
			err:    &TypeMismatchError{},
		},
		{
			obj:    &TestObject{},
			filter: "float == null",
			err:    &TypeMismatchError{},
		},
		{
			obj:    &TestObject{},
			filter: "missingField == 11.11",
			err:    &TypeMismatchError{},
		},
		{
			obj:    &TestProtoMessage{},
			filter: "missingField == 11.11",
			err:    &TypeMismatchError{},
		},
		{
			obj:    &TestObject{Str: "111"},
			filter: "str ~ '11[1'",
			err:    &syntax.Error{},
		},
	}

	for _, test := range tests {
		res, err := Filter(test.obj, test.filter)
		assert.False(t, res)
		assert.IsType(t, test.err, err)
	}

}
