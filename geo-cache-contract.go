/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha1"
	"encoding/hex"
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

func myHash(s string) string {
	n := 1
	for n < 100 {
		h := sha1.New()
		h.Write([]byte(s))
		bs := h.Sum(nil)
		s = string(bs)
		n++
	}

	return hex.EncodeToString([]byte(s))
}

func generateSalt() string {
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	salt := make([]rune, 8)
	for i := range salt {
		salt[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(salt)
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
	geoCache.Owner.Salt = generateSalt()
	geoCache.Owner.Id = myHash(user.Id + geoCache.Owner.Salt)
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
func (c *GeoCacheContract) UpdateGeoCache(ctx contractapi.TransactionContextInterface, user User, geoCacheId string, newName string, newDescription string) error {
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

	if geoCache.Owner.Id != myHash(user.Id+geoCache.Owner.Salt) {
		return fmt.Errorf("Only the owner can update a cache!")
	}

	geoCache.Name = newName
	geoCache.Description = newDescription

	newBytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheId, newBytes)
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

	bytes, _ := ctx.GetStub().GetState(geoCacheId)

	geoCache := new(GeoCache)

	err = json.Unmarshal(bytes, geoCache)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal world state data to type GeoCache")
	}

	cacheTrackable := geoCache.Trackable
	geoCache.Trackable = trackable

	newBytes, _ := json.Marshal(geoCache)

	return &cacheTrackable, ctx.GetStub().PutState(geoCacheId, newBytes)
}

// UpdateGeoCache retrieves two list of new koordinates of GeoCache from the world state and updates its value
func (c *GeoCacheContract) UpdateCoordGeoCache(ctx contractapi.TransactionContextInterface, user User, geoCacheId string, newXcoordRange [2]int, newYcoordRange [2]int) error {
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

	if geoCache.Owner.Id != myHash(user.Id+geoCache.Owner.Salt) {
		return fmt.Errorf("Only the owner can update a cache!")
	}

	geoCache.XcoordRange = newXcoordRange
	geoCache.YcoordRange = newYcoordRange

	newBytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheId, newBytes)
}

// DeleteGeoCache deletes an instance of GeoCache from the world state
func (c *GeoCacheContract) DeleteGeoCache(ctx contractapi.TransactionContextInterface, user User, geoCacheId string) error {
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

	if geoCache.Owner.Id != myHash(user.Id+geoCache.Owner.Salt) {
		return fmt.Errorf("Only the owner can update a cache!")
	}

	return ctx.GetStub().DelState(geoCacheId)
}

//ReportGeoCache make a report for a cache
func (c *GeoCacheContract) ReportGeoCache(ctx contractapi.TransactionContextInterface, user User, message string, geoCacheId string) error {
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

	report := new(Report)
	report.Message = message
	report.Notifier = user

	geoCache.Reports = append(geoCache.Reports, *report)

	newBytes, _ := json.Marshal(geoCache)

	return ctx.GetStub().PutState(geoCacheId, newBytes)
}

func (c *GeoCacheContract) GetReports(ctx contractapi.TransactionContextInterface, user User, geoCacheId string) ([]Report, error) {
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

	if geoCache.Owner.Id != myHash(user.Id+geoCache.Owner.Salt) {
		return nil, fmt.Errorf("Only the owner can get the reports!")
	}
	return geoCache.Reports, nil
}
