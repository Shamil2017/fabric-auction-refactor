package handler

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"

	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/domain"
	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/service"
)

type Contract struct {
	contractapi.Contract
	auctionService *service.AuctionService
}

func NewContract(auctionService *service.AuctionService) *Contract {
	return &Contract{
		auctionService: auctionService,
	}
}

func (c *Contract) QueryAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
) (*domain.Auction, error) {
	return c.auctionService.QueryAuction(ctx, auctionID)
}

func (c *Contract) CreateAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	itemSold string,
	quantity int,
	withAuditor string,
) error {
	clientID, err := getSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity: %w", err)
	}

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %w", err)
	}

	return c.auctionService.CreateAuction(
		ctx,
		auctionID,
		itemSold,
		quantity,
		withAuditor,
		clientID,
		clientOrgID,
	)
}
