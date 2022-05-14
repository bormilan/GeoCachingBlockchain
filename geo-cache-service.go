package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type GeoCacheService struct {
	contract GeoCacheContract
}

func (s *GeoCacheService) LogUserInCache(ctx contractapi.TransactionContextInterface, user User, cacheId string, Xcoord int, Ycoord int, trackable Trackable) (*Trackable, error) {

	err := s.contract.AddVisitorToGeoCache(ctx, user, cacheId, Xcoord, Ycoord)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	newTrackable, err2 := s.contract.SwitchTrackable(ctx, trackable, cacheId)
	if err2 != nil {
		return nil, fmt.Errorf(err2.Error())
	}

	return newTrackable, nil
}
