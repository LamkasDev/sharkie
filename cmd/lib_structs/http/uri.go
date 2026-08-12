package http

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

type HttpUriElement struct {
	Opaque   bool
	Scheme   Cstring
	Username Cstring
	Password Cstring
	Hostname Cstring
	Path     Cstring
	Query    Cstring
	Fragment Cstring
	Port     uint16
	Reserved [10]byte
}
