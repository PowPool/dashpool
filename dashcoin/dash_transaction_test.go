package dashcoin

import (
	"bytes"
	"fmt"
	"github.com/mutalisk999/bitcoin-lib/src/blob"
	"io"
	"testing"
)

func TestDashTransaction(t *testing.T) {
	Blob := new(blob.Byteblob)
	_ = Blob.SetHex("03000500010000000000000000000000000000000000000000000000000000000000000000ffffffff1f0222070414a3c05f08f8000001010000000d2f7374726174756d506f6f6c2f00000000013ca93d4e040000001976a914521dbb202daf1dec4d36181479508513d10d4cd088ac000000004602002207000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	fmt.Println("byte blob:", Blob.GetData())

	bytesBuf := bytes.NewBuffer(Blob.GetData())
	bufReader := io.Reader(bytesBuf)
	trx := new(DashTransaction)
	_ = trx.UnPack(bufReader)
	fmt.Println("trx version:", trx.Version)
	fmt.Println("trx locktime", trx.LockTime)
	fmt.Println("trx vin size:", len(trx.Vin))
	for i := 0; i < len(trx.Vin); i++ {
		fmt.Println("vin prevout:", trx.Vin[i].PrevOut)
		fmt.Println("vin scriptsig:", trx.Vin[i].ScriptSig)
		fmt.Println("vin sequence:", trx.Vin[i].Sequence)
	}
	fmt.Println("trx vout size:", len(trx.Vout))
	for i := 0; i < len(trx.Vout); i++ {
		fmt.Println("vout value", trx.Vout[i].Value)
		fmt.Println("vout scriptpubkey:", trx.Vout[i].ScriptPubKey)
	}
	fmt.Println("trx version16:", trx.Version16)
	fmt.Println("trx type16:", trx.Type16)
	fmt.Println("trx extrapayload:", trx.ExtraPayload)
	bytesBuf = bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	_ = trx.Pack(bufWriter)
	fmt.Println("byte buffer:", bytesBuf.Bytes())
}

func TestDashTransactionHex(t *testing.T) {
	trx := new(DashTransaction)
	_ = trx.UnPackFromHex("03000500010000000000000000000000000000000000000000000000000000000000000000ffffffff1f0222070414a3c05f08f8000001010000000d2f7374726174756d506f6f6c2f00000000013ca93d4e040000001976a914521dbb202daf1dec4d36181479508513d10d4cd088ac000000004602002207000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	fmt.Println("trx version:", trx.Version)
	fmt.Println("trx locktime", trx.LockTime)
	fmt.Println("trx vin size:", len(trx.Vin))
	for i := 0; i < len(trx.Vin); i++ {
		fmt.Println("vin prevout:", trx.Vin[i].PrevOut)
		fmt.Println("vin scriptsig:", trx.Vin[i].ScriptSig)
		fmt.Println("vin sequence:", trx.Vin[i].Sequence)
	}
	fmt.Println("trx vout size:", len(trx.Vout))
	for i := 0; i < len(trx.Vout); i++ {
		fmt.Println("vout value", trx.Vout[i].Value)
		fmt.Println("vout scriptpubkey:", trx.Vout[i].ScriptPubKey)
	}
	fmt.Println("trx version16:", trx.Version16)
	fmt.Println("trx type16:", trx.Type16)
	fmt.Println("trx extrapayload:", trx.ExtraPayload)
	hexStr, _ := trx.PackToHex()
	fmt.Println("hex string:", hexStr)
}

func TestCalcTrxId(t *testing.T) {
	trx := new(DashTransaction)
	_ = trx.UnPackFromHex("03000500010000000000000000000000000000000000000000000000000000000000000000ffffffff1f0222070414a3c05f08f8000001010000000d2f7374726174756d506f6f6c2f00000000013ca93d4e040000001976a914521dbb202daf1dec4d36181479508513d10d4cd088ac000000004602002207000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	trxId, _ := trx.CalcTrxId()
	fmt.Println("trx id:", trxId.GetHex())
}
