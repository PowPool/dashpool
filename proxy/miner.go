package proxy

import (
	"bytes"
	"encoding/hex"
	"github.com/PowPool/dashpool/dashcoin"
	"github.com/PowPool/dashpool/goX11"
	. "github.com/PowPool/dashpool/util"
	"github.com/mutalisk999/bitcoin-lib/src/blob"
	"github.com/mutalisk999/txid_merkle_tree"
	"io"
	"math/big"
	"strconv"
)

func (s *ProxyServer) processShare(login, id, eNonce1, ip string, shareDiff int64, t *BlockTemplate, params []string) (bool, bool) {
	tplJobId := params[1]
	eNonce2Hex := params[2]
	nTimeHex := params[3]
	nonceHex := params[4]

	h, ok := t.BlockTplJobMap[tplJobId]
	if !ok {
		Error.Printf("Stale share from %v.%v@%v", login, id, ip)
		ShareLog.Printf("Stale share from %v.%v@%v", login, id, ip)

		ms := MakeTimestamp()
		ts := ms / 1000

		err := s.backend.WriteInvalidShare(ms, ts, login, id, shareDiff)
		if err != nil {
			Error.Println("Failed to insert invalid share data into backend:", err)
		}
		return false, false
	}

	share := Block{
		difficulty:   big.NewInt(shareDiff),
		coinBase1:    h.CoinBase1,
		coinBase2:    h.CoinBase2,
		extraNonce1:  eNonce1,
		extraNonce2:  eNonce2Hex,
		merkleBranch: h.MerkleBranch,
		nVersion:     t.Version,
		prevHash:     t.PrevHash,
		sTime:        nTimeHex,
		nBits:        t.NBits,
		sNonce:       nonceHex,
	}

	block := Block{
		difficulty:   t.Difficulty,
		coinBase1:    h.CoinBase1,
		coinBase2:    h.CoinBase2,
		extraNonce1:  eNonce1,
		extraNonce2:  eNonce2Hex,
		merkleBranch: h.MerkleBranch,
		nVersion:     t.Version,
		prevHash:     t.PrevHash,
		sTime:        nTimeHex,
		nBits:        t.NBits,
		sNonce:       nonceHex,
	}

	if !X11HashVerify(&share) {
		ms := MakeTimestamp()
		ts := ms / 1000

		err := s.backend.WriteRejectShare(ms, ts, login, id, shareDiff)
		if err != nil {
			Error.Println("Failed to insert reject share data into backend:", err)
		}
		return false, false
	}

	paramIn := []string{nonceHex, eNonce1, eNonce2Hex}
	if X11HashVerify(&block) {
		// construct new block
		rawBlockHex, err := ConstructRawDashBlockHex(&block, &h, t)
		if err != nil {
			return false, false
		}
		err = s.rpc().SubmitBlock([]interface{}{rawBlockHex})
		if err != nil {
			Error.Printf("Block submission failure at height %v for %v: %v", t.Height, t.PrevHash, err)
			BlockLog.Printf("Block submission failure at height %v for %v: %v", t.Height, t.PrevHash, err)
		} else {
			s.fetchBlockTemplate()
			exist, err := s.backend.WriteBlock(login, id, paramIn, shareDiff, t.Difficulty.Int64(), uint64(t.Height),
				h.CoinBaseValue, h.JobTxsFeeTotal, s.hashrateExpiration)
			if exist {
				ms := MakeTimestamp()
				ts := ms / 1000

				err := s.backend.WriteInvalidShare(ms, ts, login, id, shareDiff)
				if err != nil {
					Error.Println("Failed to insert invalid share data into backend:", err)
				}
				return true, false
			}
			if err != nil {
				Error.Println("Failed to insert block candidate into backend:", err)
				BlockLog.Println("Failed to insert block candidate into backend:", err)
			} else {
				Info.Printf("Inserted block %v to backend", t.Height)
				BlockLog.Printf("Inserted block %v to backend", t.Height)
			}
			Info.Printf("Block found by miner %v@%v at height %d", login, ip, t.Height)
			BlockLog.Printf("Block found by miner %v@%v at height %d", login, ip, t.Height)
		}
	} else {
		exist, err := s.backend.WriteShare(login, id, paramIn, shareDiff, uint64(t.Height), s.hashrateExpiration)
		if exist {
			ms := MakeTimestamp()
			ts := ms / 1000

			err := s.backend.WriteInvalidShare(ms, ts, login, id, shareDiff)
			if err != nil {
				Error.Println("Failed to insert invalid share data into backend:", err)
			}
			return true, false
		}
		if err != nil {
			Error.Println("Failed to insert share data into backend:", err)
		}
	}
	return false, true
}

