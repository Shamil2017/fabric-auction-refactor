package service

import (
	"errors"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"

	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/domain"
	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/repository"
)

type AuctionService struct {
	repo repository.AuctionRepository
}

func NewAuctionService(repo repository.AuctionRepository) *AuctionService {
	return &AuctionService{repo: repo}
}

func (s *AuctionService) QueryAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
) (*domain.Auction, error) {
	if strings.TrimSpace(auctionID) == "" {
		return nil, errors.New("auctionID is required")
	}

	return s.repo.GetAuction(ctx, auctionID)
}

func (s *AuctionService) CreateAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	itemSold string,
	quantity int,
	withAuditor string,
	sellerID string,
	sellerOrgID string,
) error {
	auction, err := domain.NewAuction(
		auctionID,
		itemSold,
		quantity,
		sellerID,
		sellerOrgID,
		withAuditor,
	)
	if err != nil {
		return err
	}

	if err := s.repo.SaveAuction(ctx, auctionID, auction); err != nil {
		return err
	}

	if err := s.repo.SetAuctionEndorsement(ctx, auctionID, auction.Orgs, auction.Auditor); err != nil {
		return err
	}

	return nil
}
