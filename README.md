[![last commit](https://img.shields.io/github/last-commit/noah-blockchain/noah-go-node.svg)]()
[![license](https://img.shields.io/packagist/l/doctrine/orm.svg)](https://github.com/noah-blockchain/noah-go-node/blob/master/LICENSE)
[![version](https://img.shields.io/github/tag/noah-blockchain/noah-go-node.svg)](https://github.com/noah-blockchain/noah-go-node/releases/latest)
[![Go version](https://img.shields.io/badge/go-1.12.0-blue.svg)](https://github.com/moovweb/gvm)
[![](https://tokei.rs/b1/github/noah-blockchain/noah-go-node?category=lines)](https://github.com/noah-blockchain/noah-go-node)

# NOAH-blockchain go-node

## Sub-modules

#### [1) Remote cluster using terraform and ansible](https://github.com/tendermint/tendermint/blob/master/docs/networks/terraform-and-ansible.md)
#### [2) Amino](https://github.com/tendermint/go-amino)
#### [3) IAVL+ Tree](https://github.com/tendermint/iavl)

###[Guide how to configure and delegate your validator](https://docs.google.com/document/d/19sZeIFy6aE8xuPg1-mq-0Cah2fiyj-BpyZCSQXEZKFc/edit)

## Quick Installation from Docker

1) Pull docker from official docker hub

```
docker pull noahblockchain/node
```
2) Run you validator for initialization node
```
docker run -p 26656:26656 -p 26657:26657 -p 26660:26660 -p 8841:8841 -p 3000:3000 -v ~/node:/root/noah/ noahblockchain/node noah node --network-id=noah-mainnet-1 --chain-id=mainnet --validator-mode
```

--network-id=X, where X its choose network (noah-mainnet-1 or noah-testnet-1)

--chain-id=Y, where Y its choose chain (mainnet or testnet)

--validator-mode if node working in Validator mode

## Starting validator from source code

###### 1. Clone source code to your machine
```
mkdir -p $GOPATH/src/github.com/noah-blockchain or $HOME/noah
cd $GOPATH/src/github.com/noah-blockchain
git clone https://github.com/noah-blockchain/noah-go-node.git
cd noah-go-node
```

###### 2. Install Node Modules
```
make create_vendor
```

###### 3. Compile
```
make build
```
After this command compiled node will be in folder build and node configuration will be in folder **$HOME/noah.**

###### 4. Run node
For running validator use command 
```
./build/noah node --network-id=noah-mainnet-1 --chain-id=mainnet --validator-mode
```
We recommend using our official node docker.
###### 5. Use GUI
Open http://localhost:3000/ in local browser to see node’s GUI.

P.S. Available only in NOT validator mode.