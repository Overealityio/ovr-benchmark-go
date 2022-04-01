package account

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"math/big"
)

type Account struct {
	prvKey *ecdsa.PrivateKey
	signer crypto.Signer
}

func NewAccount(prvKey *ecdsa.PrivateKey) *Account {
	ac := Account{
		prvKey: prvKey,
		signer: crypto.NewDefaultSigner(prvKey),
	}
	return &ac
}

func NewAccountFromPrvHex(prvKeyHex string) (*Account, error) {
	data, err := hex.DecodeString(prvKeyHex)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.DecodeSecp256k1PrivateKey(data)
	if err != nil {
		return nil, err
	}
	return NewAccount(privKey), nil
}

func NewFromPhrase(mnemonic string) (*Account, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}
	prv, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}
	return NewAccount(prv), nil
}

func (ac *Account) Address() common.Address {
	addr, _ := ac.signer.EthereumAddress()
	return addr
}

func (ac *Account) SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return ac.signer.SignTx(transaction, chainID)
}

func (ac *Account) PrvHex() string {
	return hex.EncodeToString(crypto.EncodeSecp256k1PrivateKey(ac.prvKey))
}

func (ac *Account) KeySigner(chainID *big.Int) bind.SignerFn {
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		keyAddr := crypto2.PubkeyToAddress(ac.prvKey.PublicKey)
		if address != keyAddr {
			return nil, errors.New("not authorized to sign this account")
		}
		return types.SignTx(tx, types.NewEIP155Signer(chainID), ac.prvKey)
	}
}

func GenAccount() *Account {
	for {
		key, err := crypto.GenerateSecp256k1Key()
		if err != nil {
			continue
		}
		return NewAccount(key)
	}
}

func GenAccounts(n int) []*Account {
	var list []*Account
	for i := 0; i < n; i++ {
		list = append(list, GenAccount())
	}
	return list
}
