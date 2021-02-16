// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"bytes"
	"testing"
)

func TestCopyBytes(t *testing.T) {
	data1 := []byte{1, 2, 3, 4}
	exp1 := []byte{1, 2, 3, 4}
	res1 := CopyBytes(data1)

	if !bytes.Equal(res1, exp1) {
		t.Error("Bytes are not the same")
	}

	if CopyBytes(nil) != nil {
		t.Error("Incorrect result of copy bytes")
	}
}

func TestLeftPadBytes(t *testing.T) {
	val1 := []byte{1, 2, 3, 4}
	exp1 := []byte{0, 0, 0, 0, 1, 2, 3, 4}

	res1 := LeftPadBytes(val1, 8)
	res2 := LeftPadBytes(val1, 2)

	if !bytes.Equal(res1, exp1) || !bytes.Equal(res2, val1) {
		t.Error("Bytes are not the same")
	}
}

func TestRightPadBytes(t *testing.T) {
	val := []byte{1, 2, 3, 4}
	exp := []byte{1, 2, 3, 4, 0, 0, 0, 0}

	resstd := RightPadBytes(val, 8)
	resshrt := RightPadBytes(val, 2)

	if !bytes.Equal(resstd, exp) || !bytes.Equal(resshrt, val) {
		t.Error("Bytes are not the same")
	}
}

func TestFromHex(t *testing.T) {
	input := "NOAHx01"
	expected := []byte{1}
	result := FromHex(input, "NOAHx")
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"", true},
		{"0", false},
		{"00", true},
		{"a9e67e", true},
		{"A9E67E", true},
		{"NOAHxa9e67e", false},
		{"a9e67e001", false},
		{"NOAHxHELLO_MY_NAME_IS_STEVEN_@#$^&*", false},
	}
	for _, test := range tests {
		if ok := isHex(test.input); ok != test.ok {
			t.Errorf("isHex(%q) = %v, want %v", test.input, ok, test.ok)
		}
	}
}

func TestFromHexOddLength(t *testing.T) {
	input := "NOAHx1"
	expected := []byte{1}
	result := FromHex(input, "NOAHx")
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestNoPrefixShortHexOddLength(t *testing.T) {
	input := "1"
	expected := []byte{1}
	result := FromHex(input, "NOAHx")
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestToHex(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if ToHex(b, "NOAHx") != "NOAHx0102030405000000000000000000000000000000" {
		t.Error("Incorrect hex representation")
	}

	if ToHex(nil, "NOAHx") != "NOAHx0" {
		t.Error("Incorrect hex representation")
	}
}

func TestBytes2Hex(t *testing.T) {
	if Bytes2Hex([]byte{1, 2, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) != "0102030405000000000000000000000000000000" {
		t.Error("Incorrect hex representation")
	}
}

func TestHex2BytesFixed(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b2 := Hex2BytesFixed("0102030405000000000000000000000000000000", 20)
	if !bytes.Equal(b2, b) {
		t.Error("Incorrect hex representation")
	}

	b = []byte{0, 1, 2, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b2 = Hex2BytesFixed("0102030405000000000000000000000000000000", 21)
	if !bytes.Equal(b2, b) {
		t.Error("Incorrect hex representation")
	}

	b = []byte{2, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b2 = Hex2BytesFixed("0102030405000000000000000000000000000000", 19)
	if !bytes.Equal(b2, b) {
		t.Error("Incorrect hex representation")
	}
}
