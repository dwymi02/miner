package minerpool

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/coinbase"
	"net"
	"sync"
	"sync/atomic"
)

type RealtimePeriod struct {
	minerpool *MinerPool

	miningSuccessBlock interfaces.Block

	targetBlock interfaces.Block

	realtimeAccounts map[string]*Account // [*Account]

	autoIncrementCoinbaseMsgNum uint32

	outputBlockCh *chan interfaces.Block


	changeLock sync.Mutex
}

func NewRealtimePeriod(minerpool *MinerPool, block interfaces.Block) *RealtimePeriod {
	per := &RealtimePeriod{
		miningSuccessBlock:          nil,
		minerpool:                   minerpool,
		targetBlock:                 block,
		realtimeAccounts:            make(map[string]*Account),
		autoIncrementCoinbaseMsgNum: 0,
		outputBlockCh:               nil,
	}
	return per
}

func (r *RealtimePeriod) getAutoIncrementCoinbaseMsgNum() uint32 {

	atomic.AddUint32(&r.autoIncrementCoinbaseMsgNum, 1)
	return r.autoIncrementCoinbaseMsgNum
}

func (r *RealtimePeriod) sendMiningStuffMsg(conn net.Conn) {
	r.changeLock.Lock()
	defer r.changeLock.Unlock()

	msgobj := message.NewPowMasterMsg()
	msgobj.CoinbaseMsgNum = fields.VarInt4(r.getAutoIncrementCoinbaseMsgNum())
	//fmt.Println("sendMiningStuffMsg", uint32(msgobj.CoinbaseMsgNum) )
	coinbase.UpdateBlockCoinbaseMessageForMiner(r.targetBlock, uint32(msgobj.CoinbaseMsgNum))
	r.targetBlock.SetMrklRoot(blocks.CalculateMrklRoot(r.targetBlock.GetTransactions()))
	msgobj.BlockHeadMeta = r.targetBlock
	// send data
	data, _ := msgobj.Serialize()
	go conn.Write(data)
}

// find ok
func (r *RealtimePeriod) successFindNewBlock(block interfaces.Block) {
	if r.outputBlockCh != nil {
		go func() {
			*r.outputBlockCh <- block // 挖出区块，传递给miner
		}()
	}
}


func (r *RealtimePeriod) IsOverEndBlock(blkheibts []byte) bool {
	tarhei := fields.VarInt5(0)
	tarhei.Parse(blkheibts[0:5], 0)
	return uint64(tarhei) == r.targetBlock.GetHeight()
}

// 结束当前挖矿
func (r *RealtimePeriod) endCurrentMining() {

	//fmt.Println("+++++++++++++++++++++ endCurrentMining ")
	go func() {
		for _, acc := range r.realtimeAccounts {
			clients := acc.activeClients.ToSlice()
			for _, cli := range clients {
				client := cli.(*Client)
				//fmt.Println(" -client.conn.Write([]byte(end_current_mining) ")
				client.conn.Write([]byte("end_current_mining"))
				// 不能结束连接，等待上传算力统计
			}
		}
	}()
}

///////////////////////////

func (r *RealtimePeriod) GetAccounts() []*Account {
	res := make([]*Account, 0)
	for _, acc := range r.realtimeAccounts {
		//fmt.Println("-----", acc.address.ToReadable())
		res = append(res, acc)
	}
	return res
}
