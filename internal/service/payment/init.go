package payment

import (
	"context"

	"github.com/dilmurodov/online_banking/config"
	"github.com/dilmurodov/online_banking/pkg/logger"
	"github.com/dilmurodov/online_banking/pkg/models"
	"github.com/dilmurodov/online_banking/storage"
)

type ServiceI interface {
	CaptureTransactions(ctx context.Context, req *models.CaptureTransactionsRequest) error
	WithDrawal(ctx context.Context, req *models.WithDrawalRequest) (*models.WithDrawalResponse, error)
	Transfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResponse, error)
	Deposit(ctx context.Context, req *models.DepositRequest) (*models.DepositResponse, error)
}

type Service struct {
	cfg  config.Config
	log  logger.LoggerI
	strg storage.StorageI
}

func NewService(cfg config.Config, log logger.LoggerI, strg storage.StorageI) ServiceI {
	return &Service{
		cfg:  cfg,
		log:  log,
		strg: strg,
	}
}
