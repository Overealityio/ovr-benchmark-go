package common

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
	"math/big"
	"os"
)

var (
	W3RpcUrl    = "http://localhost:8545"
	Phrase      = "cigar cat metal speak shoulder tower bread seed unknown ugly carbon sibling"
	PrivateKey  = ""
	SinkAddress = common.HexToAddress("76905f6CE511A0f3410A75e68d689832EDaDf836")
	GasPrice    = big.NewInt(10)
	ChainId     = big.NewInt(9527)
)

func init() {
	if str := os.Getenv("W3RPC_URL"); str != "" {
		W3RpcUrl = str
	}
	if str := os.Getenv("ROOT_PHRASE"); str != "" {
		Phrase = str
	}
	if str := os.Getenv("ROOT_PRIVATE_KEY"); str != "" {
		PrivateKey = str
	}
	if str := os.Getenv("SINK_ACCOUNT"); str != "" {
		SinkAddress = common.HexToAddress(str)
	}
}

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func SaveToFile(text string, filename string) {
	if FileExist(filename) {
		err := os.Remove(filename)
		if err != nil {
			fmt.Printf("Remove: %s\n", err)
			filename = filename + ".temp"
		}
	}
	err := ioutil.WriteFile(filename, []byte(text), 0666)
	if err != nil {
		fmt.Printf("WriteFile: %s\n", err)
	}
}
