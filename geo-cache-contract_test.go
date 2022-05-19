/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	u := new(User)
	u.Id = "4ebe56ee0099cc1af664ad67b3410c2b0a18cfba" // result of myHash("123" + "123"), this way it has become testable
	u.Name = "TestUser"
	u.Salt = "123"

	testGeoCache.Owner = *u
	testGeoCache.XcoordRange = [2]int{5, 10}
	testGeoCache.YcoordRange = [2]int{5, 10}

	trackable := new(Trackable)
	testGeoCache.Trackable = *trackable
	testGeoCache.Trackable.Id = "testId"
	testGeoCache.Trackable.Value = "testValue"

	report := new(Report)
	report.Id = "testId"
	report.Message = "TestMessage"
	report.Notifier = *u
	testGeoCache.Reports = append(testGeoCache.Reports, *report)

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

	//statebad returns nilBytes and an error, so the exist function should return with false or error
	exists, err = c.GeoCacheExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false on error")

	//missingkey returns with nilBytes and no error, so the function should return a false, bc the key's format is valid, but not exists
	exists, err = c.GeoCacheExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	//existingkey returns with same valid value, and no error, so the function should return with true bc the object exists
	exists, err = c.GeoCacheExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}

func TestCreateGeoCache(t *testing.T) {
	var err error

	ctx, _ := configureStub()
	c := new(GeoCacheContract)
	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"

	// statebad returns nilBytes and an error, so the function should return with an error
	err = c.CreateGeoCache(ctx, *u, "statebad", "some value", "testDescription", [2]int{5, 10}, [2]int{5, 10}, "asd")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	// existingkey returns with same valid value, and no error, so the function should return with error, bc the key already exist
	err = c.CreateGeoCache(ctx, *u, "existingkey", "some value", "testDescription", [2]int{5, 10}, [2]int{5, 10}, "asd")
	assert.EqualError(t, err, "The asset existingkey already exists", "should error when exists returns true")

	//create a cache with Create function, and assert that, it does not return an error
	err = c.CreateGeoCache(ctx, *u, "missingkey", "some value", "testDescription", [2]int{5, 10}, [2]int{5, 10}, "asd")
	assert.Nil(t, err)
}

