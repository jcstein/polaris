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

package precompile

import (
	"errors"
	"fmt"
	"reflect"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"pkg.berachain.dev/polaris/eth/core/vm"
	errorslib "pkg.berachain.dev/polaris/lib/errors"
	"pkg.berachain.dev/polaris/lib/utils"
)

const (
	// container impl names stored as constants, to be used in error messages.
	statelessContainerName = `StatelessContainerImpl`
	statefulContainerName  = `StatefulContainerImpl`
)

// AbstractFactory is an interface that all precompile container factories must adhere to.
type AbstractFactory interface {
	// Build builds and returns the precompile container for the type of container/factory.
	Build(Registrable, Plugin) (vm.PrecompileContainer, error)
}

// Compile-time assertions to ensure these container factories adhere to `AbstractFactory`.
var (
	_ AbstractFactory = (*StatelessFactory)(nil)
	_ AbstractFactory = (*StatefulFactory)(nil)
)

// ===========================================================================
// Stateless Container Factory
// ===========================================================================

// StatelessFactory is used to build stateless precompile containers.
type StatelessFactory struct{}

// NewStatelessFactory creates and returns a new `StatelessFactory`.
func NewStatelessFactory() *StatelessFactory {
	return &StatelessFactory{}
}

// Build returns a stateless precompile container for the given base contract implememntation.
// This function will return an error if the given contract is not a stateless implementation.
//
// Build implements `AbstractFactory`.
func (sf *StatelessFactory) Build(
	rp Registrable, _ Plugin,
) (vm.PrecompileContainer, error) {
	pc, ok := utils.GetAs[StatelessImpl](rp)
	if !ok {
		return nil, errorslib.Wrap(ErrWrongContainerFactory, statelessContainerName)
	}
	return pc, nil
}

// ===========================================================================
// Stateful Container Factory
// ===========================================================================

// StatefulFactory is used to build stateful precompile containers.
type StatefulFactory struct{}

// NewStatefulFactory creates and returns a new `StatefulFactory`.
func NewStatefulFactory() *StatefulFactory {
	return &StatefulFactory{}
}

// Build returns a stateful precompile container for the given base contract implementation. This
// function will return an error if the given contract is not a stateful implementation.
//
// Build implements `AbstractFactory`.
func (sf *StatefulFactory) Build(
	rp Registrable, p Plugin,
) (vm.PrecompileContainer, error) {
	sci, ok := utils.GetAs[StatefulImpl](rp)
	if !ok {
		return nil, errorslib.Wrap(ErrWrongContainerFactory, statefulContainerName)
	}

	// attach the precompile plugin to the stateful contract
	sci.SetPlugin(p)

	// add precompile methods to stateful container, if any exist
	idsToMethods, err := buildIdsToMethods(sci.ABIMethods(), reflect.ValueOf(sci))
	if err != nil {
		return nil, err
	}

	return NewStateful(rp, idsToMethods)
}

// This function matches each Go implementation of the precompile to the ABI's respective function.
// It searches for the ABI function in the Go precompile contract and performs basic validation on
// the implemented function.
func buildIdsToMethods(pcABI map[string]abi.Method, contractImpl reflect.Value) (map[string]*Method, error) {
	contractImplType := contractImpl.Type()
	idsToMethods := make(map[string]*Method)
	for m := 0; m < contractImplType.NumMethod(); m++ {
		implMethod := contractImplType.Method(m)
		implMethodName := formatName(implMethod.Name)

		if abiMethod, found := pcABI[implMethodName]; found {
			if err := validateReturnTypes(implMethod); err != nil {
				return nil, errorslib.Wrap(err, implMethodName)
			}
			idsToMethods[utils.UnsafeBytesToStr(abiMethod.ID)] = NewMethod(
				&abiMethod,
				abiMethod.Sig,
				implMethod.Func,
			)
		}
	}

	// verify that every abi method has a corresponding precompile implementation
	for _, abiMethod := range pcABI {
		if _, found := idsToMethods[utils.UnsafeBytesToStr(abiMethod.ID)]; !found {
			return nil, errorslib.Wrap(ErrNoPrecompileMethodForABIMethod, abiMethod.Name)
		}
	}

	return idsToMethods, nil
}

// formatName converts to first character of name to lowercase.
func formatName(name string) string {
	if len(name) == 0 {
		return name
	}

	ret := []rune(name)
	ret[0] = unicode.ToLower(ret[0])
	return string(ret)
}

// validateReturnTypes checks if the precompile method returns a []any and an error. Returns an
// error otherwise.
func validateReturnTypes(implMethod reflect.Method) error {
	if implMethod.Type.NumOut() != 2 { //nolint:gomnd // always expect 2 return values.
		return errors.New(
			"precompile methods must return ([]any, error), but found wrong number of return types for precompile method", //nolint:lll // it's okay.
		)
	}

	// check if the first return type is of type []any
	firstReturnType := implMethod.Type.Out(0)
	if firstReturnType.Kind() != reflect.Slice && firstReturnType.Elem().Kind() != reflect.Interface {
		return fmt.Errorf(
			"first parameter should be []any, but found %s for precompile method",
			firstReturnType.String(),
		)
	}

	// check that the second return value is an error
	secondReturnType := implMethod.Type.Out(1)
	if secondReturnType.Name() != "error" {
		return fmt.Errorf(
			"second parameter should be error, but found %s for precompile method",
			secondReturnType.String(),
		)
	}

	return nil
}
