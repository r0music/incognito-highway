// Code generated by mockery v1.0.0. DO NOT EDIT.

package chain_test

import (
	chain "highway/chain"
	common "highway/common"

	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// GetBlockByHash provides a mock function with given fields: ctx, req, hashes
func (_m *Provider) GetBlockByHash(ctx context.Context, req chain.GetBlockByHashRequest, hashes [][]byte) ([][]byte, error) {
	ret := _m.Called(ctx, req, hashes)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func(context.Context, chain.GetBlockByHashRequest, [][]byte) [][]byte); ok {
		r0 = rf(ctx, req, hashes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, chain.GetBlockByHashRequest, [][]byte) error); ok {
		r1 = rf(ctx, req, hashes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetBlockByHeight provides a mock function with given fields: ctx, req, heights, blocks
func (_m *Provider) SetBlockByHeight(ctx context.Context, req chain.GetBlockByHeightRequest, heights []uint64, blocks [][]byte) error {
	ret := _m.Called(ctx, req, heights, blocks)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, chain.GetBlockByHeightRequest, []uint64, [][]byte) error); ok {
		r0 = rf(ctx, req, heights, blocks)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSingleBlockByHeight provides a mock function with given fields: ctx, req, data
func (_m *Provider) SetSingleBlockByHeight(ctx context.Context, req chain.RequestBlockByHeight, data common.ExpectedBlkByHeight) error {
	ret := _m.Called(ctx, req, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, chain.RequestBlockByHeight, common.ExpectedBlkByHeight) error); ok {
		r0 = rf(ctx, req, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StreamBlkByHeight provides a mock function with given fields: ctx, req, blkChan
func (_m *Provider) StreamBlkByHeight(ctx context.Context, req chain.RequestBlockByHeight, blkChan chan common.ExpectedBlkByHeight) error {
	ret := _m.Called(ctx, req, blkChan)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, chain.RequestBlockByHeight, chan common.ExpectedBlkByHeight) error); ok {
		r0 = rf(ctx, req, blkChan)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
