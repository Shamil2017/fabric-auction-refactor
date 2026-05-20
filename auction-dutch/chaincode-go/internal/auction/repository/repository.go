package repository

import (
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"

	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/domain"
)

type AuctionRepository interface {
	GetAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*domain.Auction, error)
	SaveAuction(ctx contractapi.TransactionContextInterface, auctionID string, auction *domain.Auction) error
	SetAuctionEndorsement(ctx contractapi.TransactionContextInterface, auctionID string, mspids []string, auditor bool) error

	GetTxID(ctx contractapi.TransactionContextInterface) string
	CreateBidKey(ctx contractapi.TransactionContextInterface, auctionID string, txID string) (string, error)

	SavePrivateBid(ctx contractapi.TransactionContextInterface, collection string, bidKey string, bidJSON []byte) error
	GetPrivateBidHash(ctx contractapi.TransactionContextInterface, collection string, bidKey string) ([]byte, error)
	GetPrivateBid(ctx contractapi.TransactionContextInterface, collection string, bidKey string) (*domain.FullBid, error)
}
