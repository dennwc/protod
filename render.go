package protod

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// Render writes a .proto file based on a FileDescriptorProto.
func Render(w io.Writer, fd *FileDescriptorProto) error {
	wr := &writer{bw: bufio.NewWriter(w)}
	wr.renderDescriptor(fd)
	if err := wr.err; err != nil {
		return err
	}
	return wr.bw.Flush()
}

type writer struct {
	bw   *bufio.Writer
	tabs []byte
	path []string
	err  error

	pkg    string
	proto3 bool
}

func (w *writer) writeString(s string) {
	if w.err == nil {
		_, w.err = w.bw.WriteString(s)
	}
}

func (w *writer) writeUint(v uint64) {
	if w.err == nil {
		_, w.err = w.bw.WriteString(strconv.FormatUint(v, 10))
	}
}

func (w *writer) push(name string) {
	w.tabs = append(w.tabs, '\t')
	w.path = append(w.path, name)
}

func (w *writer) pop() {
	w.tabs = w.tabs[:len(w.tabs)-1]
	w.path = w.path[:len(w.path)-1]
}

func (w *writer) tab() {
	if w.err != nil {
		return
	}
	_, w.err = w.bw.Write(w.tabs)
}

func (w *writer) printf(format string, args ...interface{}) {
	w.tab()
	if w.err != nil {
		return
	}
	_, w.err = fmt.Fprintf(w.bw, format, args...)
}

func (w *writer) renderHeader(fd *FileDescriptorProto) {
	syntax := fd.GetSyntax()
	if syntax == "" {
		syntax = "proto2"
	}
	w.printf("syntax = %q;\n\n", syntax)

	w.printf("package %s;\n\n", fd.GetPackage())

	for i, d := range fd.Dependency {
		w.printf("import %q;\n", d)
		if i+1 == len(fd.Dependency) {
			w.writeString("\n")
		}
	}
	if fd.Options != nil {
		// TODO(dennwc): support file options
		log.Printf("TODO: %s: file options %v", w.pkg, fd.Options)
	}
	if len(fd.Extension) != 0 {
		// TODO(dennwc): support extensions
		log.Printf("TODO: %s: extensions", w.pkg)
	}
	if fd.SourceCodeInfo != nil && len(fd.SourceCodeInfo.Location) != 0 {
		// TODO(dennwc): support location info; haven't seen it so far
	}
}

func (w *writer) renderDescriptor(fd *FileDescriptorProto) {
	w.pkg = fd.GetPackage()
	w.path = strings.Split(w.pkg, ".")
	w.proto3 = fd.GetSyntax() == "proto3"
	w.renderHeader(fd)

	for _, e := range fd.EnumType {
		w.renderEnum(e)
	}
	for _, m := range fd.MessageType {
		w.renderMessage(m)
	}
	for _, s := range fd.Service {
		w.renderService(s)
	}
}

func (w *writer) renderReservedNames(names []string) {
	if len(names) == 0 {
		return
	}
	w.tab()
	w.writeString("reserved ")
	for i, r := range names {
		if i != 0 {
			w.writeString(", ")
		}
		w.writeString(strconv.Quote(r))
	}
	w.writeString(";\n")
}

func (w *writer) renderEnum(e *descriptor.EnumDescriptorProto) {
	name := e.GetName()
	w.printf("enum %s {\n", name)
	w.push(name)
	defer func() {
		w.pop()
		w.printf("}\n\n")
	}()
	if e.Options != nil {
		// TODO(dennwc): enum options
		log.Printf("TODO: %s: enum options %q", w.pkg, name)
	}
	for _, v := range e.Value {
		w.tab()
		w.writeString(v.GetName())
		w.writeString("\t= ")
		w.writeUint(uint64(v.GetNumber()))
		if v.Options != nil {
			// TODO(dennwc): enum value options
			log.Printf("TODO: %s: enum value options %q", w.pkg, v.GetName())
		}
		w.writeString(";\n")
	}
	w.renderReservedNames(e.ReservedName)
	if len(e.ReservedRange) != 0 {
		w.tab()
		w.writeString("reserved ")
		for i, r := range e.ReservedRange {
			if i != 0 {
				w.writeString(", ")
			}
			if r.Start != nil && r.End != nil {
				s, e := *r.Start, *r.End
				w.writeUint(uint64(*r.Start))
				if s != e {
					w.writeString(" to ")
					w.writeUint(uint64(*r.End))
				}
			} else if r.Start != nil {
				w.writeUint(uint64(*r.Start))
				w.writeString(" to max")
			}
		}
		w.writeString(";\n")
	}
}

