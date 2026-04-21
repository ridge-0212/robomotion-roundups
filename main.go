package main

import (
	"roundups/v1"

	"github.com/robomotionio/robomotion-go/runtime"
)

func main() {
	runtime.RegisterNodes(
		&v1.CreateRoundup{},
		&v1.FetchRoundup{},
	)
	runtime.Start()
}
