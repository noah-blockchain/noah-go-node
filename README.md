[![last commit](https://img.shields.io/github/last-commit/noah-blockchain/noah-go-node.svg)]()
[![license](https://img.shields.io/packagist/l/doctrine/orm.svg)](https://github.com/noah-blockchain/noah-go-node/blob/master/LICENSE)
[![version](https://img.shields.io/github/tag/noah-blockchain/noah-go-node.svg)](https://github.com/noah-blockchain/noah-go-node/releases/latest)
[![Go version](https://img.shields.io/badge/go-1.12.0-blue.svg)](https://github.com/moovweb/gvm)
[![](https://tokei.rs/b1/github/noah-blockchain/noah-go-node?category=lines)](https://github.com/noah-blockchain/noah-go-node)

# NOAH-blockchain go-node

### [dev-branch](https://github.com/noah-blockchain/noah-go-node/tree/dev)
The branch contains the most current version

#### [alpha-branch](https://github.com/noah-blockchain/noah-go-node/tree/alpha)
The branch contains a version for alpha-testing

#### [beta-branch](https://github.com/noah-blockchain/noah-go-node/tree/beta)
The branch contains a version for beta-testing

#### [master-branch](https://github.com/noah-blockchain/noah-go-node/tree/master)
Public release

## Sub-modules

#### [Remote cluster using terraform and ansible](https://github.com/tendermint/tendermint/blob/master/docs/networks/terraform-and-ansible.md)

#### [Amino](https://github.com/tendermint/go-amino)

#### [IAVL+ Tree](https://github.com/tendermint/iavl)

##  Install and build  node

###### 1. [Install MeMDB](https://github.com/hashicorp/go-memdb) 
Node will be working with memdb.
   For using this db is necessary to fix file config.toml and change **db_backend = "memdb"**
   
###### 2. Download Noah
Clone source code to your machine
```
mkdir -p $GOPATH/src/github.com/noah-blockchain or $HOME/noah
cd $GOPATH/src/github.com/noah-blockchain
git clone https://github.com/noah-blockchain/noah-go-node.git
cd noah-go-node
```

Get Tools & Dependencies
```
make create_vendor
```
Compile

```
make build
```
After this command compiled node will be in folder build.

###### 3. Run Noah Node
make file config.toml
Change config.toml.example to config.toml and push in $HOME/noah/config folder

```
./config/config.toml
```
```
noah version
noah node 
```
###### 4. Use GUI
Open http://localhost:3000/ in local browser to see node’s GUI.
