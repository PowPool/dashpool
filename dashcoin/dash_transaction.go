package dashcoin

import (
	"bytes"
	"encoding/hex"
	"github.com/mutalisk999/bitcoin-lib/src/bigint"
	"github.com/mutalisk999/bitcoin-lib/src/blob"
	"github.com/mutalisk999/bitcoin-lib/src/script"
	"github.com/mutalisk999/bitcoin-lib/src/serialize"
	"github.com/mutalisk999/bitcoin-lib/src/utility"
	"io"
	"strings"
)

const (
	TRANSACTION_NORMAL                    = 0
	TRANSACTION_PROVIDER_REGISTER         = 1
	TRANSACTION_PROVIDER_UPDATE_SERVICE   = 2
	TRANSACTION_PROVIDER_UPDATE_REGISTRAR = 3
	TRANSACTION_PROVIDER_UPDATE_REVOKE    = 4
	TRANSACTION_COINBASE                  = 5
	TRANSACTION_QUORUM_COMMITMENT         = 6
)

type OutPoint struct {
	Hash bigint.Uint256
	N    uint32
}

func (o OutPoint) Pack(writer io.Writer) error {
	err := o.Hash.Pack(writer)
	if err != nil {
		return err
	}
	err = serialize.PackUint32(writer, o.N)
	if err != nil {
		return err
	}
	return nil
}

func (o *OutPoint) UnPack(reader io.Reader) error {
	err := o.Hash.UnPack(reader)
	if err != nil {
		return err
	}
	o.N, err = serialize.UnPackUint32(reader)
	if err != nil {
		return err
	}
	return nil
}

type OutPointPrintAble struct {
	Hash string
	N    uint32
}

type TxIn struct {
	PrevOut   OutPoint
	ScriptSig script.Script
	Sequence  uint32
}

func (t TxIn) Pack(writer io.Writer) error {
	err := t.PrevOut.Pack(writer)
	if err != nil {
		return err
	}
	err = t.ScriptSig.Pack(writer)
	if err != nil {
		return err
	}
	err = serialize.PackUint32(writer, t.Sequence)
	if err != nil {
		return err
	}
	return nil
}

func (t *TxIn) UnPack(reader io.Reader) error {
	err := t.PrevOut.UnPack(reader)
	if err != nil {
		return err
	}
	err = t.ScriptSig.UnPack(reader)
	if err != nil {
		return err
	}
	t.Sequence, err = serialize.UnPackUint32(reader)
	if err != nil {
		return err
	}
	return nil
}

type TxInPrintAble struct {
	PrevOut   OutPointPrintAble
	ScriptSig string
	Sequence  uint32
}

type TxOut struct {
	Value        int64
	ScriptPubKey script.Script
}

func (t TxOut) Pack(writer io.Writer) error {
	err := serialize.PackInt64(writer, t.Value)
	if err != nil {
		return err
	}
	err = t.ScriptPubKey.Pack(writer)
	if err != nil {
		return err
	}
	return nil
}

func (t *TxOut) UnPack(reader io.Reader) error {
	var err error
	t.Value, err = serialize.UnPackInt64(reader)
	if err != nil {
		return err
	}
	err = t.ScriptPubKey.UnPack(reader)
	if err != nil {
		return err
	}
	return nil
}

type TxOutPrintAble struct {
	Value        int64
	ScriptPubKey string
	Address      string
	ScriptType   string
}

type DashTransaction struct {
	Vin          []TxIn
	Vout         []TxOut
	Version      int32
	LockTime     uint32
	ExtraPayload script.Script
	Version16    int16
	Type16       int16
}

