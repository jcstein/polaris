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

package staking

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	generated "pkg.berachain.dev/polaris/contracts/bindings/cosmos/precompile/staking"
	cosmlib "pkg.berachain.dev/polaris/cosmos/lib"
	"pkg.berachain.dev/polaris/cosmos/precompile"
	"pkg.berachain.dev/polaris/eth/common"
	ethprecompile "pkg.berachain.dev/polaris/eth/core/precompile"
	"pkg.berachain.dev/polaris/lib/utils"
)

// Contract is the precompile contract for the staking module.
type Contract struct {
	ethprecompile.BaseContract

	msgServer stakingtypes.MsgServer
	querier   stakingtypes.QueryServer
}

// NewContract is the constructor of the staking contract.
func NewPrecompileContract(sk *stakingkeeper.Keeper) *Contract {
	return &Contract{
		BaseContract: ethprecompile.NewBaseContract(
			generated.StakingModuleMetaData.ABI,
			cosmlib.AccAddressToEthAddress(authtypes.NewModuleAddress(stakingtypes.ModuleName)),
		),
		msgServer: stakingkeeper.NewMsgServerImpl(sk),
		querier:   stakingkeeper.Querier{Keeper: sk},
	}
}

// GetValidators implements the `getValidator(address)` method.
func (c *Contract) GetValidator(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	validatorAddr common.Address,
) ([]any, error) {
	return c.validatorHelper(ctx, sdk.ValAddress(validatorAddr[:]).String())
}

// GetDelegatorValidators implements the `getDelegatorValidators(address)` method.
func (c *Contract) GetDelegatorValidators(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	delegatorAddr common.Address,
) ([]any, error) {
	return c.delegatorValidatorsHelper(ctx, cosmlib.Bech32FromEthAddress(delegatorAddr))
}

// GetDelegation implements `getDelegation(address)` method.
func (c *Contract) GetDelegation(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	delegatorAddress common.Address,
	validatorAddress common.Address,
) ([]any, error) {
	return c.getDelegationHelper(
		ctx, cosmlib.AddressToAccAddress(delegatorAddress), cosmlib.AddressToValAddress(validatorAddress),
	)
}

// GetUnbondingDelegation implements the `getUnbondingDelegation(address,address)` method.
func (c *Contract) GetUnbondingDelegation(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	delegatorAddress common.Address,
	validatorAddress common.Address,
) ([]any, error) {
	return c.getUnbondingDelegationHelper(
		ctx, cosmlib.AddressToAccAddress(delegatorAddress), cosmlib.AddressToValAddress(validatorAddress),
	)
}

// GetRedelegations implements the `getRedelegations(address,address)` method.
func (c *Contract) GetRedelegations(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	delegatorAddress common.Address,
	srcValidator common.Address,
	dstValidator common.Address,
) ([]any, error) {
	return c.getRedelegationsHelper(
		ctx,
		cosmlib.AddressToAccAddress(delegatorAddress),
		cosmlib.AddressToValAddress(srcValidator),
		cosmlib.AddressToValAddress(dstValidator),
	)
}

// Delegate implements the `delegate(address,uint256)` method.
func (c *Contract) Delegate(
	ctx context.Context,
	_ ethprecompile.EVM,
	caller common.Address,
	_ *big.Int,
	validatorAddress common.Address,
	amount *big.Int,
) ([]any, error) {
	return c.delegateHelper(ctx, caller, amount, cosmlib.AddressToValAddress(validatorAddress))
}

// Undelegate implements the `undelegate(address,uint256)` method.
func (c *Contract) Undelegate(
	ctx context.Context,
	_ ethprecompile.EVM,
	caller common.Address,
	_ *big.Int,
	validatorAddress common.Address,
	amount *big.Int,
) ([]any, error) {
	return c.undelegateHelper(ctx, caller, amount, cosmlib.AddressToValAddress(validatorAddress))
}

// BeginRedelegate implements the `beginRedelegate(address,address,uint256)` method.
func (c *Contract) BeginRedelegate(
	ctx context.Context,
	_ ethprecompile.EVM,
	caller common.Address,
	_ *big.Int,
	srcValidator common.Address,
	dstValidator common.Address,
	amount *big.Int,
) ([]any, error) {
	return c.beginRedelegateHelper(
		ctx,
		caller,
		amount,
		cosmlib.AddressToValAddress(srcValidator),
		cosmlib.AddressToValAddress(dstValidator),
	)
}

// CancelRedelegate implements the `cancelRedelegate(address,address,uint256,int64)` method.
func (c *Contract) CancelUnbondingDelegation(
	ctx context.Context,
	_ ethprecompile.EVM,
	caller common.Address,
	_ *big.Int,
	validatorAddress common.Address,
	amount *big.Int,
	creationHeight int64,
) ([]any, error) {
	return c.cancelUnbondingDelegationHelper(
		ctx, caller, amount, cosmlib.AddressToValAddress(validatorAddress), creationHeight)
}

// GetActiveValidators implements the `getActiveValidators()` method.
func (c *Contract) GetActiveValidators(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	_ ...any,
) ([]any, error) {
	return c.activeValidatorsHelper(ctx)
}

// GetValidators implements the `getValidators()` method.
func (c *Contract) GetValidators(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	_ ...any,
) ([]any, error) {
	return c.validatorsHelper(ctx)
}

// GetValidators implements the `getValidator(address)` method.
func (c *Contract) GetValidatorAddrInput(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	args ...any,
) ([]any, error) {
	val, ok := utils.GetAs[common.Address](args[0])
	if !ok {
		return nil, precompile.ErrInvalidHexAddress
	}

	return c.validatorHelper(ctx, sdk.ValAddress(val[:]).String())
}

// GetDelegatorValidatorsAddrInput implements the `getDelegatorValidators(address)` method.
func (c *Contract) GetDelegatorValidatorsAddrInput(
	ctx context.Context,
	_ ethprecompile.EVM,
	_ common.Address,
	_ *big.Int,
	args ...any,
) ([]any, error) {
	del, ok := utils.GetAs[common.Address](args[0])
	if !ok {
		return nil, precompile.ErrInvalidHexAddress
	}

	return c.delegatorValidatorsHelper(ctx, cosmlib.Bech32FromEthAddress(del))
}
