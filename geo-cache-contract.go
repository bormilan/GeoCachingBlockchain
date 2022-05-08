/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// GeoCacheContract contract for managing CRUD for GeoCache
type GeoCacheContract struct {
	contractapi.Contract
}

// GeoCacheExists returns true when asset with given ID exists in world state
func (c *GeoCacheContract) GeoCacheExists(ctx contractapi.TransactionContextInterface, geoCacheID string) (bool, error) {
	data, err := ctx.GetStub().GetState(geoCacheID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateGeoCache creates a new instance of GeoCache
func (c *GeoCacheContract) CreateGeoCache(ctx contractapi.TransactionContextInterface, geoCacheID string, value string, newXcoordRange [2]int, newYcoordRange [2]int) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("The asset %s already exists", geoCacheID)
	}

	//create object
	geoCache := new(GeoCache)
	geoCache.Value = value
	geoCache.XcoordRange = newXcoordRange
	geoCache.YcoordRange = newYcoordRange

	bytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheID, bytes)
}

// ReadGeoCache retrieves an instance of GeoCache from the world state
func (c *GeoCacheContract) ReadGeoCache(ctx contractapi.TransactionContextInterface, geoCacheID string, Xcoord int, Ycoord int) (*GeoCache, error) {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	bytes, _ := ctx.GetStub().GetState(geoCacheID)

	geoCache := new(GeoCache)

	err = json.Unmarshal(bytes, geoCache)

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal world state data to type GeoCache")
	}

	Xin := Xcoord > geoCache.XcoordRange[0] && Xcoord < geoCache.XcoordRange[1]
	Yin := Ycoord > geoCache.YcoordRange[0] && Ycoord < geoCache.YcoordRange[1]

	if !Xin || !Yin {
		return nil, fmt.Errorf("You are not in the cache's location range!")
	}

	return geoCache, nil
}

// UpdateGeoCache retrieves an instance of GeoCache from the world state and updates its value
func (c *GeoCacheContract) UpdateGeoCache(ctx contractapi.TransactionContextInterface, geoCacheID string, newValue string) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	geoCache := new(GeoCache)
	geoCache.Value = newValue

	bytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheID, bytes)
}

// UpdateGeoCache retrieves two list of new koordinates of GeoCache from the world state and updates its value
func (c *GeoCacheContract) UpdateCoordGeoCache(ctx contractapi.TransactionContextInterface, geoCacheID string, newValue string, newXcoordRange [2]int, newYcoordRange [2]int) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	//create object
	geoCache := new(GeoCache)
	geoCache.XcoordRange = newXcoordRange
	geoCache.YcoordRange = newYcoordRange

	bytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheID, bytes)
}

// DeleteGeoCache deletes an instance of GeoCache from the world state
func (c *GeoCacheContract) DeleteGeoCache(ctx contractapi.TransactionContextInterface, geoCacheID string) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	return ctx.GetStub().DelState(geoCacheID)
}
