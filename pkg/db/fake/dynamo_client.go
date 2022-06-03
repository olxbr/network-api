// Code generated by mockery v2.12.3. DO NOT EDIT.

package fake

import (
	context "context"

	dynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	mock "github.com/stretchr/testify/mock"
)

// DynamoClient is an autogenerated mock type for the DynamoClient type
type DynamoClient struct {
	mock.Mock
}

// DeleteItem provides a mock function with given fields: ctx, params, optFns
func (_m *DynamoClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *dynamodb.DeleteItemOutput
	if rf, ok := ret.Get(0).(func(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) *dynamodb.DeleteItemOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dynamodb.DeleteItemOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetItem provides a mock function with given fields: ctx, params, optFns
func (_m *DynamoClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *dynamodb.GetItemOutput
	if rf, ok := ret.Get(0).(func(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) *dynamodb.GetItemOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dynamodb.GetItemOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutItem provides a mock function with given fields: ctx, params, optFns
func (_m *DynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *dynamodb.PutItemOutput
	if rf, ok := ret.Get(0).(func(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) *dynamodb.PutItemOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dynamodb.PutItemOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Scan provides a mock function with given fields: _a0, _a1, _a2
func (_m *DynamoClient) Scan(_a0 context.Context, _a1 *dynamodb.ScanInput, _a2 ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *dynamodb.ScanOutput
	if rf, ok := ret.Get(0).(func(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) *dynamodb.ScanOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dynamodb.ScanOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewDynamoClientT interface {
	mock.TestingT
	Cleanup(func())
}

// NewDynamoClient creates a new instance of DynamoClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDynamoClient(t NewDynamoClientT) *DynamoClient {
	mock := &DynamoClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
