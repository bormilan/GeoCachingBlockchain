/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

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
func (c *GeoCacheContract) CreateGeoCache(ctx contractapi.TransactionContextInterface, user User, geoCacheID string, name string, description string, newXcoordRange [2]int, newYcoordRange [2]int) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("The asset %s already exists", geoCacheID)
	}

	//create object
	geoCache := new(GeoCache)
	geoCache.Name = name
	geoCache.Description = description
	geoCache.XcoordRange = newXcoordRange
	geoCache.YcoordRange = newYcoordRange
	geoCache.Owner = user
	geoCache.Reports = []Report{}
	geoCache.Visitors = []User{}

	trackable := new(Trackable)
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	trackable.Id = string(b)
	trackable.Value = "You are the first, who discovered this cache!"

	geoCache.Trackable = *trackable

	bytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheID, bytes)
}

// ReadGeoCache retrieves an instance of GeoCache from the world state
func (c *GeoCacheContract) ReadGeoCache(ctx contractapi.TransactionContextInterface, geoCacheId string) (*GeoCache, error) {
	exists, err := c.GeoCacheExists(ctx, geoCacheId)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The asset %s does not exist", geoCacheId)
	}

	bytes, _ := ctx.GetStub().GetState(geoCacheId)

	geoCache := new(GeoCache)

	err = json.Unmarshal(bytes, geoCache)

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal world state data to type GeoCache")
	}

	return geoCache, nil
}

// UpdateGeoCache retrieves an instance of GeoCache from the world state and updates its value
func (c *GeoCacheContract) UpdateGeoCache(ctx contractapi.TransactionContextInterface, userId string, geoCacheID string, newName string, newDescription string) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	geoCache := new(GeoCache)
	if geoCache.Owner.Id != userId {
		return fmt.Errorf("Only the owner can update a cache!")
	}

	geoCache.Name = newName
	geoCache.Description = newDescription

	bytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheID, bytes)
}

// UpdateGeoCache retrieves an instance of GeoCache from the world state and updates its value
func (c *GeoCacheContract) AddVisitorToGeoCache(ctx contractapi.TransactionContextInterface, user User, geoCacheId string, Xcoord int, Ycoord int) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheId)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheId)
	}

	bytes, _ := ctx.GetStub().GetState(geoCacheId)

	geoCache := new(GeoCache)

	err = json.Unmarshal(bytes, geoCache)

	if err != nil {
		return fmt.Errorf("Could not unmarshal world state data to type GeoCache")
	}

	Xin := Xcoord > geoCache.XcoordRange[0] && Xcoord < geoCache.XcoordRange[1]
	Yin := Ycoord > geoCache.YcoordRange[0] && Ycoord < geoCache.YcoordRange[1]

	if !Xin || !Yin {
		return fmt.Errorf("You are not in the cache's location range!")
	}

	geoCache.Visitors = append(geoCache.Visitors, user)

	newBytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheId, newBytes)
}

func (c *GeoCacheContract) SwitchTrackable(ctx contractapi.TransactionContextInterface, trackable Trackable, geoCacheId string) (*Trackable, error) {
	exists, err := c.GeoCacheExists(ctx, geoCacheId)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The asset %s does not exist", geoCacheId)
	}

	geoCache := new(GeoCache)
	cacheTrackable := geoCache.Trackable
	geoCache.Trackable = trackable

	bytes, _ := json.Marshal(geoCache)

	return &cacheTrackable, ctx.GetStub().PutState(geoCacheId, bytes)
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
func (c *GeoCacheContract) DeleteGeoCache(ctx contractapi.TransactionContextInterface, user User, geoCacheID string) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	geoCache := new(GeoCache)
	if geoCache.Owner.Id != user.Id {
		return fmt.Errorf("Only the owner can delete a cache!")
	}

	return ctx.GetStub().DelState(geoCacheID)
}

//ReportGeoCache make a report for a cache
func (c *GeoCacheContract) ReportGeoCache(ctx contractapi.TransactionContextInterface, user User, message string, geoCacheID string) error {
	exists, err := c.GeoCacheExists(ctx, geoCacheID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", geoCacheID)
	}

	report := new(Report)
	report.Message = message
	report.Notifier = user

	geoCache := new(GeoCache)
	geoCache.Reports = append(geoCache.Reports, *report)

	bytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheID, bytes)
}
