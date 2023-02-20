package main

import (
	"log"

	"github.com/gofaith/go2cpp"
)

func main() {
	e := go2cpp.ConvertFile("out.cpp", "hello.go1")
	if e != nil {
		log.Println(e)
		return
	}
}
