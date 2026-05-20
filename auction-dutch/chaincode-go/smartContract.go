/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"

	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/handler"
	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/repository"
	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/service"
)

func main() {
	auctionRepository := repository.NewFabricRepository()
	auctionService := service.NewAuctionService(auctionRepository)
	auctionContract := handler.NewContract(auctionService)

	auctionChaincode, err := contractapi.NewChaincode(auctionContract)
	if err != nil {
		log.Panicf("Error creating auction chaincode: %v", err)
	}

	if err := auctionChaincode.Start(); err != nil {
		log.Panicf("Error starting auction chaincode: %v", err)
	}
}
