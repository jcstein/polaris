// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2023, Berachain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package distribution

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	generated "pkg.berachain.dev/polaris/contracts/bindings/cosmos/precompile/distribution"
	"pkg.berachain.dev/polaris/cosmos/x/evm/plugins/precompile/log"
	"pkg.berachain.dev/polaris/eth/common"
	ethprecompile "pkg.berachain.dev/polaris/eth/core/precompile"
)

// Contract is the precompile contract for the distribution module.
type Contract struct {
	ethprecompile.BaseContract

	msgServer distributiontypes.MsgServer
	querier   distributiontypes.QueryServer
}

// NewPrecompileContract returns a new instance of the distribution module precompile contract.
func NewPrecompileContract(
	m distributiontypes.MsgServer, q distributiontypes.QueryServer,
) *Contract {
	return &Contract{
		BaseContract: ethprecompile.NewBaseContract(
			generated.DistributionModuleMetaData.ABI,
			common.BytesToAddress([]byte{0x69}),
		),
		msgServer: m,
		querier:   q,
	}
}

// CustomValueDecoders overrides the `coreprecompile.StatefulImpl` interface.
func (c *Contract) CustomValueDecoders() ethprecompile.ValueDecoders {
	return ethprecompile.ValueDecoders{
		distributiontypes.AttributeKeyWithdrawAddress: log.ConvertAccAddressFromBech32,
	}
}

// SetWithdrawAddress is the precompile contract method for the `setWithdrawAddress(address)` method.
func (c *Contract) SetWithdrawAddress(
	ctx context.Context,
	_ ethprecompile.EVM,
	caller common.Address,
	_ *big.Int,
	withdrawAddress common.Address,
) ([]any, error) {
	return c.setWithdrawAddressHelper(ctx, sdk.AccAddress(caller.Bytes()), sdk.AccAddress(withdrawAddress.Bytes()))
}

// GetWithdrawEnabled is the precompile contract method for the `getWithdrawEnabled()` method.
func (c *Contract) GetWithdrawEnabled(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
) ([]any, error) {
	return c.getWithdrawAddrEnabled(ctx)
}

// WithdrawDelegatorReward is the precompile contract method for the `withdrawDelegatorReward(address,address)`
// method.
func (c *Contract) WithdrawDelegatorReward(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	delegator common.Address,
	validator common.Address,
) ([]any, error) {
	return c.withdrawDelegatorRewardsHelper(ctx, sdk.AccAddress(delegator.Bytes()), sdk.ValAddress(validator.Bytes()))
}
