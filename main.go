package main

import (
	"github.com/robomotionio/robomotion-go/runtime"
	"github.com/robomotionio/robomotion-roundups/v1"
)

func main() {
	runtime.RegisterNodes(
		&v1.CreateRoundup{},
		&v1.FetchRoundup{},
	)
	runtime.Start()
}
