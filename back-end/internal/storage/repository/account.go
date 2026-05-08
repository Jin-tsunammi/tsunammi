package repository

import (
	"context"
	"mm/internal/model"
	"mm/pkg/repository"
	"time"

	"github.com/uptrace/bun"
)

type AccountRepository struct {
	repository.Generic[model.Account, uint64]
}

func NewAccountRepository(genericRepository repository.Generic[model.Account, uint64]) *AccountRepository {
	return &AccountRepository{Generic: genericRepository}
}

func (a *AccountRepository) WithTx(tx bun.Tx) *AccountRepository {
	return &AccountRepository{Generic: a.Generic.WithTx(tx)}
}

func (a *AccountRepository) AddExchangeAccount(ctx context.Context, acc *model.Account) (*model.Account, error) {
	err := a.DB.NewInsert().
		Model(acc).
		Returning("*").
		Scan(ctx, acc)

	if err != nil {
		return nil, err
	}

	var exchangeName string
	err = a.DB.NewSelect().
		Model((*model.Exchange)(nil)).
		Column("name").
		Where("id = ?", acc.ExchangeID).
		Scan(ctx, &exchangeName)

	if err != nil {
		return nil, err
	}

	acc.ExchangeName = exchangeName

	return acc, nil
}

func (a *AccountRepository) UpdateAll(ctx context.Context, accounts []model.Account) error {
	values := a.DB.NewValues(&accounts)

	_, err := a.DB.NewUpdate().
		With("_data", values).
		Model((*model.Account)(nil)).
		TableExpr("_data").
		Set("withdraw_limit = _data.withdraw_limit").
		Where("accounts.id = _data.id").
		Exec(ctx)

	return err
}

func (a *AccountRepository) GetByIDAndUserID(ctx context.Context, accountID, userID uint64) (*model.Account, error) {

	acc := new(accountWithExchange)

	err := a.newAccountByStatus(model.InitializedAccountStatuses).
		Where("a.user_id = ?", userID).
		Where("a.id = ?", accountID).
		Group("a.id", "e.id").
		Scan(ctx, acc)

	if err != nil {
		return nil, err
	}

	account := &acc.Account

	account.ExchangeName = acc.ExchangeName
	account.DepositBalance = acc.DepositBalance

	return account, nil

}

func (a *AccountRepository) FindAllByUserID(ctx context.Context, parsedPage, parsedPageSize int, userID uint64) ([]model.Account, int, error) {

	accs := make([]accountWithExchange, 0)

	query := a.newAccountByStatus(model.InitializedAccountStatuses).
		Where("a.user_id = ?", userID).
		OrderExpr("COALESCE(SUM(d.amount), 0) DESC").
		OrderExpr("a.id DESC").
		Group("a.id", "e.id")

	if parsedPage > 0 && parsedPageSize > 0 {
		query = query.Offset(parsedPageSize * (parsedPage - 1)).Limit(parsedPageSize)
	}

	err := query.Scan(ctx, &accs)
	if err != nil {
		return nil, 0, err
	}

	accounts := make([]model.Account, 0, len(accs))

	for i := 0; i < len(accs); i++ {
		acc := &accs[i]
		account := &acc.Account

		account.ExchangeName = acc.ExchangeName
		account.DepositBalance = acc.DepositBalance

		accounts = append(accounts, *account)
	}

	total, err := a.DB.NewSelect().
		Model((*model.Account)(nil)).
		Where("user_id = ?", userID).
		Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (a *AccountRepository) FindAllOlderThanByStatus(ctx context.Context, t time.Duration, status model.AccountStatus) ([]model.Account, error) {
	accs := make([]accountWithExchange, 0)

	thresholdTime := time.Now().Add(-t)

	err := a.newAccountByStatus([]model.AccountStatus{status}).
		Where("a.created_at < ?", thresholdTime).
		Group("a.id", "e.id").
		Scan(ctx, &accs)

	if err != nil {
		return nil, err
	}

	accounts := make([]model.Account, 0, len(accs))

	for i := 0; i < len(accs); i++ {
		acc := &accs[i]
		account := &acc.Account

		account.ExchangeName = acc.ExchangeName
		account.DepositBalance = acc.DepositBalance

		accounts = append(accounts, *account)
	}

	return accounts, nil
}

func (a *AccountRepository) DeleteByIDAndUserID(ctx context.Context, accountID, userID uint64) error {
	_, err := a.DB.NewDelete().
		Model((*model.Account)(nil)).
		Where("user_id = ? AND id = ?", userID, accountID).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

type accountWithExchange struct {
	model.Account
	ExchangeName   string
	DepositBalance float64
}

func (a *AccountRepository) newAccountByStatus(statuses []model.AccountStatus) *bun.SelectQuery {
	query := a.DB.NewSelect().
		Model((*model.Account)(nil)).
		Column("a.*").
		ColumnExpr("e.name AS exchange_name").
		ColumnExpr("COALESCE(SUM(d.amount), 0) AS deposit_balance").
		Join("LEFT JOIN exchanges AS e ON e.id = a.exchange_id").
		Join("LEFT JOIN deposit_orders AS wo ON wo.account_id = a.id").
		Join("LEFT JOIN deposits AS d ON d.deposit_order_id = wo.id").
		Where("a.status in (?)", bun.In(statuses))

	return query
}
