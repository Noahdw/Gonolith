package main

import (
	"github.com/noahdw/Gonolith/internal/node"
)

func main() {
	gonolith := node.NewGonolith()
	gonolith.Serve()
}
