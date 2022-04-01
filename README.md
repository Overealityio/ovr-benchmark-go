# ovr-benchmark

## Env

LocalNet(Default)
```
export W3RPC_URL=http://localhost:8545`
export ROOT_PHRASE="cigar cat metal speak shoulder tower bread seed unknown ugly carbon sibling"
```

DevNet
```
export W3RPC_URL=https://dev-qa01.dev.overeality.io:8545
export ROOT_PHRASE="xxxxxxx"`
```

## Build
```
make 
```

## Test native token transfers

Create a batch of accounts and transfer a fixed number of tokens
```
make native_init
```
Test tps for transferring native tokens
```
make native_test
```

## Test ERC20 token transfers

Create an erc20 contract and transfer a fixed number of tokens
```
make erc20_init
```
Test tps for transferring erc20 tokens
```
make erc20_test
```

### Command-Line Usage

`./output/native_init  -h`
```
Usage of ./output/native_init:
  -keys string
    	the file path to save private keys (default "keys.data")
  -n int
    	number of account to generated (default 10)

```

`./output/native -h`
```
Usage of ./output/native:
  -c int
    	The number of concurrent requests to send (default 1)
  -d int
    	Test duration in seconds (default 60)
  -keys string
    	the file path of private keys (default "keys.data")

```

`./output/erc20_init -h`
```
Usage of ./output/erc20_init:
  -contract string
    	the file path to save contract address (default "contract.data")
  -keys string
    	the file path to save private keys (default "keys.data")
  -n int
    	number of account to generated (default 10)

```

`./output/erc20 -h`
```
Usage of ./output/erc20:
  -c int
    	The number of concurrent requests to send (default 1)
  -contract string
    	the file path to save contract address (default "contract.data")
  -d int
    	Test duration in seconds (default 60)
  -keys string
    	the file path of private keys (default "keys.data")
```