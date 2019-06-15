package protod

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"io"
	"io/ioutil"
)

// FileDescriptorProto is an alias for protobuf's FileDescriptorProto.
type FileDescriptorProto = descriptor.FileDescriptorProto

// ExtractFunc is a callback for Extract functions. The first argument is an encoded proto file descriptor,
// and the second one is a decoded file descriptor.
type ExtractFunc func(data []byte, fd *FileDescriptorProto) error

// Extract reads data from r and calls a function for each occurrence of protobuf file descriptors.
func Extract(r io.Reader, fnc ExtractFunc) error {
	// TODO: do not load everything into memory
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return ExtractBytes(data, fnc)
}

// ExtractBytes calls a function for each occurrence of protobuf file descriptors in a bytes slice.
func ExtractBytes(data []byte, fnc ExtractFunc) error {
	seen := make(map[string]struct{})
loop:
	for len(data) > 0 {
		i := bytes.Index(data, []byte(".proto"))
		if i < 0 {
			break
		}
		for s := 0; s < 64; s++ {
			fi := i - s - 1
			if fi < 0 || data[fi] != 0x0a {
				continue
			}
			if v, n := binary.Uvarint(data[fi+1:]); int(v) != s-n+6 || !isFileName(data[fi+1+n:i+6]) {
				continue
			}
			psize := skipProto(data[fi:])
			if psize < 0 {
				continue
			}
			// Skip size is not perfect, so we add a delta of 1024 bytes
			// to make sure we not miss anything.
			for sz := psize + 1024; sz > 0; sz-- {
				pdata := data[fi : fi+sz]
				var fd descriptor.FileDescriptorProto
				if err := proto.Unmarshal(pdata, &fd); err == nil {
					data = data[fi+sz:]
					if _, ok := seen[string(pdata)]; ok {
						continue loop
					}
					seen[string(pdata)] = struct{}{}
					if err = fnc(pdata, &fd); err != nil {
						return err
					}
					break
				}
			}
			break
		}
		data = data[i+6:]
	}
	return nil
}

func skipProto(data []byte) int {
	if len(data) == 0 {
		return -1
	}
	off := 0
loop:
	for off < len(data) {
		tag, n := binary.Uvarint(data[off:])
		if n == 0 || tag == 0 {
			break
		}
		i := off + n
		switch wireType := tag & 0x07; wireType {
		case 0: // varint
			_, n2 := binary.Uvarint(data[i:])
			if n2 == 0 {
				break loop
			}
			i += n2
		case 1: // 64-bit
			i += 8
		case 2: // length-delimited
			sz, n2 := binary.Uvarint(data[i:])
			if n2 == 0 {
				break loop
			}
			i += int(sz) + n2
		case 3, 4: // groups, deprecated
		case 5: // 32-bit
			i += 4
		default:
			break loop
		}
		if i > len(data) {
			break
		}
		off = i
	}
	return off
}

func isFileName(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_/$,.[]()"
	for _, b := range data {
		if b == 0x00 || bytes.IndexByte([]byte(charset), b) < 0 {
			return false
		}
	}
	return true
}
