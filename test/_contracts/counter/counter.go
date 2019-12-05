package main

import (
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/state"
)

var PUBLIC = sdk.Export(inc, value)
var SYSTEM = sdk.Export(_init)

func _init() {

}

func inc() uint64 {
	val := value() + 1
	state.WriteUint64([]byte("counter"), val)
	return val
}

func value() uint64 {
	return state.ReadUint64([]byte("counter"))
}
