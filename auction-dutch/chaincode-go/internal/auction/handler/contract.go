package handler

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
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

func (c *Contract) Bid(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
) (string, error) {
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("get transient data: %w", err)
	}

	bidJSON, ok := transientMap["bid"]
	if !ok {
		return "", errors.New("bid key not found in transient map")
	}

	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get client MSP ID: %w", err)
	}

	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get peer MSP ID: %w", err)
	}

	return c.auctionService.Bid(
		ctx,
		auctionID,
		bidJSON,
		clientMSPID,
		peerMSPID,
	)
}

func (c *Contract) SubmitBid(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	txID string,
) error {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %w", err)
	}

	return c.auctionService.SubmitBid(
		ctx,
		auctionID,
		txID,
		clientMSPID,
	)
}

func (c *Contract) RevealBid(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	txID string,
) error {
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("get transient data: %w", err)
	}

	bidJSON, ok := transientMap["bid"]
	if !ok {
		return errors.New("bid key not found in transient map")
	}

	clientID, err := getSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity: %w", err)
	}

	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %w", err)
	}

	return c.auctionService.RevealBid(
		ctx,
		auctionID,
		txID,
		bidJSON,
		clientID,
		clientMSPID,
	)
}
