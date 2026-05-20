/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"go.uber.org/fx"

	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/handler"
	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/repository"
	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/service"
)

func newAuctionChaincode(contract *handler.Contract) (*contractapi.ContractChaincode, error) {
	chaincode, err := contractapi.NewChaincode(contract)
	if err != nil {
		return nil, fmt.Errorf("create auction chaincode: %w", err)
	}

	return chaincode, nil
}

func registerChaincodeLifecycle(
	lifecycle fx.Lifecycle,
	chaincode *contractapi.ContractChaincode,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting auction chaincode with Uber FX")

			go func() {
				if err := chaincode.Start(); err != nil {
					log.Printf("Error starting auction chaincode: %v", err)
					os.Exit(1)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping auction chaincode")
			return nil
		},
	})
}

func main() {
	app := fx.New(
		fx.Provide(
			fx.Annotate(
				repository.NewFabricRepository,
				fx.As(new(repository.AuctionRepository)),
			),
			service.NewAuctionService,
			handler.NewContract,
			newAuctionChaincode,
		),
		fx.Invoke(registerChaincodeLifecycle),
	)

	app.Run()
}