func (t DashTransaction) packVin(writer io.Writer, vin *[]TxIn) error {
	err := serialize.PackCompactSize(writer, uint64(len(*vin)))
	if err != nil {
		return err
	}
	for _, v := range *vin {
		err = v.Pack(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t DashTransaction) packVout(writer io.Writer, vout *[]TxOut) error {
	err := serialize.PackCompactSize(writer, uint64(len(*vout)))
	if err != nil {
		return err
	}
	for _, v := range *vout {
		err = v.Pack(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t DashTransaction) Pack(writer io.Writer) error {
	t.Version = int32(t.Type16)<<16 | int32(t.Version16)
	err := serialize.PackInt32(writer, t.Version)
	if err != nil {
		return err
	}
	// pack Vin
	err = t.packVin(writer, &t.Vin)
	if err != nil {
		return err
	}
	// pack Vout
	err = t.packVout(writer, &t.Vout)
	if err != nil {
		return err
	}
	err = serialize.PackUint32(writer, t.LockTime)
	if err != nil {
		return err
	}
	if t.Version16 == 3 && t.Type16 != TRANSACTION_NORMAL {
		err = t.ExtraPayload.Pack(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t DashTransaction) PackToHex() (string, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err := t.Pack(bufWriter)
	if err != nil {
		return "", err
	}
	var Blob blob.Byteblob
	Blob.SetData(bytesBuf.Bytes())
	return Blob.GetHex(), nil
}

func (t DashTransaction) CalcTrxId() (bigint.Uint256, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err := t.Pack(bufWriter)
	if err != nil {
		return bigint.Uint256{}, err
	}
	bytesHash := utility.Sha256(utility.Sha256(bytesBuf.Bytes()))
	var ui256 bigint.Uint256
	_ = ui256.SetData(bytesHash)
	return ui256, nil
}

func (t *DashTransaction) unpackVin(reader io.Reader) (*[]TxIn, error) {
	var vin []TxIn
	ui64, err := serialize.UnPackCompactSize(reader)
	if err != nil {
		return nil, err
	}
	vin = make([]TxIn, ui64, ui64)
	for i := 0; i < int(ui64); i++ {
		var v TxIn
		err = v.UnPack(reader)
		if err != nil {
			return nil, err
		}
		vin[i] = v
	}
	return &vin, nil
}

func (t *DashTransaction) unpackVout(reader io.Reader) (*[]TxOut, error) {
	var vout []TxOut
	ui64, err := serialize.UnPackCompactSize(reader)
	if err != nil {
		return nil, err
	}
	vout = make([]TxOut, ui64, ui64)
	for i := 0; i < int(ui64); i++ {
		var v TxOut
		err = v.UnPack(reader)
		if err != nil {
			return nil, err
		}
		vout[i] = v
	}
	return &vout, nil
}

func (t *DashTransaction) UnPack(reader io.Reader) error {
	var err error
	var vin *[]TxIn
	var vout *[]TxOut
	t.Version, err = serialize.UnPackInt32(reader)
	if err != nil {
		return err
	}
	t.Version16 = int16(t.Version & 0xFFFF)
	t.Type16 = int16(t.Version >> 16)
	// unpack Vin
	vin, err = t.unpackVin(reader)
	if err != nil {
		return err
	}
	t.Vin = *vin

	// unpack Vout
	vout, err = t.unpackVout(reader)
	if err != nil {
		return err
	}
	t.Vout = *vout

	t.LockTime, err = serialize.UnPackUint32(reader)
	if err != nil {
		return err
	}

	if t.Version16 == 3 && t.Type16 != TRANSACTION_NORMAL {
		// unpack ExtraPayload
		err = t.ExtraPayload.UnPack(reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *DashTransaction) UnPackFromHex(hexStr string) error {
	var Blob blob.Byteblob
	err := Blob.SetHex(hexStr)
	if err != nil {
		return err
	}
	bytesBuf := bytes.NewBuffer(Blob.GetData())
	bufReader := io.Reader(bytesBuf)
	err = t.UnPack(bufReader)
	if err != nil {
		return err
	}
	return nil
}

type TrxPrintAble struct {
	Vin          []TxInPrintAble
	Vout         []TxOutPrintAble
	Version      int32
	LockTime     uint32
	ExtraPayload string
	Version16    int16
	Type16       int16
}

func (t *DashTransaction) GetTrxPrintAble() TrxPrintAble {
	var trxPrintAble TrxPrintAble
	trxPrintAble.Vin = make([]TxInPrintAble, len(t.Vin), len(t.Vin))
	trxPrintAble.Vout = make([]TxOutPrintAble, len(t.Vout), len(t.Vout))

	for i := 0; i < len(t.Vin); i++ {
		var vinPrintAble TxInPrintAble
		vinPrintAble.PrevOut.Hash = t.Vin[i].PrevOut.Hash.GetHex()
		vinPrintAble.PrevOut.N = t.Vin[i].PrevOut.N
		if t.Vin[i].PrevOut.Hash.GetHex() == "0000000000000000000000000000000000000000000000000000000000000000" {
			vinPrintAble.PrevOut.Hash = ""
		}
		vinPrintAble.ScriptSig = hex.EncodeToString(t.Vin[i].ScriptSig.GetScriptBytes())
		vinPrintAble.Sequence = t.Vin[i].Sequence
		trxPrintAble.Vin[i] = vinPrintAble
	}
	for i := 0; i < len(t.Vout); i++ {
		var voutPrintAble TxOutPrintAble
		voutPrintAble.Value = t.Vout[i].Value
		voutPrintAble.ScriptPubKey = hex.EncodeToString(t.Vout[i].ScriptPubKey.GetScriptBytes())
		isSucc, scriptType, addresses := script.ExtractDestination(t.Vout[i].ScriptPubKey)
		var addrStr string
		if isSucc {
			addrStr = ""
			if script.IsSingleAddress(scriptType) {
				addrStr = addresses[0]
			} else if script.IsMultiAddress(scriptType) {
				addrStr = strings.Join(addresses, ",")
			}
		}
		voutPrintAble.Address = addrStr
		voutPrintAble.ScriptType = script.GetScriptTypeStr(scriptType)
		trxPrintAble.Vout[i] = voutPrintAble
	}
	trxPrintAble.Version = t.Version
	trxPrintAble.LockTime = t.LockTime
	trxPrintAble.Version16 = t.Version16
	trxPrintAble.Type16 = t.Type16
	trxPrintAble.ExtraPayload = hex.EncodeToString(t.ExtraPayload.GetScriptBytes())

	return trxPrintAble
}
