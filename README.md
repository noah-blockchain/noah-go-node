<p align="center" style="text-align: center;">
    <a href="https://github.com/noah-blockchain/noah-go-node/blob/master/LICENSE">
        <img src="https://img.shields.io/packagist/l/doctrine/orm.svg" alt="License">
    </a>
    <img alt="undefined" src="https://img.shields.io/github/last-commit/noah-blockchain/noah-go-node.svg">

[![version](https://img.shields.io/github/tag/noah-blockchain/noah-go-node.svg)](https://github.com/tendermint/tendermint/releases/latest)
[![Go version](https://img.shields.io/badge/go-1.12.0-blue.svg)](https://github.com/moovweb/gvm)
[![](https://tokei.rs/b1/github/noah-blockchain/noah-go-node?category=lines)](https://github.com/noah-blockchain/noah-go-node)
</p>

#NOAH-blockchain go-node



### [dev-branch](https://github.com/noah-blockchain/noah-go-node/tree/dev)
The branch contains the most current version

#### [alpha-branch](https://github.com/noah-blockchain/noah-go-node/tree/alpha)
The branch contains a version for alpha-testing

#### [beta-branch](https://github.com/noah-blockchain/noah-go-node/tree/beta)
The branch contains a version for beta-testing

#### [master-branch](https://github.com/noah-blockchain/noah-go-node/tree/master)
Public release

## Sub-modules

####[Remote cluster using terraform and ansible](https://github.com/tendermint/tendermint/blob/master/docs/networks/terraform-and-ansible.md)

####[Amino](https://github.com/tendermint/go-amino)

####[IAVL+ Tree](https://github.com/tendermint/iavl)

## 1. Install and build  node
```
go mod vendor
make build
```
After this command compiled node will be in folder build.

## 2. Configuration

make file config.toml
```
./config/config.toml
```

## 3. Init and start
```
noah version
noah node 
```



