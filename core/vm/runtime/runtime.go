// Copyright 2015 The go-ethereum Authors
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

// Package runtime provides a basic execution model for executing EVM code.
package runtime

import (
	"math/big"
	"time"

	"github.com/eth-classic/go-ethereum/common"
	"github.com/eth-classic/go-ethereum/core/state"
	"github.com/eth-classic/go-ethereum/core/vm"
	"github.com/eth-classic/go-ethereum/crypto"
	"github.com/eth-classic/go-ethereum/ethdb"
)

// The default, always homestead, rule set for the vm env
type ruleSet struct{}

func (ruleSet) IsHomestead(*big.Int) bool { return true }
func (ruleSet) IsAtlantis(*big.Int) bool  { return true }
func (ruleSet) IsAgharta(*big.Int) bool   { return true }
func (ruleSet) GasTable(*big.Int) *vm.GasTable {

	// IsAgharta will always return true
	// just have gastable default to returning the Agharta GasTable
	return &vm.GasTable{
		ExtcodeSize:     big.NewInt(700),
		ExtcodeCopy:     big.NewInt(700),
		ExtcodeHash:     big.NewInt(400),
		Balance:         big.NewInt(400),
		SLoad:           big.NewInt(200),
		Calls:           big.NewInt(700),
		Suicide:         big.NewInt(5000),
		ExpByte:         big.NewInt(50),
		CreateBySuicide: big.NewInt(25000),
	}
}

// Config is a basic type specifying certain configuration flags for running
// the EVM.
type Config struct {
	RuleSet     vm.RuleSet
	Difficulty  *big.Int
	Origin      common.Address
	Coinbase    common.Address
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    *big.Int
	GasPrice    *big.Int
	Value       *big.Int
	DisableJit  bool // "disable" so it's enabled by default
	Debug       bool

	State     *state.StateDB
	GetHashFn func(n uint64) common.Hash
}

// sets defaults on the config
func setDefaults(cfg *Config) {
	if cfg.RuleSet == nil {
		cfg.RuleSet = ruleSet{}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.Time == nil {
		cfg.Time = big.NewInt(time.Now().Unix())
	}
	if cfg.GasLimit == nil {
		cfg.GasLimit = new(big.Int).Set(common.MaxBig)
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
}

// Execute executes the code using the input as call data during the execution.
// It returns the EVM's return value, the new state and an error if it failed.
//
// Executes sets up a in memory, temporarily, environment for the execution of
// the given code. It enabled the JIT by default and make sure that it's restored
// to it's original state afterwards.
func Execute(code, input []byte, cfg *Config) ([]byte, *state.StateDB, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	setDefaults(cfg)

	if cfg.State == nil {
		db, _ := ethdb.NewMemDatabase()
		cfg.State, _ = state.New(common.Hash{}, state.NewDatabase(db))
	}
	var (
		vmenv    = NewEnv(cfg, cfg.State)
		sender   = cfg.State.CreateAccount(cfg.Origin)
		receiver = cfg.State.CreateAccount(common.StringToAddress("contract"))
	)
	// set the receiver's (the executing contract) code for execution.
	receiver.SetCode(crypto.Keccak256Hash(code), code)

	// Call the code with the given configuration.
	ret, err := vmenv.Call(
		sender,
		receiver.Address(),
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
	)

	return ret, cfg.State, err
}

// Call executes the code given by the contract's address. It will return the
// EVM's return value or an error if it failed.
//
// Call, unlike Execute, requires a config and also requires the State field to
// be set.
func Call(address common.Address, input []byte, cfg *Config) ([]byte, error) {
	setDefaults(cfg)

	vmenv := NewEnv(cfg, cfg.State)

	sender := cfg.State.GetOrNewStateObject(cfg.Origin)
	// Call the code with the given configuration.
	ret, err := vmenv.Call(
		sender,
		address,
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
	)

	return ret, err
}
