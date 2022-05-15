/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const getStateError = "world state get error"

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (ms *MockStub) GetState(key string) ([]byte, error) {
	args := ms.Called(key)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) PutState(key string, value []byte) error {
	args := ms.Called(key, value)

	return args.Error(0)
}

func (ms *MockStub) DelState(key string) error {
	args := ms.Called(key)

	return args.Error(0)
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()

	return args.Get(0).(*MockStub)
}

func configureStub() (*MockContext, *MockStub) {
	var nilBytes []byte

	testGeoCache := new(GeoCache)
	testGeoCache.Name = "set value"
	geoCacheBytes, _ := json.Marshal(testGeoCache)

	ms := new(MockStub)
	ms.On("GetState", "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetState", "missingkey").Return(nilBytes, nil)
	ms.On("GetState", "existingkey").Return([]byte("some value"), nil)
	ms.On("GetState", "geoCachekey").Return(geoCacheBytes, nil)
	ms.On("PutState", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("DelState", mock.AnythingOfType("string")).Return(nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)

	return mc, ms
}

func TestGeoCacheExists(t *testing.T) {
	var exists bool
	var err error

	ctx, _ := configureStub()
	c := new(GeoCacheContract)

	exists, err = c.GeoCacheExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false on error")

	exists, err = c.GeoCacheExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	exists, err = c.GeoCacheExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}

// u := new(User)
// 	u.Id = "123"
// 	u.Name = "TestUser"
// 	assert.True(t, (c.CreateGeoCache(ctx, *u, "testId", "testName", "testDescription", [2]int{5, 10}, [2]int{5, 10}) != nil), "mitkellideirni")

// func TestCreateGeoCache(t *testing.T) {
// 	var err error

// 	ctx, stub := configureStub()
// 	c := new(GeoCacheContract)
// 	u := new(User)
// 	u.Id = "123"
// 	u.Name = "TestUser"

// 	err = c.CreateGeoCache(ctx, *u, "statebad", "some value", "testDescription", [2]int{5, 10}, [2]int{5, 10})
// 	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

// 	err = c.CreateGeoCache(ctx, *u, "existingkey", "some value", "testDescription", [2]int{5, 10}, [2]int{5, 10})
// 	assert.EqualError(t, err, "The asset existingkey already exists", "should error when exists returns true")

// 	err = c.CreateGeoCache(ctx, *u, "missingkey", "some value", "testDescription", [2]int{5, 10}, [2]int{5, 10})
// 	stub.AssertCalled(t, "PutState", "missingkey", []byte("{\"value\":\"some value\"}"))

// 	//TODO: implement more asserts for coordinate validation
// }

// func TestReadGeoCache(t *testing.T) {
// 	var geoCache *GeoCache
// 	var err error

// 	ctx, _ := configureStub()
// 	c := new(GeoCacheContract)

// 	geoCache, err = c.ReadGeoCache(ctx, "statebad", 7, 7)
// 	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when reading")
// 	assert.Nil(t, geoCache, "should not return GeoCache when exists errors when reading")

// 	geoCache, err = c.ReadGeoCache(ctx, "missingkey", 7, 7)
// 	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when reading")
// 	assert.Nil(t, geoCache, "should not return GeoCache when key does not exist in world state when reading")

// 	geoCache, err = c.ReadGeoCache(ctx, "existingkey", 7, 7)
// 	assert.EqualError(t, err, "Could not unmarshal world state data to type GeoCache", "should error when data in key is not GeoCache")
// 	assert.Nil(t, geoCache, "should not return GeoCache when data in key is not of type GeoCache")

// 	geoCache, err = c.ReadGeoCache(ctx, "geoCachekey", 7, 7)
// 	expectedGeoCache := new(GeoCache)
// 	expectedGeoCache.Value = "set value"
// 	assert.Nil(t, err, "should not return error when GeoCache exists in world state when reading")
// 	assert.Equal(t, expectedGeoCache, geoCache, "should return deserialized GeoCache from world state")
// }

// func TestUpdateGeoCache(t *testing.T) {
// 	var err error

// 	ctx, stub := configureStub()
// 	c := new(GeoCacheContract)

// 	err = c.UpdateGeoCache(ctx, "statebad", "new value")
// 	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when updating")

// 	err = c.UpdateGeoCache(ctx, "missingkey", "new value")
// 	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when updating")

// 	err = c.UpdateGeoCache(ctx, "geoCachekey", "new value")
// 	expectedGeoCache := new(GeoCache)
// 	expectedGeoCache.Value = "new value"
// 	expectedGeoCacheBytes, _ := json.Marshal(expectedGeoCache)
// 	assert.Nil(t, err, "should not return error when GeoCache exists in world state when updating")
// 	stub.AssertCalled(t, "PutState", "geoCachekey", expectedGeoCacheBytes)
// }

// func TestDeleteGeoCache(t *testing.T) {
// 	var err error

// 	ctx, stub := configureStub()
// 	c := new(GeoCacheContract)

// 	err = c.DeleteGeoCache(ctx, "statebad")
// 	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

// 	err = c.DeleteGeoCache(ctx, "missingkey")
// 	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

// 	err = c.DeleteGeoCache(ctx, "geoCachekey")
// 	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")
// 	stub.AssertCalled(t, "DelState", "geoCachekey")
// }
