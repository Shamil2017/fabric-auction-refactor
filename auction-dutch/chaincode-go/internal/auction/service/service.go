package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

func (s *AuctionService) RevealBid(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	txID string,
	bidJSON []byte,
	clientID string,
	clientMSPID string,
) error {
	if strings.TrimSpace(auctionID) == "" {
		return errors.New("auctionID is required")
	}

	if strings.TrimSpace(txID) == "" {
		return errors.New("txID is required")
	}

	if len(bidJSON) == 0 {
		return errors.New("bid is required")
	}

	collection := "_implicit_org_" + clientMSPID

	bidKey, err := s.repo.CreateBidKey(ctx, auctionID, txID)
	if err != nil {
		return err
	}

	privateBidHash, err := s.repo.GetPrivateBidHash(ctx, collection, bidKey)
	if err != nil {
		return err
	}

	if privateBidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	auction, err := s.repo.GetAuction(ctx, auctionID)
	if err != nil {
		return err
	}

	if auction.Status != domain.StatusClosed {
		return errors.New("cannot reveal bid for open or ended auction")
	}

	calculatedHash := sha256.Sum256(bidJSON)

	if !bytes.Equal(calculatedHash[:], privateBidHash) {
		return fmt.Errorf("revealed bid hash does not match private bid hash")
	}

	publicBidHash, ok := auction.PrivateBids[bidKey]
	if !ok {
		return fmt.Errorf("bid hash was not submitted to auction: %s", bidKey)
	}

	if publicBidHash.Hash != fmt.Sprintf("%x", privateBidHash) {
		return errors.New("private bid hash does not match public auction hash")
	}

	var bidInput struct {
		Quantity int    `json:"quantity"`
		Price    int    `json:"price"`
		Org      string `json:"org"`
		Buyer    string `json:"buyer"`
	}

	if err := json.Unmarshal(bidJSON, &bidInput); err != nil {
		return fmt.Errorf("unmarshal bid: %w", err)
	}

	if bidInput.Buyer != clientID {
		return fmt.Errorf("permission denied, client is not the owner of the bid")
	}

	auction.RevealedBids[bidKey] = domain.FullBid{
		Type:     domain.BidKeyType,
		Quantity: bidInput.Quantity,
		Price:    bidInput.Price,
		Org:      bidInput.Org,
		Buyer:    bidInput.Buyer,
	}

	if err := s.repo.SaveAuction(ctx, auctionID, auction); err != nil {
		return err
	}

	return nil
}

func (s *AuctionService) CloseAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	clientID string,
) error {
	if strings.TrimSpace(auctionID) == "" {
		return errors.New("auctionID is required")
	}

	auction, err := s.repo.GetAuction(ctx, auctionID)
	if err != nil {
		return err
	}

	if auction.Seller != clientID {
		return errors.New("auction can only be closed by seller")
	}

	if auction.Status != domain.StatusOpen {
		return errors.New("cannot close auction that is not open")
	}

	auction.Status = domain.StatusClosed

	if err := s.repo.SaveAuction(ctx, auctionID, auction); err != nil {
		return err
	}

	return nil
}

func (s *AuctionService) EndAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	clientID string,
	peerMSPID string,
) error {
	if strings.TrimSpace(auctionID) == "" {
		return errors.New("auctionID is required")
	}

	auction, err := s.repo.GetAuction(ctx, auctionID)
	if err != nil {
		return err
	}

	if auction.Seller != clientID {
		return errors.New("auction can only be ended by seller")
	}

	if auction.Status != domain.StatusClosed {
		return errors.New("can only end a closed auction")
	}

	if len(auction.RevealedBids) == 0 {
		return errors.New("no bids have been revealed, cannot end auction")
	}

	bidders := make([]domain.FullBid, 0, len(auction.RevealedBids))
	for _, bid := range auction.RevealedBids {
		bidders = append(bidders, bid)
	}

	sort.Slice(bidders, func(i, j int) bool {
		if bidders[i].Price > bidders[j].Price {
			return true
		}

		if bidders[i].Price < bidders[j].Price {
			return false
		}

		return bidders[i].Quantity < bidders[j].Quantity
	})

	auction.Winners = []domain.Winner{}
	remainingQuantity := auction.Quantity

	for i := 0; remainingQuantity > 0 && i < len(bidders); i++ {
		winnerQuantity := bidders[i].Quantity

		if winnerQuantity > remainingQuantity {
			winnerQuantity = remainingQuantity
		}

		auction.Winners = append(auction.Winners, domain.Winner{
			Buyer:    bidders[i].Buyer,
			Quantity: winnerQuantity,
		})

		auction.Price = bidders[i].Price
		remainingQuantity -= winnerQuantity
	}

	if err := s.checkForHigherBid(ctx, auction.Price, auction.RevealedBids, auction.PrivateBids, peerMSPID); err != nil {
		return fmt.Errorf("cannot end auction: %w", err)
	}

	auction.Status = domain.StatusEnded

	if err := s.repo.SaveAuction(ctx, auctionID, auction); err != nil {
		return err
	}

	return nil
}

func (s *AuctionService) checkForHigherBid(
	ctx contractapi.TransactionContextInterface,
	auctionPrice int,
	revealedBids map[string]domain.FullBid,
	privateBids map[string]domain.BidHash,
	peerMSPID string,
) error {
	for bidKey, privateBid := range privateBids {
		if _, alreadyRevealed := revealedBids[bidKey]; alreadyRevealed {
			continue
		}

		collection := "_implicit_org_" + privateBid.Org

		if privateBid.Org == peerMSPID {
			bid, err := s.repo.GetPrivateBid(ctx, collection, bidKey)
			if err != nil {
				return err
			}

			if bid.Price > auctionPrice {
				return errors.New("unrevealed bid has a higher price")
			}
		} else {
			hash, err := s.repo.GetPrivateBidHash(ctx, collection, bidKey)
			if err != nil {
				return err
			}

			if hash == nil {
				return fmt.Errorf("bid hash does not exist: %s", bidKey)
			}
		}
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