func TestReadGeoCache(t *testing.T) {
	var geoCache *GeoCache
	var err error

	ctx, _ := configureStub()
	c := new(GeoCacheContract)

	// statebad returns nilBytes and an error, so the function should return with nil
	geoCache, err = c.ReadGeoCache(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when reading")
	assert.Nil(t, geoCache, "should not return GeoCache when exists errors when reading")

	// missingkey returns with nilBytes and no error, so the function should return true, bc the object does not exists
	geoCache, err = c.ReadGeoCache(ctx, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when reading")
	assert.Nil(t, geoCache, "should not return GeoCache when key does not exist in world state when reading")

	// existingkey returns with same valid value, and no error, so the function should return with error, bc the object does not exists
	geoCache, err = c.ReadGeoCache(ctx, "existingkey")
	assert.EqualError(t, err, "Could not unmarshal world state data to type GeoCache", "should error when data in key is not GeoCache")
	assert.Nil(t, geoCache, "should not return GeoCache when data in key is not of type GeoCache")

	//expected values
	geoCache, err = c.ReadGeoCache(ctx, "geoCachekey")
	expectedGeoCache := new(GeoCache)
	expectedGeoCache.Name = "set value"

	u := new(User)
	u.Id = "4ebe56ee0099cc1af664ad67b3410c2b0a18cfba" // result of myHash("123" + "123"), this way it has become testable
	u.Name = "TestUser"
	u.Salt = "123"

	trackable := new(Trackable)
	expectedGeoCache.Trackable = *trackable
	expectedGeoCache.Trackable.Id = "testId"
	expectedGeoCache.Trackable.Value = "testValue"

	report := new(Report)
	report.Id = "testId"
	report.Message = "TestMessage"
	report.Notifier = *u
	expectedGeoCache.Reports = append(expectedGeoCache.Reports, *report)

	expectedGeoCache.XcoordRange = [2]int{5, 10}
	expectedGeoCache.YcoordRange = [2]int{5, 10}

	expectedGeoCache.Owner = *u

	//does not return error, bc the object exists. and should return woth the expected data
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when reading")
	assert.Equal(t, expectedGeoCache, geoCache, "should return deserialized GeoCache from world state")
}

func TestUpdateGeoCache(t *testing.T) {
	var err error

	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"
	u.Salt = "123"

	ctx, stub := configureStub()
	c := new(GeoCacheContract)

	// statebad returns nilBytes and an error, so the function should return with error
	err = c.UpdateGeoCache(ctx, *u, "statebad", "new value", "newDescription")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when updating")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	err = c.UpdateGeoCache(ctx, *u, "missingkey", "new value", "newDescription")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when updating")

	// existingkey returns with same valid value, and no error, so the function should return with the success, and the object should updated
	err = c.UpdateGeoCache(ctx, *u, "geoCachekey", "new value", "newDescription")
	expectedGeoCache := new(GeoCache)
	expectedGeoCache.Name = "new value"
	expectedGeoCache.Description = "newDescription"

	// expected user in the expected cache
	u2 := new(User)
	u2.Id = "4ebe56ee0099cc1af664ad67b3410c2b0a18cfba"
	u2.Name = "TestUser"
	u2.Salt = "123"

	trackable := new(Trackable)
	expectedGeoCache.Trackable = *trackable
	expectedGeoCache.Trackable.Id = "testId"
	expectedGeoCache.Trackable.Value = "testValue"

	report := new(Report)
	report.Id = "testId"
	report.Message = "TestMessage"
	report.Notifier = *u2
	expectedGeoCache.Reports = append(expectedGeoCache.Reports, *report)

	expectedGeoCache.Owner = *u2
	expectedGeoCache.XcoordRange = [2]int{5, 10}
	expectedGeoCache.YcoordRange = [2]int{5, 10}
	expectedGeoCacheBytes, _ := json.Marshal(expectedGeoCache)

	//does not return an error, because the new user's id and salt combination is equal the stored id hash
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when updating")
	//put state should called with the expected cache value
	stub.AssertCalled(t, "PutState", "geoCachekey", expectedGeoCacheBytes)
}

func TestDeleteGeoCache(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(GeoCacheContract)

	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"
	u.Salt = "123"

	// statebad returns nilBytes and an error, so the function should return with error
	err = c.DeleteGeoCache(ctx, *u, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	err = c.DeleteGeoCache(ctx, *u, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	// geoCachekey returns with a valid value and no error, so the function shouldnt return woth an error and delState should called with the "geoCachekey" value
	err = c.DeleteGeoCache(ctx, *u, "geoCachekey")
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")
	//del state should called
	stub.AssertCalled(t, "DelState", "geoCachekey")
}

func TestAddVisitorToGeoCache(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(GeoCacheContract)

	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"
	u.Salt = "123"

	// statebad returns nilBytes and an error, so the function should return with error
	err = c.AddVisitorToGeoCache(ctx, *u, "statebad", 6, 6)
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	err = c.AddVisitorToGeoCache(ctx, *u, "missingkey", 6, 6)
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	// geoCachekey returns with a valid value and no error, so the function shouldnt return with no error, and the given coordinates is in the cache's range
	err = c.AddVisitorToGeoCache(ctx, *u, "geoCachekey", 6, 6)
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")

	expectedGeoCache := new(GeoCache)
	expectedGeoCache.Name = "set value"

	// expected user in the expected cache
	u2 := new(User)
	u2.Id = "4ebe56ee0099cc1af664ad67b3410c2b0a18cfba"
	u2.Name = "TestUser"
	u2.Salt = "123"

	trackable := new(Trackable)
	expectedGeoCache.Trackable = *trackable
	expectedGeoCache.Trackable.Id = "testId"
	expectedGeoCache.Trackable.Value = "testValue"

	report := new(Report)
	report.Id = "testId"
	report.Message = "TestMessage"
	report.Notifier = *u2
	expectedGeoCache.Reports = append(expectedGeoCache.Reports, *report)

	expectedGeoCache.Owner = *u2
	//adding the new visitor
	expectedGeoCache.Visitors = append(expectedGeoCache.Visitors, *u)
	expectedGeoCache.XcoordRange = [2]int{5, 10}
	expectedGeoCache.YcoordRange = [2]int{5, 10}
	expectedGeoCacheBytes, _ := json.Marshal(expectedGeoCache)

	//put state should called with the expected cache value
	stub.AssertCalled(t, "PutState", "geoCachekey", expectedGeoCacheBytes)
}

func TestSwitchTrackable(t *testing.T) {
	var err error

	ctx, _ := configureStub()
	c := new(GeoCacheContract)

	trackable := new(Trackable)
	trackable.Id = "testId"
	trackable.Value = "testValue"

	// statebad returns nilBytes and an error, so the function should return with error
	_, err = c.SwitchTrackable(ctx, *trackable, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	_, err = c.SwitchTrackable(ctx, *trackable, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	// geoCachekey returns with a valid value and no error, so the function shouldnt return woth an error
	switchedTrackable, err := c.SwitchTrackable(ctx, *trackable, "geoCachekey")
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")

	expectedTrackable := new(Trackable)
	expectedTrackable.Id = "testId"
	expectedTrackable.Value = "testValue"

	assert.Equal(t, switchedTrackable, expectedTrackable)
}

func TestUpdateCoordGeoCache(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(GeoCacheContract)

	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"
	u.Salt = "123"

	// statebad returns nilBytes and an error, so the function should return with error
	err = c.UpdateCoordGeoCache(ctx, *u, "statebad", [2]int{7, 10}, [2]int{7, 10})
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	err = c.UpdateCoordGeoCache(ctx, *u, "missingkey", [2]int{7, 10}, [2]int{7, 10})
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	// geoCachekey returns with a valid value and no error, so the function shouldnt return woth an error
	err = c.UpdateCoordGeoCache(ctx, *u, "geoCachekey", [2]int{7, 10}, [2]int{7, 10})
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")

	expectedGeoCache := new(GeoCache)
	expectedGeoCache.Name = "set value"

	// expected user in the expected cache
	u2 := new(User)
	u2.Id = "4ebe56ee0099cc1af664ad67b3410c2b0a18cfba"
	u2.Name = "TestUser"
	u2.Salt = "123"

	trackable := new(Trackable)
	expectedGeoCache.Trackable = *trackable
	expectedGeoCache.Trackable.Id = "testId"
	expectedGeoCache.Trackable.Value = "testValue"

	report := new(Report)
	report.Id = "testId"
	report.Message = "TestMessage"
	report.Notifier = *u2
	expectedGeoCache.Reports = append(expectedGeoCache.Reports, *report)

	expectedGeoCache.Owner = *u2
	//adding the new visitor
	expectedGeoCache.XcoordRange = [2]int{7, 10}
	expectedGeoCache.YcoordRange = [2]int{7, 10}
	expectedGeoCacheBytes, _ := json.Marshal(expectedGeoCache)

	//put state should called with the expected cache value
	stub.AssertCalled(t, "PutState", "geoCachekey", expectedGeoCacheBytes)
}

func TestReportGeoCache(t *testing.T) {
	var err error

	ctx, _ := configureStub()
	c := new(GeoCacheContract)

	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"
	u.Salt = "123"

	// statebad returns nilBytes and an error, so the function should return with error
	err = c.ReportGeoCache(ctx, *u, "reportMessage", "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	err = c.ReportGeoCache(ctx, *u, "reportMessage", "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	// geoCachekey returns with a valid value and no error, so the function shouldnt return with an error
	err = c.ReportGeoCache(ctx, *u, "reportMessage", "geoCachekey")
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")
}

func TestGetReports(t *testing.T) {
	var err error

	ctx, _ := configureStub()
	c := new(GeoCacheContract)

	u := new(User)
	u.Id = "123"
	u.Name = "TestUser"
	u.Salt = "123"

	// statebad returns nilBytes and an error, so the function should return with error
	_, err = c.GetReports(ctx, *u, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	//missingkey returns with nilBytes and no error, so the function should return error, bc the object does not exists
	_, err = c.GetReports(ctx, *u, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	// geoCachekey returns with a valid value and no error, so the function shouldnt return with an error
	reports, err := c.GetReports(ctx, *u, "geoCachekey")
	assert.Nil(t, err, "should not return error when GeoCache exists in world state when deleting")

	u2 := new(User)
	u2.Id = "4ebe56ee0099cc1af664ad67b3410c2b0a18cfba"
	u2.Name = "TestUser"
	u2.Salt = "123"

	expectedReport := new(Report)
	expectedReport.Id = "testId"
	expectedReport.Message = "TestMessage"
	expectedReport.Notifier = *u2

	assert.Equal(t, *expectedReport, reports[0])
}
