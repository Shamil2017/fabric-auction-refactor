package service

import (
	"errors"
	"fmt"
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

func (s *AuctionService) Bid(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	bidJSON []byte,
	clientMSPID string,
	peerMSPID string,
) (string, error) {
	if strings.TrimSpace(auctionID) == "" {
		return "", errors.New("auctionID is required")
	}

	if len(bidJSON) == 0 {
		return "", errors.New("bid is required")
	}

	if clientMSPID != peerMSPID {
		return "", fmt.Errorf(
			"client from org %s is not authorized to write private data on peer from org %s",
			clientMSPID,
			peerMSPID,
		)
	}

	collection := "_implicit_org_" + clientMSPID

	txID := s.repo.GetTxID(ctx)

	bidKey, err := s.repo.CreateBidKey(ctx, auctionID, txID)
	if err != nil {
		return "", err
	}

	if err := s.repo.SavePrivateBid(ctx, collection, bidKey, bidJSON); err != nil {
		return "", err
	}

	return txID, nil
}

func (s *AuctionService) SubmitBid(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	txID string,
	clientMSPID string,
) error {
	if strings.TrimSpace(auctionID) == "" {
		return errors.New("auctionID is required")
	}

	if strings.TrimSpace(txID) == "" {
		return errors.New("txID is required")
	}

	auction, err := s.repo.GetAuction(ctx, auctionID)
	if err != nil {
		return err
	}

	if auction.Status != domain.StatusOpen {
		return errors.New("cannot join closed or ended auction")
	}

	collection := "_implicit_org_" + clientMSPID

	bidKey, err := s.repo.CreateBidKey(ctx, auctionID, txID)
	if err != nil {
		return err
	}

	bidHash, err := s.repo.GetPrivateBidHash(ctx, collection, bidKey)
	if err != nil {
		return err
	}

	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	auction.PrivateBids[bidKey] = domain.BidHash{
		Org:  clientMSPID,
		Hash: fmt.Sprintf("%x", bidHash),
	}

	if !containsString(auction.Orgs, clientMSPID) {
		auction.Orgs = append(auction.Orgs, clientMSPID)

		if err := s.repo.SetAuctionEndorsement(ctx, auctionID, auction.Orgs, auction.Auditor); err != nil {
			return err
		}
	}

	if err := s.repo.SaveAuction(ctx, auctionID, auction); err != nil {
		return err
	}

	return nil
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}

	return false
}
