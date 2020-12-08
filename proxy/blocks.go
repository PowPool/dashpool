package proxy

import (
	"encoding/hex"
	"github.com/MiningPool0826/dashpool/dashcoin"
	"github.com/MiningPool0826/dashpool/rpc"
	. "github.com/MiningPool0826/dashpool/util"
	"github.com/mutalisk999/bitcoin-lib/src/utility"
	"math/big"
	"sync"
)

//const maxBacklog = 3
//
//type heightDiffPair struct {
//	diff   *big.Int
//	height uint64
//}

type BlockTemplateJob struct {
	BlkTplJobId   string
	BlkTplJobTime uint32
	TxIdList      []string
	MerkleBranch  []string
	CoinBase1     string
	CoinBase2     string
}

type BlockTemplate struct {
	sync.RWMutex
	Version        uint32
	Height         uint32
	PrevHash       string
	NBits          uint32
	Target         string
	Difficulty     *big.Int
	BlockTplJobMap map[string]BlockTemplateJob
	TxDetailMap    map[string]string
	updateTime     int64
	newBlkTpl      bool
	lastBlkTplId   string
}

//type Block struct {
//	difficulty  *big.Int
//	hashNoNonce common.Hash
//	nonce       uint64
//	mixDigest   common.Hash
//	number      uint64
//}

type Block struct {
	difficulty   *big.Int
	coinBase1    string
	coinBase2    string
	extraNonce1  string
	extraNonce2  string
	merkleBranch []string
	nVersion     uint32
	prevHash     string
	sTime        string
	nBits        uint32
	sNonce       string
}

//func (b Block) Difficulty() *big.Int     { return b.difficulty }
//func (b Block) HashNoNonce() common.Hash { return b.hashNoNonce }
//func (b Block) Nonce() uint64            { return b.nonce }
//func (b Block) MixDigest() common.Hash   { return b.mixDigest }
//func (b Block) NumberU64() uint64        { return b.number }

func (s *ProxyServer) fetchBlockTemplate() {
	rpcClient := s.rpc()
	prevBlockHash, err := rpcClient.GetPrevBlockHash()
	if err != nil {
		Error.Printf("Error while refreshing block template on %s: %s", rpcClient.Name, err)
		return
	}

	// No need to update, we have had fresh job
	blkTplIntv := MustParseDuration(s.config.Proxy.BlockTemplateCollectInterval)
	t := s.currentBlockTemplate()
	if t != nil && t.PrevHash == prevBlockHash && (MakeTimestamp()/1000-t.updateTime < int64(blkTplIntv.Seconds())) {
		return
	}

	blkTplReply, err := s.fetchPendingBlock()
	if err != nil {
		Error.Printf("Error while refreshing pending block on %s: %s", rpcClient.Name, err)
		return
	}

	var newTpl BlockTemplate
	if t == nil || t.PrevHash != blkTplReply.PreviousBlockHash {
		newTpl.Version = blkTplReply.Version
		newTpl.Height = blkTplReply.Height
		newTpl.PrevHash = blkTplReply.PreviousBlockHash
		newTpl.NBits = blkTplReply.Bits
		newTpl.Target = blkTplReply.Target
		newTpl.Difficulty = TargetHexToDiff(blkTplReply.Target)
		newTpl.BlockTplJobMap = make(map[string]BlockTemplateJob)
		newTpl.TxDetailMap = make(map[string]string)
		newTpl.updateTime = MakeTimestamp() / 1000
		newTpl.newBlkTpl = true
	} else {
		newTpl.Version = t.Version
		newTpl.Height = t.Height
		newTpl.PrevHash = t.PrevHash
		newTpl.NBits = t.NBits
		newTpl.Target = t.Target
		newTpl.Difficulty = TargetHexToDiff(blkTplReply.Target)
		newTpl.BlockTplJobMap = t.BlockTplJobMap
		newTpl.TxDetailMap = t.TxDetailMap
		newTpl.updateTime = MakeTimestamp() / 1000
		newTpl.newBlkTpl = false
	}

	var newTplJob BlockTemplateJob
	newTplJob.BlkTplJobTime = blkTplReply.CurTime
	for _, tx := range blkTplReply.Transactions {
		newTplJob.TxIdList = append(newTplJob.TxIdList, tx.Hash)
	}
	newTplJob.MerkleBranch, err = dashcoin.GetMerkleBranchHexFromTxIdsWithoutCoinBase(newTplJob.TxIdList)
	if err != nil {
		Error.Printf("Error while get merkle branch on %s: %s", rpcClient.Name, err)
		return
	}

	var coinBaseTx dashcoin.DashCoinBaseTransaction
	err = coinBaseTx.Initialize(s.config.UpstreamCoinBase, newTplJob.BlkTplJobTime, newTpl.Height, blkTplReply.CoinBaseValue,
		blkTplReply.CoinBaseAux.Flags, blkTplReply.CoinbasePayload, s.config.CoinBaseExtraData)
	if err != nil {
		Error.Printf("Error while initialize coinbase transaction on %s: %s", rpcClient.Name, err)
		return
	}
	newTplJob.CoinBase1 = hex.EncodeToString(coinBaseTx.CoinBaseTx1)
	newTplJob.CoinBase2 = hex.EncodeToString(coinBaseTx.CoinBaseTx2)
	newTplJob.BlkTplJobId = hex.EncodeToString(utility.Sha256(coinBaseTx.CoinBaseTx1))[0:16]

	newTpl.lastBlkTplId = newTplJob.BlkTplJobId
	newTpl.BlockTplJobMap[newTplJob.BlkTplJobId] = newTplJob
	for _, tx := range blkTplReply.Transactions {
		newTpl.TxDetailMap[tx.Hash] = tx.Data
	}

	s.blockTemplate.Store(&newTpl)
	Info.Printf("NEW pending block on %s at height %d / %s", rpcClient.Name, newTpl.Height, newTplJob.BlkTplJobId)

	// Stratum
	if s.config.Proxy.Stratum.Enabled {
		go s.broadcastNewJobs()
	}
}

func (s *ProxyServer) fetchPendingBlock() (*rpc.GetBlockTemplateReplyPart, error) {
	rpcClient := s.rpc()
	reply, err := rpcClient.GetPendingBlock()
	if err != nil {
		Error.Printf("Error while refreshing pending block on %s: %s", rpcClient.Name, err)
		return nil, err
	}
	return reply, nil
}
