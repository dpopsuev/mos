// Package chain re-exports from artifact/chain for backward compatibility.
package chain

import achain "github.com/dpopsuev/mos/moslib/artifact/chain"

type ChainLink = achain.ChainLink
type ChainResult = achain.ChainResult
type NegativeChainResult = achain.NegativeChainResult
type NegativeSpaceEntry = achain.NegativeSpaceEntry

func WalkChain(root, startKind, startID string) (*ChainResult, error) {
	return achain.WalkChain(root, startKind, startID)
}

func WalkNegativeChain(root, startKind, startID string) (*NegativeChainResult, error) {
	return achain.WalkNegativeChain(root, startKind, startID)
}

func FormatChain(c *ChainResult) string               { return achain.FormatChain(c) }
func FormatNegativeChain(nc *NegativeChainResult) string { return achain.FormatNegativeChain(nc) }
