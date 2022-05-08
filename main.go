/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func main() {
	geoCacheContract := new(GeoCacheContract)
	geoCacheContract.Info.Version = "0.0.1"
	geoCacheContract.Info.Description = "My Smart Contract"
	geoCacheContract.Info.License = new(metadata.LicenseMetadata)
	geoCacheContract.Info.License.Name = "Apache-2.0"
	geoCacheContract.Info.Contact = new(metadata.ContactMetadata)
	geoCacheContract.Info.Contact.Name = "John Doe"

	chaincode, err := contractapi.NewChaincode(geoCacheContract)
	chaincode.Info.Title = "GeoCache chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		panic("Could not create chaincode from GeoCacheContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