func X11HashVerify(block *Block) bool {
	bytes1, err := hex.DecodeString(block.coinBase1)
	if err != nil {
		Error.Println("X11HashVerify: hex decode coinBase1 error")
		return false
	}
	bytes2, err := hex.DecodeString(block.extraNonce1)
	if err != nil {
		Error.Println("X11HashVerify: hex decode extraNonce1 error")
		return false
	}
	bytes3, err := hex.DecodeString(block.extraNonce2)
	if err != nil {
		Error.Println("X11HashVerify: hex decode extraNonce2 error")
		return false
	}
	bytes4, err := hex.DecodeString(block.coinBase2)
	if err != nil {
		Error.Println("X11HashVerify: hex decode coinBase2 error")
		return false
	}

	Debug.Printf("block.coinBase1: %s", block.coinBase1)
	Debug.Printf("block.extraNonce1: %s", block.extraNonce1)
	Debug.Printf("block.extraNonce2: %s", block.extraNonce2)
	Debug.Printf("block.coinBase2: %s", block.coinBase2)

	// construct coin base transaction
	bytesCoinBaseTx := append(append(append(append([]byte{}, bytes1...), bytes2...), bytes3...), bytes4...)
	bytesBuf := bytes.NewBuffer(bytesCoinBaseTx)
	bufReader := io.Reader(bytesBuf)
	var cbTrx dashcoin.DashTransaction
	err = cbTrx.UnPack(bufReader)
	if err != nil {
		Error.Println("X11HashVerify: unpack coinBase transaction error")
		return false
	}

	// get coin base transaction id
	cbTrxId, err := cbTrx.CalcTrxId()
	if err != nil {
		Error.Println("X11HashVerify: CalcTrxId error")
		return false
	}

	Debug.Printf("coinBase trx id: %s", cbTrxId.GetHex())
	Debug.Printf("block.merkleBranch: %v", block.merkleBranch)

	// get merkle root hash
	merkleRootHex, err := txid_merkle_tree.GetMerkleRootHexFromCoinBaseAndMerkleBranch(cbTrxId.GetHex(), block.merkleBranch)
	if err != nil {
		Error.Println("X11HashVerify: GetMerkleRootHexFromCoinBaseAndMerkleBranch error")
		return false
	}

	Debug.Printf("merkleRootHex: %s", merkleRootHex)

	// construct block header
	var blockHeader dashcoin.BlockHeader
	blockHeader.Version = int32(block.nVersion)
	err = blockHeader.HashPrevBlock.SetHex(block.prevHash)
	if err != nil {
		Error.Println("X11HashVerify: HashPrevBlock SetHex error")
		return false
	}
	err = blockHeader.HashMerkleRoot.SetHex(merkleRootHex)
	if err != nil {
		Error.Println("X11HashVerify: HashMerkleRoot SetHex error")
		return false
	}
	nTime, err := strconv.ParseUint(block.sTime, 16, 32)
	if err != nil {
		Error.Println("X11HashVerify: ParseUint sTime error")
		return false
	}
	blockHeader.Time = uint32(nTime)
	blockHeader.Bits = block.nBits
	nNonce, err := strconv.ParseUint(block.sNonce, 16, 32)
	if err != nil {
		Error.Println("X11HashVerify: ParseUint sNonce error")
		return false
	}
	blockHeader.Nonce = uint32(nNonce)

	bytesBuf = bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err = blockHeader.Pack(bufWriter)
	if err != nil {
		Error.Println("X11HashVerify: blockHeader Pack error")
		return false
	}

	Debug.Printf("blockHeader.Version: %d", blockHeader.Version)
	Debug.Printf("blockHeader.HashPrevBlock: %s", blockHeader.HashPrevBlock.GetHex())
	Debug.Printf("blockHeader.HashMerkleRoot: %s", blockHeader.HashMerkleRoot.GetHex())
	Debug.Printf("blockHeader.Time: %d", blockHeader.Time)
	Debug.Printf("blockHeader.Bits: %d", blockHeader.Bits)
	Debug.Printf("blockHeader.Nonce: %d", blockHeader.Nonce)

	Debug.Printf("blockHeader Hex: %s", hex.EncodeToString(bytesBuf.Bytes()))

	// calc block header hash
	bytesRes := goX11.CalcX11Hash(bytesBuf.Bytes())
	var res blob.Baseblob
	res.SetData(bytesRes)
	resHex := res.GetHex()

	Debug.Printf("Target Hex: %064s", resHex)

	hashDiff := TargetHexToDiff(resHex)

	if hashDiff.Cmp(block.difficulty) > 0 {
		return true
	} else {
		return false
	}
}
