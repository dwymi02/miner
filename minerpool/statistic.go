package minerpool

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/difficulty"
	"math/big"
)

// 统计算力
func (a *Account) addPowWorth(hash fields.Hash) {
	a.change.Lock()
	defer a.change.Unlock()

	//fmt.Println("addPowWorth", a, hash.ToHex())

	val := difficulty.CalculateHashWorth(hash)
	a.realtimePowWorth = new(big.Int).Add(a.realtimePowWorth, val)
}
