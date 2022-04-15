package common

import (
	"context"
	"fmt"
	"github.com/Overealityio/ovr-benchmark-go/account"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/liyue201/erc20-go/erc20"
	"github.com/shopspring/decimal"
	"github.com/ybbus/jsonrpc/v3"
	"math"
	"math/big"
	"sync"
	"time"
)

func InitParam(web3 *ethclient.Client) {
	height, err := web3.BlockNumber(context.Background())
	if err != nil {
		fmt.Printf("failed to get BlockNumber: %v\n", err)
		panic(err)
	}
	fmt.Printf("height: %d\n", height)

	gasPrice, err := web3.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Printf("failed to gasPrice: %v\n", err)
		panic(err)
	}
	GasPrice = gasPrice
	fmt.Printf("gasPrice: %s\n", gasPrice.String())

	chainID, err := web3.ChainID(context.Background())
	if err != nil {
		fmt.Printf("failed to get ChainID: %v\n", err)
		panic(err)
	}
	ChainId = chainID

	fmt.Printf("chainID: %s\n", chainID.String())
}

func MakeTx(from *account.Account, to *common.Address, nonce uint64, value *big.Int, data []byte) *types.Transaction {
	tx, _ := from.SignTx(types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       to,
		Value:    value,
		Gas:      2100000,
		GasPrice: GasPrice,
		Data:     data,
	}), ChainId)
	return tx
}

func GetBlockTransactionCountByNumber(height uint64) (uint64, error) {
	rpcClient := jsonrpc.NewClient(W3RpcUrl)
	var count hexutil.Uint64
	strHeight := fmt.Sprintf("0x%x", height)
	err := rpcClient.CallFor(context.Background(), &count, "eth_getBlockTransactionCountByNumber", strHeight)

	return uint64(count), err
}

func SendNativeToken(web3 *ethclient.Client, from *account.Account, amount *big.Int) error {
	nonce, err := web3.NonceAt(context.Background(), from.Address(), nil)
	if err != nil {
		fmt.Printf("failed to get nonce: %s\n", err)
		return err
	}
	tx := MakeTx(from, &SinkAddress, nonce, amount, nil)
	err = web3.SendTransaction(context.Background(), tx)
	if err != nil {
		//fmt.Printf("failed to send tx: %s\n", err)
		return err
	}
	//fmt.Printf("send: %v\n", tx.Hash())
	return nil
}

func SendErc20Token(web3 *ethclient.Client, from *account.Account, contract common.Address, amount *big.Int) error {
	token, err := erc20.NewGGToken(contract, web3)
	if err != nil {
		fmt.Printf("Failed to instantiate a Token contract: %v", err)
		return err
	}

	nonce, err := web3.NonceAt(context.Background(), from.Address(), nil)
	if err != nil {
		fmt.Printf("failed to get nonce: %s\n", err)
		return err
	}
	opts := bind.TransactOpts{
		From:     from.Address(),
		Signer:   from.KeySigner(ChainId),
		Nonce:    big.NewInt(int64(nonce)),
		GasLimit: 2100000,
		GasPrice: GasPrice,
	}
	_, err = token.Transfer(&opts, SinkAddress, amount)
	return err
}

func batchSendNativeToken(web3 *ethclient.Client, from *account.Account, accounts []*account.Account, value *big.Int) error {
	for i, account := range accounts {
		time.Sleep(time.Second * 3)

		to := account.Address()
		nonce, err := web3.NonceAt(context.Background(), from.Address(), nil)
		if err != nil {
			fmt.Printf("failed to get nonce: %s\n", err)
			continue
		}
		tx := MakeTx(from, &to, nonce, value, nil)
		err = web3.SendTransaction(context.Background(), tx)
		if err != nil {
			//try again
			//fmt.Printf("faild to send: %s\n", err)
			//tx := MakeTx(from, &to, nonce+1, value, nil)
			//web3.SendTransaction(context.Background(), tx)
		}
		if (i+1)%10 == 0 {
			fmt.Printf("batchSendNativeToken: %d\n", i+1)
		}
	}

	return nil
}

func pickEmptyNativeTokenAccount(web3 *ethclient.Client, accounts []*account.Account) []*account.Account {
	var emptyAccounts []*account.Account
	for _, acc := range accounts {
		balance, err := web3.BalanceAt(context.Background(), acc.Address(), nil)
		if err != nil {
			fmt.Printf("failed to get balance: %v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		if balance.Cmp(big.NewInt(0)) == 0 {
			emptyAccounts = append(emptyAccounts, acc)

		}
	}
	return emptyAccounts
}

func BatchSendNativeToken(web3 *ethclient.Client, from *account.Account, accounts []*account.Account, value *big.Int) error {
	round := 1
	fmt.Printf("BatchSendNativeToken round %d: %v txs to be sent\n", round, len(accounts))
	batchSendNativeToken(web3, from, accounts, value)
	emptyAccounts := accounts
	for {
		time.Sleep(time.Second * 3)
		emptyAccounts = pickEmptyNativeTokenAccount(web3, emptyAccounts)
		if len(emptyAccounts) == 0 {
			break
		}
		round++
		fmt.Printf("BatchSendNativeToken round %d: %v txs to be sent\n", round, len(emptyAccounts))
		batchSendNativeToken(web3, from, emptyAccounts, value)
	}
	return nil
}

func BatchSendNativeToken2(web3 *ethclient.Client, from *account.Account, accounts []*account.Account, value *big.Int) error {
	if len(accounts) > 50 {
		seedCount := int(math.Sqrt(float64(len(accounts))))
		if seedCount > 100 {
			seedCount = 100
		}

		seeds := account.GenAccounts(seedCount)
		c := int64(len(accounts)/seedCount) * 3

		seedValue := big.NewInt(1)
		seedValue.Mul(seedValue, value)
		BatchSendNativeToken(web3, from, seeds, seedValue.Mul(seedValue, big.NewInt(c)))

		segment := len(accounts) / seedCount
		wg := sync.WaitGroup{}
		for i, seed := range seeds {
			wg.Add(1)
			go func(i int, seed *account.Account) {
				defer wg.Done()
				BatchSendNativeToken(web3, seed, accounts[i*segment:(i+1)*segment], value)
			}(i, seed)
		}
		if len(accounts) > segment*seedCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				BatchSendNativeToken(web3, from, accounts[segment*seedCount:], value)
			}()
		}
		wg.Wait()
	} else {
		BatchSendNativeToken(web3, from, accounts, value)
	}
	return nil
}

