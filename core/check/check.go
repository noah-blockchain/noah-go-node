package check

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/crypto"
	"github.com/noah-blockchain/noah-go-node/crypto/sha3"
	"github.com/noah-blockchain/noah-go-node/rlp"
	"math/big"
)

var (
	ErrInvalidSig = errors.New("invalid transaction v, r, s values")
)

type Check struct {
	Nonce    []byte
	ChainID  types.ChainID
	DueBlock uint64
	Coin     types.CoinSymbol
	Value    *big.Int
	GasCoin  types.CoinSymbol
	Lock     *big.Int
	V        *big.Int
	R        *big.Int
	S        *big.Int
}

func (check *Check) Sender() (types.Address, error) {
	return recoverPlain(check.Hash(), check.R, check.S, check.V)
}

func (check *Check) LockPubKey() ([]byte, error) {
	sig := check.Lock.Bytes()

	if len(sig) < 65 {
		sig = append(make([]byte, 65-len(sig)), sig...)
	}

	hash := check.HashWithoutLock()

	pub, err := crypto.Ecrecover(hash[:], sig)
	if err != nil {
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}

	return pub, nil
}

// HashWithoutLock returns a types.Hash to be used in process of signing and checking Lock
func (check *Check) HashWithoutLock() types.Hash {
	return rlpHash([]interface{}{
		check.Nonce,
		check.ChainID,
		check.DueBlock,
		check.Coin,
		check.Value,
		check.GasCoin,
	})
}

// Hash returns a types.Hash to be used in process of signing a Check by sender
func (check *Check) Hash() types.Hash {
	return rlpHash([]interface{}{
		check.Nonce,
		check.ChainID,
		check.DueBlock,
		check.Coin,
		check.Value,
		check.GasCoin,
		check.Lock,
	})
}

// Sign signs the check with given private key, returns error
func (check *Check) Sign(prv *ecdsa.PrivateKey) error {
	h := check.Hash()
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return err
	}

	check.setSignature(sig)

	return nil
}

func (check *Check) setSignature(sig []byte) {
	check.R = new(big.Int).SetBytes(sig[:32])
	check.S = new(big.Int).SetBytes(sig[32:64])
	check.V = new(big.Int).SetBytes([]byte{sig[64] + 27})
}

func (check *Check) String() string {
	sender, _ := check.Sender()

	return fmt.Sprintf("Check sender: %s nonce: %x, dueBlock: %d, value: %s %s", sender.String(), check.Nonce,
		check.DueBlock, check.Value.String(), check.Coin.String())
}

// DecodeFromBytes decodes check from bytes
func DecodeFromBytes(buf []byte) (*Check, error) {
	var check Check
	err := rlp.Decode(bytes.NewReader(buf), &check)
	if err != nil {
		return nil, err
	}

	if check.S == nil || check.R == nil || check.V == nil {
		return nil, errors.New("incorrect tx signature")
	}

	return &check, nil
}

func rlpHash(x interface{}) (h types.Hash) {
	hw := sha3.NewLegacyKeccak256()
	err := rlp.Encode(hw, x)
	if err != nil {
		panic(err)
	}
	hw.Sum(h[:0])
	return h
}

func recoverPlain(sighash types.Hash, R, S, Vb *big.Int) (types.Address, error) {
	if Vb.BitLen() > 8 {
		return types.Address{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, true) {
		return types.Address{}, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the snature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return types.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return types.Address{}, errors.New("invalid public key")
	}
	var addr types.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}
