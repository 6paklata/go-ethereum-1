// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"

	"github.com/eth-classic/go-ethereum/common"
	"github.com/eth-classic/go-ethereum/core/state"
	"github.com/eth-classic/go-ethereum/core/types"
	"github.com/eth-classic/go-ethereum/core/vm"
)

// GetHashFn returns a function for which the VM env can query block hashes through
// up to the limit defined by the Yellow Paper and uses the given block chain
// to query for information.
func GetHashFn(ref common.Hash, chain *BlockChain) func(n uint64) common.Hash {
	return func(n uint64) common.Hash {
		for block := chain.GetBlock(ref); block != nil; block = chain.GetBlock(block.ParentHash()) {
			if block.NumberU64() == n {
				return block.Hash()
			}
		}

		return common.Hash{}
	}
}

type VMEnv struct {
	chainConfig *ChainConfig   // Chain configuration
	state       *state.StateDB // State to use for executing
	evm         *vm.EVM        // The Ethereum Virtual Machine
	depth       int            // Current execution depth
	returnData  []byte
	msg         Message // Message applied

	header    *types.Header            // Header information
	chain     *BlockChain              // Blockchain handle
	getHashFn func(uint64) common.Hash // getHashFn callback is used to retrieve block hashes
}

func NewEnv(state *state.StateDB, chainConfig *ChainConfig, chain *BlockChain, msg Message, header *types.Header) *VMEnv {
	env := &VMEnv{
		chainConfig: chainConfig,
		chain:       chain,
		state:       state,
		header:      header,
		msg:         msg,
		getHashFn:   GetHashFn(header.ParentHash, chain),
	}

	env.evm = vm.New(env)
	return env
}

func (self *VMEnv) RuleSet() vm.RuleSet       { return self.chainConfig }
func (self *VMEnv) Vm() vm.Vm                 { return self.evm }
func (self *VMEnv) Origin() common.Address    { f, _ := self.msg.From(); return f }
func (self *VMEnv) BlockNumber() *big.Int     { return self.header.Number }
func (self *VMEnv) Coinbase() common.Address  { return self.header.Coinbase }
func (self *VMEnv) Time() *big.Int            { return self.header.Time }
func (self *VMEnv) Difficulty() *big.Int      { return self.header.Difficulty }
func (self *VMEnv) GasLimit() *big.Int        { return self.header.GasLimit }
func (self *VMEnv) Value() *big.Int           { return self.msg.Value() }
func (self *VMEnv) Db() vm.Database           { return self.state }
func (self *VMEnv) Depth() int                { return self.depth }
func (self *VMEnv) SetDepth(i int)            { self.depth = i }
func (self *VMEnv) ReturnData() []byte        { return self.returnData }
func (self *VMEnv) SetReturnData(data []byte) { self.returnData = data }
func (self *VMEnv) GetHash(n uint64) common.Hash {
	return self.getHashFn(n)
}

func (self *VMEnv) AddLog(log *vm.Log) {
	self.state.AddLog(*log)
}
func (self *VMEnv) CanTransfer(from common.Address, balance *big.Int) bool {
	return self.state.GetBalance(from).Cmp(balance) >= 0
}

func (self *VMEnv) SnapshotDatabase() int {
	return self.state.Snapshot()
}

func (self *VMEnv) RevertToSnapshot(snapshot int) {
	self.state.RevertToSnapshot(snapshot)
}

func (self *VMEnv) Transfer(from, to vm.Account, amount *big.Int) {
	Transfer(from, to, amount)
}

func (self *VMEnv) Call(me vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return Call(self, me, addr, data, gas, price, value)
}
func (self *VMEnv) CallCode(me vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return CallCode(self, me, addr, data, gas, price, value)
}

func (self *VMEnv) DelegateCall(me vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return DelegateCall(self, me, addr, data, gas, price)
}

func (self *VMEnv) StaticCall(me vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return StaticCall(self, me, addr, data, gas, price)
}

func (self *VMEnv) Create(me vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return Create(self, me, data, gas, price, value)
}

func (self *VMEnv) Create2(me vm.ContractRef, data []byte, gas, price, salt, value *big.Int) ([]byte, common.Address, error) {
	return Create2(self, me, data, gas, price, salt, value)
}
