### Install From Source

You'll need golang installed https://golang.org/doc/install and the required environment variables set

Clone NOAH source code to your machine

```
mkdir -p $GOPATH/src/github.com/noah-blockchain

cd $GOPATH/src/github.com/noah-blockchain

git clone https://github.com/noah-blockchain/noah-go-node.git

cd noah-blockchain
```

Install Dep [dependency management tool for Go](https://github.com/golang/dep)
```
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
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

The latest NOAH-node version is now installed.

Run NOAH

```
noah
```

Then open http://localhost:3000/ in local browser to see node's GUI.