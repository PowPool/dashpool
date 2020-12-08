package proxy

import (
	"math/big"
	"strconv"
	"strings"

	. "github.com/MiningPool0826/dashpool/util"
	"github.com/ethereum/ethash"
	"github.com/ethereum/go-ethereum/common"
)

var hasher = ethash.New()

func (s *ProxyServer) processShare(login, id, eNonce1, ip string, shareDiff int64, t *BlockTemplate, params []string) (bool, bool) {
	tplJobId := params[1]
	eNonce2Hex := params[2]
	nTimeHex := params[3]
	nonceHex := params[4]
	//nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)

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

	//share := Block{
	//	number:      h.height,
	//	hashNoNonce: common.HexToHash(hashNoNonce),
	//	difficulty:  big.NewInt(shareDiff),
	//	nonce:       nonce,
	//	mixDigest:   common.HexToHash(mixDigest),
	//}
	//
	//block := Block{
	//	number:      h.height,
	//	hashNoNonce: common.HexToHash(hashNoNonce),
	//	difficulty:  h.diff,
	//	nonce:       nonce,
	//	mixDigest:   common.HexToHash(mixDigest),
	//}

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

	if !hasher.Verify(share) {
		ms := MakeTimestamp()
		ts := ms / 1000

		err := s.backend.WriteRejectShare(ms, ts, login, id, shareDiff)
		if err != nil {
			Error.Println("Failed to insert reject share data into backend:", err)
		}

		return false, false
	}

	if hasher.Verify(block) {
		ok, err := s.rpc().SubmitBlock(params)
		if err != nil {
			Error.Printf("Block submission failure at height %v for %v: %v", t.Height, t.PrevHash, err)
			BlockLog.Printf("Block submission failure at height %v for %v: %v", t.Height, t.PrevHash, err)
		} else if !ok {
			Error.Printf("Block rejected at height %v for %v", t.Height, t.PrevHash)
			BlockLog.Printf("Block rejected at height %v for %v", t.Height, t.PrevHash)
			return false, false
		} else {
			s.fetchBlockTemplate()
			exist, err := s.backend.WriteBlock(login, id, params, shareDiff, t.Difficulty.Int64(), uint64(t.Height), s.hashrateExpiration)
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
		exist, err := s.backend.WriteShare(login, id, params, shareDiff, uint64(t.Height), s.hashrateExpiration)
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
