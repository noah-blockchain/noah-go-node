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

###### 1. Download Noah
Clone source code to your machine
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

###### 4. Configure Node Settings

1) Open folder **$HOME/noah/config** and find file **config.toml**
2) Set up db for node. For using **goleveldb** setup parameter **db_backend="goleveldb"**
3) Set up node mode (validator or not validator). For setup node mode we using env variable **VALIDATOR_MODE=(true or false)**.
But if the env **VALIDATOR_MODE** not exist we using parameter from **config.toml** named **validator_mode='(true or false)'**.
Default value **validator_mode='false'**.
4) Setup private node key for generation **Node ID**. By default, node will be generate node key automatically, 
but if you have setup your own node key you can put them in env **NODE_KEY.**

###### 5. Run node
For running node use command **./build/noah node**.
```
noah version
noah node 
```

_We recommend using our node docker._
###### 6. Use GUI
Open http://localhost:3000/ in local browser to see node’s GUI.
P.S. Available only in **not validator** mode.