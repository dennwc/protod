package protod

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"testing"
)

func str(s string) *string {
	return &s
}

var testFD = &FileDescriptorProto{
	Name:    str("some.proto"),
	Package: str("some.package"),
	Dependency: []string{
		"one.proto",
		"two.proto",
	},
	PublicDependency: []int32{
		1, 2, 3,
	},
	MessageType: []*descriptor.DescriptorProto{
		{Name: str("SomeMsg")},
	},
}

func TestSkipProto(t *testing.T) {
	data, err := proto.Marshal(testFD)
	if err != nil {
		t.Fatal(err)
	}
	exp := len(data)
	data = append(data, 1, 2, 3, 4) // "random" garbage

	i := skipProto(data)
	if i < exp {
		t.Fatal("unexpected skip index:", i, "<", exp)
	}
}
