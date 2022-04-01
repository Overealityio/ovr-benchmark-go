all: build

build:
	go build -o ./output/native_init  ./cmd/native_init/main.go
	go build -o ./output/native  ./cmd/native/main.go
	go build -o ./output/erc20_init  ./cmd/erc20_init/main.go
	go build -o ./output/erc20  ./cmd/erc20/main.go

native_init:
	./output/native_init -n 100 -keys ./output/keys.data

native_test:
	./output/native -keys ./output/keys.data -c 10 -d 300

erc20_init:
	./output/erc20_init -n 100 -keys ./output/erc20_keys.data -contract ./output/contract

erc20_test:
	./output/erc20 -keys ./output/erc20_keys.data -contract ./output/contract -c 10 -d 300