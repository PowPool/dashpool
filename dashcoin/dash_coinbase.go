package dashcoin

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/mutalisk999/bitcoin-lib/src/base58"
	"github.com/mutalisk999/bitcoin-lib/src/keyid"
	"github.com/mutalisk999/bitcoin-lib/src/pubkey"
	"github.com/mutalisk999/bitcoin-lib/src/script"
	"github.com/mutalisk999/bitcoin-lib/src/serialize"
	"github.com/mutalisk999/bitcoin-lib/src/utility"
	"io"
)

func GetCoinBaseScriptByPubKey(pubKeyHex string) ([]byte, error) {
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, errors.New("invalid pubKeyHex")
	}
	if pubKey[0] != '\x02' && pubKey[0] != '\x03' {
		return nil, errors.New("invalid pubKeyHex")
	}

	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)

	var pubkey pubkey.PubKey
	err = pubkey.SetPubKeyData(pubKey)
	if err != nil {
		return nil, err
	}
	err = pubkey.Pack(bufWriter)
	if err != nil {
		return nil, errors.New("pack pubKeyHex err")
	}
	err = serialize.PackByte(bufWriter, script.OP_CHECKSIG)
	if err != nil {
		return nil, errors.New("pack byte err")
	}
	return bytesBuf.Bytes(), nil
}

func GetCoinBaseScriptByAddress(address string) ([]byte, error) {
	addrWithCheck, err := base58.Decode(address)
	if err != nil {
		return nil, errors.New("invalid address")
	}
	if len(addrWithCheck) != 25 {
		return nil, errors.New("invalid address")
	}
	check1 := utility.Sha256(utility.Sha256(addrWithCheck[0:21]))[0:4]
	check2 := addrWithCheck[21:25]
	if bytes.Compare(check1, check2) != 0 {
		return nil, errors.New("invalid address")
	}

	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)

	err = serialize.PackByte(bufWriter, script.OP_DUP)
	if err != nil {
		return nil, errors.New("pack byte err")
	}
	err = serialize.PackByte(bufWriter, script.OP_HASH160)
	if err != nil {
		return nil, errors.New("pack byte err")
	}
	var addr keyid.KeyID
	err = addr.SetKeyIDData(addrWithCheck[1:21])
	if err != nil {
		return nil, err
	}
	err = addr.Pack(bufWriter)
	if err != nil {
		return nil, errors.New("pack address err")
	}
	err = serialize.PackByte(bufWriter, script.OP_EQUALVERIFY)
	if err != nil {
		return nil, errors.New("pack byte err")
	}
	err = serialize.PackByte(bufWriter, script.OP_CHECKSIG)
	if err != nil {
		return nil, errors.New("pack byte err")
	}

	return bytesBuf.Bytes(), nil
}

func GetCoinBaseScript(wallet string) ([]byte, error) {
	if len(wallet) == 66 {
		return GetCoinBaseScriptByPubKey(wallet)
	} else {
		return GetCoinBaseScriptByAddress(wallet)
	}
}

func GetCoinBaseScriptHex(wallet string) (string, error) {
	scriptHex, err := GetCoinBaseScript(wallet)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(scriptHex), nil
}

type DashCoinBaseTransaction struct {
}