func (w *writer) renderMessage(m *descriptor.DescriptorProto) {
	name := m.GetName()
	w.printf("message %s {\n", name)
	w.push(name)
	defer func() {
		w.pop()
		w.printf("}\n\n")
	}()
	for _, e2 := range m.EnumType {
		w.renderEnum(e2)
	}
	for _, m2 := range m.NestedType {
		w.renderMessage(m2)
	}
	if len(m.Extension)+len(m.ExtensionRange) != 0 {
		// TODO(dennwc): extensions fields
		log.Printf("TODO: %s: extensions fields", w.pkg)
	}
	for _, f := range m.Field {
		w.renderField(f)
	}
	if len(m.OneofDecl) != 0 {
		// TODO(dennwc): oneof fields
		log.Printf("TODO: %s: oneof fields", w.pkg)
	}
	w.renderReservedNames(m.ReservedName)
	if len(m.ReservedRange) != 0 {
		w.tab()
		w.writeString("reserved ")
		for i, r := range m.ReservedRange {
			if i != 0 {
				w.writeString(", ")
			}
			if r.Start != nil && r.End != nil {
				s, e := *r.Start, *r.End
				w.writeUint(uint64(*r.Start))
				if s != e {
					w.writeString(" to ")
					w.writeUint(uint64(*r.End))
				}
			} else if r.Start != nil {
				w.writeUint(uint64(*r.Start))
				w.writeString(" to max")
			}
		}
		w.writeString(";\n")
	}
}

var typeToStr = map[descriptor.FieldDescriptorProto_Type]string{
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   "double",
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    "float",
	descriptor.FieldDescriptorProto_TYPE_INT64:    "int64",
	descriptor.FieldDescriptorProto_TYPE_UINT64:   "uint64",
	descriptor.FieldDescriptorProto_TYPE_INT32:    "int32",
	descriptor.FieldDescriptorProto_TYPE_UINT32:   "uint32",
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  "fixed64",
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  "fixed32",
	descriptor.FieldDescriptorProto_TYPE_BOOL:     "bool",
	descriptor.FieldDescriptorProto_TYPE_STRING:   "string",
	descriptor.FieldDescriptorProto_TYPE_BYTES:    "bytes",
	descriptor.FieldDescriptorProto_TYPE_SINT64:   "sint64",
	descriptor.FieldDescriptorProto_TYPE_SINT32:   "sint32",
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: "sfixed64",
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: "sfixed32",
	descriptor.FieldDescriptorProto_TYPE_GROUP:    "group",
}

func (w *writer) getType(name string) string {
	name = strings.TrimPrefix(name, ".")
	sub := strings.Split(name, ".")
	if len(sub) == 1 {
		return name
	}
	pkg := w.path

	for i := 0; i < len(sub) && i < len(pkg); i++ {
		if sub[i] != pkg[i] {
			return strings.Join(sub[i:], ".")
		}
	}
	if len(pkg) < len(sub) {
		return strings.Join(sub[len(pkg):], ".")
	}
	return sub[len(sub)-1]
}

func (w *writer) renderField(f *descriptor.FieldDescriptorProto) {
	if w.err != nil {
		return
	}
	w.tab()
	switch f.GetLabel() {
	case 0:
		if !w.proto3 {
			w.err = fmt.Errorf("empty label for field %q", f.GetName())
			return
		}
	case descriptor.FieldDescriptorProto_LABEL_OPTIONAL:
		if !w.proto3 {
			w.writeString("optional ")
		}
	case descriptor.FieldDescriptorProto_LABEL_REQUIRED:
		w.writeString("required ")
	case descriptor.FieldDescriptorProto_LABEL_REPEATED:
		w.writeString("repeated ")
	default:
		w.err = fmt.Errorf("unknown label for field %q: %v", f.GetName(), f.GetLabel())
		return
	}
	typ := ""
	switch t := f.GetType(); t {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		typ = w.getType(f.GetTypeName())
	default:
		var ok bool
		typ, ok = typeToStr[t]
		if !ok {
			w.err = fmt.Errorf("unknown message field type: %v", f.GetType())
			return
		}
	}
	w.writeString(typ + "\t")
	w.writeString(f.GetName())
	w.writeString("\t= ")
	w.writeUint(uint64(f.GetNumber()))

	if def := f.GetDefaultValue(); def != "" {
		w.writeString(" [default = ")
		w.writeString(def)
		w.writeString("]")
	}
	if f.Options != nil {
		// TODO(dennwc): field options
		log.Printf("TODO: %s: field options %q", w.pkg, f.GetName())
	}
	w.writeString(";\n")
}

func (w *writer) renderService(s *descriptor.ServiceDescriptorProto) {
	name := s.GetName()
	w.printf("service %s {\n", name)
	w.push(name)
	defer func() {
		w.pop()
		w.printf("}\n\n")
	}()
	for _, m := range s.Method {
		in := w.getType(m.GetInputType())
		out := w.getType(m.GetOutputType())
		if m.GetClientStreaming() {
			in = "stream " + in
		}
		if m.GetServerStreaming() {
			out = "stream " + out
		}
		w.printf("rpc %s (%s) returns (%s) {}\n", m.GetName(), in, out)
		if m.Options != nil {
			// TODO(dennwc): service method options
			log.Printf("TODO: %s: service method options %q", w.pkg, m.GetName())
		}
	}
	if s.Options != nil {
		// TODO(dennwc): service options
		log.Printf("TODO: %s: service options %q", w.pkg, s.GetName())
	}
}
