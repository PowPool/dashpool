package proxy

import (
	"encoding/hex"
	"github.com/MiningPool0826/dashpool/dashcoin"
	"github.com/MiningPool0826/dashpool/rpc"
	. "github.com/MiningPool0826/dashpool/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mutalisk999/bitcoin-lib/src/utility"
	"math/big"
	"sync"
)

const maxBacklog = 3

type heightDiffPair struct {
	diff   *big.Int
	height uint64
}

type BlockTemplate struct {
	BlkTplId     string
	BlkTplTime   uint32
	TxIdList     []string
	MerkleBranch []string
	CoinBase1    string
	CoinBase2    string
}

type BlockTemplatesCollection struct {
	sync.RWMutex
	Version      uint32
	Height       uint32
	PrevHash     string
	NBits        uint32
	Target       string
	Difficulty   *big.Int
	BlockTplMap  map[string]BlockTemplate
	TxDetailMap  map[string]string
	updateTime   int64
	newBlkTpl    bool
	lastBlkTplId string
}

type Block struct {
	difficulty  *big.Int
	hashNoNonce common.Hash
	nonce       uint64
	mixDigest   common.Hash
	number      uint64
}

func (b Block) Difficulty() *big.Int     { return b.difficulty }
func (b Block) HashNoNonce() common.Hash { return b.hashNoNonce }
func (b Block) Nonce() uint64            { return b.nonce }
func (b Block) MixDigest() common.Hash   { return b.mixDigest }
func (b Block) NumberU64() uint64        { return b.number }

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

	var newTplCollection BlockTemplatesCollection
	if t == nil || t.PrevHash != blkTplReply.PreviousBlockHash {
		newTplCollection.Version = blkTplReply.Version
		newTplCollection.Height = blkTplReply.Height
		newTplCollection.PrevHash = blkTplReply.PreviousBlockHash
		newTplCollection.NBits = blkTplReply.Bits
		newTplCollection.Target = blkTplReply.Target
		newTplCollection.Difficulty = TargetHexToDiff(blkTplReply.Target)
		newTplCollection.BlockTplMap = make(map[string]BlockTemplate)
		newTplCollection.TxDetailMap = make(map[string]string)
		newTplCollection.updateTime = MakeTimestamp() / 1000
		newTplCollection.newBlkTpl = true
	} else {
		newTplCollection.Version = t.Version
		newTplCollection.Height = t.Height
		newTplCollection.PrevHash = t.PrevHash
		newTplCollection.NBits = t.NBits
		newTplCollection.Target = t.Target
		newTplCollection.Difficulty = TargetHexToDiff(blkTplReply.Target)
		newTplCollection.BlockTplMap = t.BlockTplMap
		newTplCollection.TxDetailMap = t.TxDetailMap
		newTplCollection.updateTime = MakeTimestamp() / 1000
		newTplCollection.newBlkTpl = false
	}

	var newTpl BlockTemplate
	newTpl.BlkTplTime = blkTplReply.CurTime
	for _, tx := range blkTplReply.Transactions {
		newTpl.TxIdList = append(newTpl.TxIdList, tx.Hash)
	}
	newTpl.MerkleBranch, err = dashcoin.GetMerkleBranchHexFromTxIdsWithoutCoinBase(newTpl.TxIdList)
	if err != nil {
		Error.Printf("Error while get merkle branch on %s: %s", rpcClient.Name, err)
		return
	}

	var coinBaseTx dashcoin.DashCoinBaseTransaction
	err = coinBaseTx.Initialize(s.config.UpstreamCoinBase, newTpl.BlkTplTime, newTplCollection.Height, blkTplReply.CoinBaseValue,
		blkTplReply.CoinBaseAux.Flags, blkTplReply.CoinbasePayload, s.config.CoinBaseExtraData)
	if err != nil {
		Error.Printf("Error while initialize coinbase transaction on %s: %s", rpcClient.Name, err)
		return
	}
	newTpl.CoinBase1 = hex.EncodeToString(coinBaseTx.CoinBaseTx1)
	newTpl.CoinBase2 = hex.EncodeToString(coinBaseTx.CoinBaseTx2)
	newTpl.BlkTplId = hex.EncodeToString(utility.Sha256(coinBaseTx.CoinBaseTx1))[0:16]

	newTplCollection.lastBlkTplId = newTpl.BlkTplId
	newTplCollection.BlockTplMap[newTpl.BlkTplId] = newTpl
	for _, tx := range blkTplReply.Transactions {
		newTplCollection.TxDetailMap[tx.Hash] = tx.Data
	}

	s.blockTemplatesCollection.Store(&newTplCollection)
	Info.Printf("NEW pending block on %s at height %d / %s", rpcClient.Name, newTplCollection.Height, newTpl.BlkTplId)

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
