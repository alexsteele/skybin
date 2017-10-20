package repo

import (
	"errors"
	core "skybin/core/proto"
)

type contractInfo struct {
	provider core.Provider
	contract *core.Contract
}

func (r *repo) negotiateContract(block blockInfo, provider core.Provider) (*core.Contract, error) {
	info, err := provider.Info()
	if err != nil {
		return nil, err
	}
	if info.MaxBlockSize < int32(block.Size) {
		return nil, errors.New("provider.MaxBlockSize < block.Size")
	}
	contract := core.Contract{
		BlockID:    block.ID,
		BlockSize:  int64(block.Size),
		RenterID:   r.config.UserId,
		ProviderID: info.ID,
	}
	c, err := provider.Negotiate(&contract)
	if err != nil {
		return nil, err
	}
	accepted := len(c.ProviderSignature) > 0
	if !accepted {
		return nil, errors.New("contract terms not accepted")
	}
	return c, nil
}

func (r *repo) negotiateContracts(block blockInfo, providers []core.Provider) ([]contractInfo, error) {
	var contracts []contractInfo
	for _, provider := range providers {
		contract, err := r.negotiateContract(block, provider)
		if err != nil {
			r.logger.Println(err)
			continue
		}
		contracts = append(contracts, contractInfo{
			provider: provider,
			contract: contract,
		})
		break
	}
	if len(contracts) == 0 {
		return nil, errors.New("cannot find provider for block")
	}
	return contracts, nil
}