func batchSendErc20Token(web3 *ethclient.Client, contract common.Address, from *account.Account, accounts []*account.Account, amount *big.Int) error {

	token, err := erc20.NewGGToken(contract, web3)
	if err != nil {
		fmt.Printf("Failed to instantiate a Token contract: %v", err)
		return err
	}
	for i, account := range accounts {
		time.Sleep(time.Second * 3)
		nonce, err := web3.NonceAt(context.Background(), from.Address(), nil)
		if err != nil {
			continue
		}
		to := account.Address()

		opts := bind.TransactOpts{
			From:     from.Address(),
			Signer:   from.KeySigner(ChainId),
			Nonce:    big.NewInt(int64(nonce)),
			GasLimit: 2100000,
			GasPrice: GasPrice,
		}
		_, err = token.Transfer(&opts, to, amount)
		if err != nil {
			//try again
			opts := bind.TransactOpts{
				From:     from.Address(),
				Signer:   from.KeySigner(ChainId),
				Nonce:    big.NewInt(int64(nonce + 1)),
				GasLimit: 2100000,
				GasPrice: GasPrice,
			}
			token.Transfer(&opts, to, amount)
		}

		if (i+1)%10 == 0 {
			fmt.Printf("batchSendErc20Token: %d\n", i+1)
		}
	}
	return nil
}

func pickEmptyErc20TokenAccount(web3 *ethclient.Client, contract common.Address, accounts []*account.Account) []*account.Account {
	token, err := erc20.NewGGToken(contract, web3)
	if err != nil {
		fmt.Printf("Failed to instantiate a Token contract: %v", err)
		return accounts
	}
	var emptyAccounts []*account.Account
	for _, acc := range accounts {
		opts := bind.CallOpts{
			From:    acc.Address(),
			Context: context.Background(),
		}
		balance, err := token.BalanceOf(&opts, acc.Address())
		if err != nil {
			fmt.Printf("failed to get balance: %v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		if balance.Cmp(big.NewInt(0)) == 0 {
			emptyAccounts = append(emptyAccounts, acc)

		}
	}
	return emptyAccounts
}

func BatchSendErc20Token(web3 *ethclient.Client, contract common.Address, from *account.Account, accounts []*account.Account, amount *big.Int) error {
	round := 1
	fmt.Printf("BatchSendErc20Token round %d: %v txs to be sent\n", round, len(accounts))

	batchSendErc20Token(web3, contract, from, accounts, amount)
	emptyAccounts := accounts
	for {
		time.Sleep(time.Second * 3)
		emptyAccounts = pickEmptyErc20TokenAccount(web3, contract, emptyAccounts)
		if len(emptyAccounts) == 0 {
			break
		}
		round++
		fmt.Printf("BatchSendErc20Token round %d: %v txs to be sent\n", round, len(emptyAccounts))
		batchSendErc20Token(web3, contract, from, emptyAccounts, amount)
	}
	return nil
}

func BatchSendErc20Token2(web3 *ethclient.Client, contract common.Address, from *account.Account, accounts []*account.Account, value *big.Int) error {
	if len(accounts) > 50 {
		seedCount := int(math.Sqrt(float64(len(accounts))))
		if seedCount > 100 {
			seedCount = 100
		}
		seeds := account.GenAccounts(seedCount)
		c := int64(len(accounts)/seedCount) * 3

		seedValue := big.NewInt(1)
		seedValue.Mul(seedValue, value)

		//send gas
		BatchSendNativeToken(web3, from, seeds, decimal.New(10, 18).BigInt())

		//send token
		BatchSendErc20Token(web3, contract, from, seeds, seedValue.Mul(seedValue, big.NewInt(c)))

		segment := len(accounts) / seedCount
		wg := sync.WaitGroup{}
		for i, seed := range seeds {
			wg.Add(1)
			go func(i int, seed *account.Account) {
				defer wg.Done()
				BatchSendErc20Token(web3, contract, seed, accounts[i*segment:(i+1)*segment], value)
			}(i, seed)
		}
		if len(accounts) > segment*seedCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				BatchSendErc20Token(web3, contract, from, accounts[segment*seedCount:], value)
			}()
		}
		wg.Wait()
	} else {
		BatchSendErc20Token(web3, contract, from, accounts, value)
	}
	return nil
}
