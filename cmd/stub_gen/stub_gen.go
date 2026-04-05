package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

type Signature struct {
	FuncType string
	Params   []Param
}

type Param struct {
	Name string
	Type string
}

var signatures = []Signature{
	// No arguments.
	{FuncType: "func() int32", Params: nil},
	{FuncType: "func() int64", Params: nil},
	{FuncType: "func() uint64", Params: nil},
	{FuncType: "func() uintptr", Params: nil},

	// 1 arguments.
	{FuncType: "func(uintptr) int32", Params: []Param{
		{Name: "DI", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
	}},
	{FuncType: "func(fs.FileDescriptor) int32", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
	}},
	{FuncType: "func(fs.FileDescriptor) uintptr", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
	}},

	// 2 arguments.
	{FuncType: "func(uintptr, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "Cstring"},
	}},
	{FuncType: "func(uint32, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uintptr"},
	}},
	{FuncType: "func(int32, uintptr) int64", Params: []Param{
		{Name: "DI", Type: "int32"},
		{Name: "SI", Type: "uintptr"},
	}},
	{FuncType: "func(uint32, uint32) int64", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uint32"},
	}},
	{FuncType: "func(uint32, uintptr) int32", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uintptr"},
	}},
	{FuncType: "func(Cstring, int64) int32", Params: []Param{
		{Name: "DI", Type: "Cstring"},
		{Name: "SI", Type: "int64"},
	}},
	{FuncType: "func(Cstring, uintptr) int32", Params: []Param{
		{Name: "DI", Type: "Cstring"},
		{Name: "SI", Type: "uintptr"},
	}},
	{FuncType: "func(fs.FileDescriptor, int64) int32", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
		{Name: "SI", Type: "int64"},
	}},
	{FuncType: "func(fs.FileDescriptor, uintptr) int32", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
		{Name: "SI", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uint64) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
	}},
	{FuncType: "func(uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
	}},

	// 3 arguments.
	{FuncType: "func(uint32, uint32, uintptr) int64", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uint32"},
		{Name: "DX", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uintptr, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "Cstring"},
	}},
	{FuncType: "func(Cstring, FileFlags, FileMode) int32", Params: []Param{
		{Name: "DI", Type: "Cstring"},
		{Name: "SI", Type: "FileFlags"},
		{Name: "DX", Type: "FileMode"},
	}},
	{FuncType: "func(uintptr, uint64, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "Cstring"},
	}},
	{FuncType: "func(uintptr, uint64, int32) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
	}},

	// 4 arguments.
	{FuncType: "func(fs.FileDescriptor, uintptr, uint64, int64) int64", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uint64"},
		{Name: "CX", Type: "int64"},
	}},
	{FuncType: "func(fs.FileDescriptor, uint64, uintptr) int32", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "uintptr"},
	}},
	{FuncType: "func(fs.FileDescriptor, uintptr, uint64) int64", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uint64"},
	}},
	{FuncType: "func(fs.FileDescriptor, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "fs.FileDescriptor"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
	}},
	{FuncType: "func(Cstring, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "Cstring"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uint64, int32, int32) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
		{Name: "CX", Type: "int32"},
	}},
	{FuncType: "func(uintptr, uint64, uint32, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "uint32"},
		{Name: "CX", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
	}},

	// 5 arguments.
	{FuncType: "func(uint32, uintptr, uintptr, uintptr, uintptr) int64", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "Cstring"},
	}},
	{FuncType: "func(uintptr, Cstring, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "Cstring"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uint64, int32, int32, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
		{Name: "CX", Type: "int32"},
		{Name: "R8", Type: "Cstring"},
	}},
	{FuncType: "func(uintptr, uint64, uint32, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "uint32"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
	}},

	// 6 arguments.
	{FuncType: "func(uint32, uint32, uintptr, uintptr, uintptr, uintptr) int64", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uint32"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
	}},
	{FuncType: "func(Cstring, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "Cstring"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, Cstring, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "Cstring"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uint64, int32, int32, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
		{Name: "CX", Type: "int32"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uint64, int32, int32, fs.FileDescriptor, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
		{Name: "CX", Type: "int32"},
		{Name: "R8", Type: "fs.FileDescriptor"},
		{Name: "R9", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uintptr, uint64, uint64, int32, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uint64"},
		{Name: "CX", Type: "uint64"},
		{Name: "R8", Type: "int32"},
		{Name: "R9", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
	}},

	// 7 arguments.
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "Cstring"},
	}},
	{FuncType: "func(uintptr, uint64, int32, int32, fs.FileDescriptor, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
		{Name: "CX", Type: "int32"},
		{Name: "R8", Type: "fs.FileDescriptor"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uintptr"},
	}},
	{FuncType: "func(uintptr, uint64, int32, int32, uintptr, uintptr, Cstring) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uint64"},
		{Name: "DX", Type: "int32"},
		{Name: "CX", Type: "int32"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "Cstring"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uintptr"},
	}},

	// 8 arguments.
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uintptr"},
		{Name: "stack1", Type: "uintptr"},
	}},

	// 9 arguments.
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, Cstring, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uintptr"},
		{Name: "stack1", Type: "Cstring"},
		{Name: "stack2", Type: "uintptr"},
	}},
	{FuncType: "func(uint32, uintptr, uintptr, uintptr, uintptr, uint32, uint32, uint32, int64) int64", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uint32"},
		{Name: "stack0", Type: "uint32"},
		{Name: "stack1", Type: "uint32"},
		{Name: "stack2", Type: "int64"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uintptr"},
		{Name: "stack1", Type: "uintptr"},
		{Name: "stack2", Type: "uintptr"},
	}},

	// 10 arguments.
	{FuncType: "func(uint32, uint32, uintptr, uintptr, uintptr, uintptr, uint32, uint32, uint32, int64) int64", Params: []Param{
		{Name: "DI", Type: "uint32"},
		{Name: "SI", Type: "uint32"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uint32"},
		{Name: "stack1", Type: "uint32"},
		{Name: "stack2", Type: "uint32"},
		{Name: "stack3", Type: "int64"},
	}},
	{FuncType: "func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr", Params: []Param{
		{Name: "DI", Type: "uintptr"},
		{Name: "SI", Type: "uintptr"},
		{Name: "DX", Type: "uintptr"},
		{Name: "CX", Type: "uintptr"},
		{Name: "R8", Type: "uintptr"},
		{Name: "R9", Type: "uintptr"},
		{Name: "stack0", Type: "uintptr"},
		{Name: "stack1", Type: "uintptr"},
		{Name: "stack2", Type: "uintptr"},
		{Name: "stack3", Type: "uintptr"},
	}},
}

