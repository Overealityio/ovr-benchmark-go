package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/Overealityio/ovr-benchmark-go/account"
	common2 "github.com/Overealityio/ovr-benchmark-go/common"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math/big"
	"os"
)

func saveAccounts(accounts []*account.Account, filename string) {
	var text string
	for i, ac := range accounts {
		if i > 0 {
			text += "\n"
		}
		text += ac.PrvHex()
	}
	if common2.FileExist(filename) {
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

func prepareAccounts(web3 *ethclient.Client, origin *account.Account, n int, filename string) {
	accounts := account.GenAccounts(n)
	err := common2.BatchSendNativeToken2(web3, origin, accounts, decimal.New(100, 18).BigInt())
	if err != nil {
		return
	}
	fmt.Printf("total accounts: %d\n", len(accounts))
	if len(accounts) > 0 {
		saveAccounts(accounts, filename)
	}
}

func main() {
	n := flag.Int("n", 10, "number of account to generated")
	keysFile := flag.String("keys", "keys.data", "the file path to save private keys")
	flag.Parse()

	var rootAccount *account.Account
	if common2.PrivateKey != "" {
		account, err := account.NewAccountFromPrvHex(common2.PrivateKey)
		if err != nil {
			panic(err)
		}
		rootAccount = account
	} else {
		account, err := account.NewFromPhrase(common2.Phrase)
		if err != nil {
			panic(err)
		}
		rootAccount = account
	}

	web3, err := ethclient.Dial(common2.W3RpcUrl)
	if err != nil {
		fmt.Printf("failed to dial: %v\n", err)
		return
	}
	height, err := web3.BlockNumber(context.Background())
	if err != nil {
		fmt.Printf("failed to get BlockNumber: %v\n", err)
		return
	}
	fmt.Printf("height: %d\n", height)

	balance, err := web3.BalanceAt(context.Background(), rootAccount.Address(), nil)
	if err != nil {
		fmt.Printf("failed to get BlockNumber: %v\n", err)
		return
	}
	fmt.Printf("balance of %s: %s\n", rootAccount.Address(), balance.String())
	if balance.Cmp(big.NewInt(0)) == 0 {
		fmt.Println("No balance!!!!!!!!!!!!!!!!!!!!!!!")
		return
	}
	prepareAccounts(web3, rootAccount, *n, *keysFile)
}
