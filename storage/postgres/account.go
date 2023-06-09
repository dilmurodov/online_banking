package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/dilmurodov/online_banking/pkg/customerrors"
	"github.com/dilmurodov/online_banking/pkg/models"
	"github.com/pkg/errors"
)

type accountRepo struct {
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *accountRepo {
	return &accountRepo{db: db}
}

func (r *accountRepo) CreateAccount(ctx context.Context, account *models.CreateAccountRequest) (*models.Account, error) {
	var accountID string
	stmt, err := r.db.PrepareContext(ctx,
		`INSERT INTO accounts (
			user_id, 
			balance
		) VALUES ($1, $2) RETURNING guid`,
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx,
		account.UserID,
		account.Balance,
	)
	err = row.Scan(&accountID)
	if err != nil {
		return nil, &customerrors.InternalServerError{Message: err.Error()}
	}
	return &models.Account{
		ID:      accountID,
		UserID:  account.UserID,
		Balance: account.Balance,
	}, nil
}

func (r *accountRepo) GetAccountByID(ctx context.Context, req *models.GetAccountByIDRequest) (*models.Account, error) {
	var account models.Account

	var (
		createdAt sql.NullString
		updatedAt sql.NullString
	)

	err := r.db.QueryRowContext(ctx,
		`SELECT 
			guid, 
			user_id, 
			balance, 
			created_at,
			updated_at 
		FROM accounts 
		WHERE guid=$1`, req.ID,
	).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&createdAt,
		&updatedAt,
	)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, &customerrors.UserNotFoundError{Guid: req.ID}
	}
	account.CreatedAt = createdAt.String
	account.UpdatedAt = updatedAt.String
	return &account, nil
}

func (r *accountRepo) GetAccountsByUserID(ctx context.Context, req *models.GetAccountsByUserIDRequest) (resp *models.GetAccountsByUserIDResponse, err error) {
	var (
		accounts []*models.Account
		count    int
	)

	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			guid, 
			user_id, 
			balance, 
			created_at,
			updated_at,
			count(1) OVER() AS count
		FROM accounts WHERE user_id=$1 AND deleted_at = 0`, req.UserID)
	if err != nil {
		return nil, &customerrors.InternalServerError{Message: err.Error()}
	}
	defer rows.Close()

	for rows.Next() {
		var a models.Account
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Balance,
			&a.CreatedAt,
			&a.UpdatedAt,
			&count,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &a)
	}
	if err = rows.Err(); err != nil {
		return nil, &customerrors.InternalServerError{Message: err.Error()}
	}

	return &models.GetAccountsByUserIDResponse{
		Accounts: accounts,
		Count:    count,
	}, nil
}

func (r *accountRepo) UpdateAccountBalance(ctx context.Context, tx *sql.Tx, account *models.Account) error {
	stmt, err := tx.PrepareContext(ctx,
		`UPDATE accounts 
		SET balance = $1 WHERE guid=$2`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare update account balance statement")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		account.Balance,
		account.ID,
	)
	// CHeck positive balance constraint
	if err != nil && strings.Contains(err.Error(), "constraint \"positive_balance\"") {
		return &customerrors.InsufficientFundsError{}
	} else if err != nil {
		return &customerrors.InternalServerError{Message: err.Error()}
	}
	return nil
}
