package proxy

import (
	"bytes"
	"encoding/hex"
	"github.com/PowPool/dashpool/dashcoin"
	"github.com/PowPool/dashpool/rpc"
	. "github.com/PowPool/dashpool/util"
	"github.com/mutalisk999/bitcoin-lib/src/utility"
	"github.com/mutalisk999/txid_merkle_tree"
	"io"
	"math/big"
	"strconv"
	"sync"
)

type BlockTemplateJob struct {
	BlkTplJobId    string
	BlkTplJobTime  uint32
	TxIdList       []string
	MerkleBranch   []string
	CoinBase1      string
	CoinBase2      string
	CoinBaseValue  int64
	JobTxsFeeTotal int64
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

func (s *ProxyServer) fetchBlockTemplate() {
	rpcClient := s.rpc()
	prevBlockHash, err := rpcClient.GetPrevBlockHash()
	if err != nil {
		Error.Printf("Error while refreshing block template on %s: %s", rpcClient.Name, err)
		return
	}

	// No need to update, we have had fresh job
	blkTplIntv := MustParseDuration(s.config.Proxy.BlockTemplateInterval)
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
		nBits, err := strconv.ParseInt(blkTplReply.Bits, 16, 32)
		if err != nil {
			Error.Printf("Error while ParseInt nBits on %s: %s", rpcClient.Name, err)
			return
		}
		newTpl.Version = blkTplReply.Version
		newTpl.Height = blkTplReply.Height
		newTpl.PrevHash = blkTplReply.PreviousBlockHash
		newTpl.NBits = uint32(nBits)
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
	newTplJob.MerkleBranch, err = txid_merkle_tree.GetMerkleBranchHexFromTxIdsWithoutCoinBase(newTplJob.TxIdList)
	if err != nil {
		Error.Printf("Error while get merkle branch on %s: %s", rpcClient.Name, err)
		return
	}

	coinBaseReward := blkTplReply.CoinBaseValue
	for _, masterNode := range blkTplReply.MasterNodes {
		coinBaseReward -= masterNode.Amount
	}

	if coinBaseReward <= 0 {
		Error.Printf("Invalid block template, coinBaseReward <= 0")
		return
	}

	var coinBaseTx dashcoin.DashCoinBaseTransaction
	err = coinBaseTx.Initialize(s.config.UpstreamCoinBase, newTplJob.BlkTplJobTime, newTpl.Height, coinBaseReward,
		blkTplReply.CoinBaseAux.Flags, blkTplReply.CoinbasePayload, s.config.CoinBaseExtraData, blkTplReply.MasterNodes)
	if err != nil {
		Error.Printf("Error while initialize coinbase transaction on %s: %s", rpcClient.Name, err)
		return
	}
	newTplJob.CoinBase1 = hex.EncodeToString(coinBaseTx.CoinBaseTx1)
	newTplJob.CoinBase2 = hex.EncodeToString(coinBaseTx.CoinBaseTx2)
	newTplJob.CoinBaseValue = coinBaseReward
	newTplJob.JobTxsFeeTotal = 0
	for _, tx := range blkTplReply.Transactions {
		newTplJob.JobTxsFeeTotal += tx.Fee
	}
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

func ConstructRawDashBlockHex(block *Block, tplJob *BlockTemplateJob, tpl *BlockTemplate) (string, error) {
	bytes1, err := hex.DecodeString(block.coinBase1)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: hex decode coinBase1 error")
		return "", err
	}
	bytes2, err := hex.DecodeString(block.extraNonce1)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: hex decode extraNonce1 error")
		return "", err
	}
	bytes3, err := hex.DecodeString(block.extraNonce2)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: hex decode extraNonce2 error")
		return "", err
	}
	bytes4, err := hex.DecodeString(block.coinBase2)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: hex decode coinBase2 error")
		return "", err
	}

	// construct coin base transaction
	bytesCoinBaseTx := append(append(append(append([]byte{}, bytes1...), bytes2...), bytes3...), bytes4...)
	bytesBuf := bytes.NewBuffer(bytesCoinBaseTx)
	bufReader := io.Reader(bytesBuf)
	var cbTrx dashcoin.DashTransaction
	err = cbTrx.UnPack(bufReader)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: unpack coinBase transaction error")
		return "", err
	}

	// get coin base transaction id
	cbTrxId, err := cbTrx.CalcTrxId()
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: CalcTrxId error")
		return "", err
	}

	// get merkle root hash
	merkleRootHex, err := txid_merkle_tree.GetMerkleRootHexFromCoinBaseAndMerkleBranch(cbTrxId.GetHex(), block.merkleBranch)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: GetMerkleRootHexFromCoinBaseAndMerkleBranch error")
		return "", err
	}

	// construct block header
	var dashBlock dashcoin.Block
	dashBlock.Header.Version = int32(block.nVersion)
	err = dashBlock.Header.HashPrevBlock.SetHex(block.prevHash)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: HashPrevBlock SetHex error")
		return "", err
	}
	err = dashBlock.Header.HashMerkleRoot.SetHex(merkleRootHex)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: HashMerkleRoot SetHex error")
		return "", err
	}
	nTime, err := strconv.ParseUint(block.sTime, 16, 32)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: ParseUint sTime error")
		return "", err
	}
	dashBlock.Header.Time = uint32(nTime)
	dashBlock.Header.Bits = block.nBits
	nNonce, err := strconv.ParseUint(block.sNonce, 16, 32)
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: ParseUint sNonce error")
		return "", err
	}
	dashBlock.Header.Nonce = uint32(nNonce)

	// add transactions
	// add coin base transaction
	dashBlock.Vtx = append(dashBlock.Vtx, cbTrx)

	// add other transaction
	for _, trxId := range tplJob.TxIdList {
		rawTrxHex, ok := tpl.TxDetailMap[trxId]
		if !ok {
			Error.Printf("ConstructRawDashBlockHex: get TxDetailMap key [%s] error", trxId)
			return "", err
		}
		var trx dashcoin.DashTransaction
		err = trx.UnPackFromHex(rawTrxHex)
		if err != nil {
			Error.Println("ConstructRawDashBlockHex: trx UnPackFromHex error")
			return "", err
		}
		dashBlock.Vtx = append(dashBlock.Vtx, trx)
	}

	rawBlockHex, err := dashBlock.PackToHex()
	if err != nil {
		Error.Println("ConstructRawDashBlockHex: dashBlock PackToHex error")
		return "", err
	}

	return rawBlockHex, nil
}
