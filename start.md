### Install From Source

You'll need golang installed https://golang.org/doc/install and the required environment variables set

Clone NOAX source code to your machine

```
mkdir -p $GOPATH/src/github.com/noax

cd $GOPATH/src/github.com/noax

git clone https://bitbucket.org/amm-core-devteam/noah-blockchain.git

cd noah-blockchain
```

Get Tools & Dependencies
```
make get_tools

make get_vendor_deps
```

Compile
```
make install
```

to put the binary in $GOPATH/bin or use:
```
make build
```

to put the binary in ./build.

The latest NOAX-node version is now installed.

Run NOAX

```
noax
```

Then open http://localhost:3000/ in local browser to see node's GUI.