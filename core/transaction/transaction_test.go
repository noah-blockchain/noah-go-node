package transaction

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestDecode3(t *testing.T) {
	tx := "0xf8901301018a4e4f414800000000000007b6f5a0028970ceef8bb76e1471ac4e6a5fd52958b624133c52b3bd13f773f262b195d08a4e4f414800000000000088d02ab486cedc0000808001b845f8431ba0a2791ef1845fb42a2abc8493aa3a90b6a4b3eb4790786d5fbca79bc14cc0aa87a01b16d41006828632480354064dcb209f2850b9033b67421eb06ccf190717242c"
	if !strings.HasPrefix(tx, "0x") {
		t.Fail()
	}

	decodeString, err := hex.DecodeString(tx[2:])
	if err != nil {
		t.Fail()
	}

	decodedTx, err := TxDecoder.DecodeFromBytes(decodeString)

	if err != nil {
		t.Fatal(err)
	}
	data := decodedTx.decodedData.(*DelegateData)
	fmt.Println(data.Value.String())
	fmt.Println(data.Coin)
	fmt.Println(data.String())
	fmt.Println(decodedTx.decodedData)

}
