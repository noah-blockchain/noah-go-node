<p align="center" style="text-align: center;">
    <a href="https://github.com/noah-blockchain/noah-go-node/blob/master/LICENSE">
        <img src="https://img.shields.io/packagist/l/doctrine/orm.svg" alt="License">
    </a>
    <img alt="undefined" src="https://img.shields.io/github/last-commit/noah-blockchain/noah-go-node.svg">
</p>

#NOAH-blockchain go-node

### [dev-branch](https://github.com/noah-blockchain/noah-go-node/tree/dev)
The branch contains the most current version

#### [alpha-branch](https://github.com/noah-blockchain/noah-go-node/tree/alpha)
The branch contains a version for alpha-testing

#### [beta-branch](https://github.com/noah-blockchain/noah-go-node/tree/beta)
The branch contains a version for beta-testing

#### [testnet-branch](https://github.com/noah-blockchain/noah-go-node/tree/testnet)
Implementation for test network

#### [master-branch](https://github.com/noah-blockchain/noah-go-node/tree/master)
Public release

## Sub-modules

####[Remote cluster using terraform and ansible](https://github.com/tendermint/tendermint/blob/master/docs/networks/terraform-and-ansible.md)

####[Amino](https://github.com/tendermint/go-amino)

####[IAVL+ Tree](https://github.com/tendermint/iavl)

## How to install node
Working folder for node - $HOME/noah
1) Change config.toml.example to config.toml and push in $HOME/noah/config folder
2) **make create_vendor** - for getting all dependencies
3) **make build** - create build.

<br>Testing - _./build/noah version_
<br>Start node - _./build/noah node_ 

