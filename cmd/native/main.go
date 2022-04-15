package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Overealityio/ovr-benchmark-go/account"
	common2 "github.com/Overealityio/ovr-benchmark-go/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	totalTxSend        uint64 = 0
	totalTxSendSuccess uint64 = 0
	totalTxInBlock     uint64 = 0
)

func loadAccount(keysFile string) []*account.Account {
	fileData, err := ioutil.ReadFile(keysFile)
	if err != nil {
		panic(err)
	}
	keysHex := strings.Split(string(fileData), "\n")
	var accounts []*account.Account
	for _, v := range keysHex {
		ac, err := account.NewAccountFromPrvHex(v)
		if err != nil {
			fmt.Printf("failed to parse account: %v, %v\n", v, err)
			continue
		}
		accounts = append(accounts, ac)
	}
	return accounts
}

func stats(ctx context.Context) {
	web3, err := ethclient.Dial(common2.W3RpcUrl)
	if err != nil {
		fmt.Printf("failed to dial: %v\n", err)
		return
	}
	preHeight, err := web3.BlockNumber(ctx)
	if err != nil {
		fmt.Printf("failed to get height: %v\n", err)
		return
	}
	fmt.Printf("startHeight: %v\n", preHeight)

	startTime := int64(0)

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			height, err := web3.BlockNumber(ctx)
			if err != nil {
				fmt.Printf("failed to get height: %v\n", err)
				continue
			}
			if height == preHeight {
				continue
			}
			preHeight = height
			count, err := common2.GetBlockTransactionCountByNumber(height)
			if err != nil {
				fmt.Printf("failed to get BlockByNumber %v: %v\n", height, err)
				continue
			}
			if startTime == 0 && count > 0 {
				startTime = time.Now().Unix() - 2
			}
			atomic.AddUint64(&totalTxInBlock, count)

			fmt.Println("#############################################")
			fmt.Printf("blockHeight:           %v\n", height)
			fmt.Printf("txInThisBlock:         %v\n", count)
			fmt.Printf("totalTxInBlock:        %v\n", atomic.LoadUint64(&totalTxInBlock))
			fmt.Printf("totalTxSend:           %v\n", atomic.LoadUint64(&totalTxSend))
			fmt.Printf("totalTxSendSuccess:    %v\n", atomic.LoadUint64(&totalTxSendSuccess))
			if startTime > 0 {
				d := time.Now().Unix() - startTime
				fmt.Printf("time:                     %v\n", d)
				fmt.Printf("TPS(totalTxInBlock/time): %0.1f\n", float32(atomic.LoadUint64(&totalTxInBlock))/float32(d))
			}
			fmt.Println("#############################################")
		}
	}
}

func run(ctx context.Context, web3 *ethclient.Client, accounts []*account.Account, concurrency int) {
	wg := sync.WaitGroup{}
	accountChan := make(chan *account.Account, len(accounts))
	for _, ac := range accounts {
		accountChan <- ac
	}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case account := <-accountChan:
					err := common2.SendNativeToken(web3, account, big.NewInt(1))
					atomic.AddUint64(&totalTxSend, 1)
					if err == nil {
						atomic.AddUint64(&totalTxSendSuccess, 1)
					} else {
						time.Sleep(time.Second)
					}
					accountChan <- account
				}
			}
		}()
	}
	wg.Wait()
}

func main() {
	keysFile := flag.String("keys", "keys.data", "the file path of private keys")
	concurrency := flag.Int("c", 1, "The number of concurrent requests to send")
	duration := flag.Int("d", 60, "Test duration in seconds")
	flag.Parse()

	web3, err := ethclient.Dial(common2.W3RpcUrl)
	if err != nil {
		fmt.Printf("failed to dial: %v\n", err)
		return
	}
	common2.InitParam(web3)

	accounts := loadAccount(*keysFile)
	fmt.Printf("account count: %v\n", len(accounts))

	ctx, cancel := context.WithCancel(context.Background())
	go run(ctx, web3, accounts, *concurrency)
	go stats(ctx)

	ticker := time.NewTicker(time.Second * time.Duration(*duration))

LOOP:
	for {
		select {
		case <-ticker.C:
			cancel()
			break LOOP
		default:
			time.Sleep(time.Second)
		}
	}

}