const templateSrc = `// Code generated by stub_gen.go; DO NOT EDIT.

package elf

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/fs"
)

func CreateDispatcher(goFn any) asm.StubDispatcher {
	switch fn := goFn.(type) {
{{- range $sig := .}}
	case {{$sig.FuncType}}:
		return func(ctx *asm.RegContext) uintptr {
{{- range $i, $p := $sig.Params}}
	{{- if hasPrefix $p.Name "stack"}}
			arg{{add $i 1}} := *(*{{$p.Type}})(unsafe.Add(unsafe.Pointer(ctx), asm.RegContextSize+{{regOffset $i}}))
	{{- end}}
{{- end}}
			return uintptr(fn({{range $i, $p := $sig.Params}}{{if hasPrefix $p.Name "stack"}}arg{{add $i 1}}{{else if eq $p.Type "uintptr"}}ctx.{{$p.Name}}{{else}}{{$p.Type}}(ctx.{{$p.Name}}){{end}}{{if ne $i (len $sig.Params | minusOne)}}, {{end}}{{end}}))		}
{{- end}}
	default:
		panic("unsupported function type")
	}
}
`

func main() {
	genTemplate := template.Must(template.New("gen").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
		"minusOne":  func(n int) int { return n - 1 },
		"add":       func(a, b int) int { return a + b },
		"regOffset": func(i int) int { return (i - 5) * 8 },
	}).Parse(templateSrc))

	var buffer bytes.Buffer
	if err := genTemplate.Execute(&buffer, signatures); err != nil {
		panic(err)
	}
	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: gofmt failed: %v\n", err)
		formatted = buffer.Bytes()
	}
	if err = os.WriteFile("../elf/stub_dispatcher.go", formatted, 0644); err != nil {
		panic(err)
	}
}
