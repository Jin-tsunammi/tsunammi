package repository

import (
	"context"
	"database/sql"
	"errors"
	"mm/internal/model"
	"mm/pkg/repository"

	"github.com/uptrace/bun"
)

type WalletRepository struct {
	repository.Generic[model.Wallet, uint64]
}

func NewWalletRepository(genericRepository repository.Generic[model.Wallet, uint64]) *WalletRepository {
	return &WalletRepository{Generic: genericRepository}
}

func (r *WalletRepository) WithTx(tx bun.Tx) *WalletRepository {
	return &WalletRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *WalletRepository) FindAllWalletsByAddr(ctx context.Context, ids []string) ([]model.Wallet, error) {
	wallets := make([]model.Wallet, 0, len(ids))
	err := r.DB.NewSelect().
		Model(&wallets).
		Where("public_key IN (?)", bun.In(ids)).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func (r *WalletRepository) SaveSolanaWallets(ctx context.Context, wallets []model.Wallet) ([]model.Wallet, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if len(wallets) == 0 {
		return wallets, nil
	}

	_, err = tx.NewInsert().
		Model(&wallets).
		//Column("status, user_id", "public_key").
		Returning("id, created_at").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	projectWallets := make([]model.ProjectWallet, 0)
	for i := range wallets {
		for _, projectID := range wallets[i].ProjectIDs {
			projectWallets = append(projectWallets, model.ProjectWallet{
				ProjectID: projectID,
				WalletID:  wallets[i].ID,
			})
		}
	}

	if len(projectWallets) > 0 {
		_, err = tx.NewInsert().Model(&projectWallets).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return wallets, nil
}

func (r *WalletRepository) GetWallets(ctx context.Context, walletIds []uint64) ([]model.Wallet, error) {
	var wallets = make([]model.Wallet, len(walletIds))

	q := r.DB.NewSelect().
		Model(&wallets)

	if len(walletIds) > 0 {
		q = q.Where("id IN (?)", bun.In(walletIds))
	}

	err := q.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return wallets, nil
}

func (r *WalletRepository) DeleteWalletsByIds(ctx context.Context, walletsID []uint64) error {
	_, err := r.DB.NewDelete().
		Model((*model.Wallet)(nil)).
		Where("id IN (?)", bun.In(walletsID)).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *WalletRepository) Delete(ctx context.Context, id, userID uint64) (string, error) {
	w := &model.Wallet{}
	res, err := r.DB.NewDelete().
		Model(w).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Returning("public_key").
		Exec(ctx)
	if err != nil {
		return "", err
	}

	if r, _ := res.RowsAffected(); r == 0 {
		return "", sql.ErrNoRows
	}

	return w.PublicKey, nil
}

func (r *WalletRepository) SaveProjectWallets(ctx context.Context, wallets []model.Wallet) error {
	projectWallets := make([]model.ProjectWallet, 0, len(wallets))
	for _, wallet := range wallets {
		for _, projectID := range wallet.ProjectIDs {
			projectWallets = append(projectWallets, model.ProjectWallet{
				ProjectID: projectID,
				WalletID:  wallet.ID,
			})
		}
	}

	if len(projectWallets) > 0 {
		_, err := r.DB.NewInsert().Model(&projectWallets).Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *WalletRepository) FetchWalletsByPublicKeys(ctx context.Context, keys []string, userID uint64) ([]model.Wallet, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	wallets := make([]model.Wallet, 0, len(keys))

	err := r.DB.NewSelect().
		Model(&wallets).
		ColumnExpr("w.*").
		ColumnExpr("COALESCE(array_agg(wp.project_id) FILTER (WHERE wp.project_id IS NOT NULL), '{}') AS project_ids").
		Join("LEFT JOIN project_wallets AS wp ON wp.wallet_id = w.id").
		Where("w.public_key IN (?)", bun.In(keys)).
		Where("w.user_id = ?", userID).
		Group("w.id").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func (r *WalletRepository) FetchByIDAndUserID(ctx context.Context, walletID, userID uint64) (*model.Wallet, error) {
	wallet := new(model.Wallet)
	err := r.DB.NewSelect().
		Model(wallet).
		Where("id = ? AND user_id = ?", walletID, userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

//3849

func (r *WalletRepository) FetchWalletsByIdsAndUserID(ctx context.Context, walletIDs []uint64, userID uint64) ([]model.Wallet, error) {
	wallets := make([]model.Wallet, 0, len(walletIDs))
	err := r.DB.NewSelect().
		Model(&wallets).
		Where("id IN (?) AND user_id = ?", bun.In(walletIDs), userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *WalletRepository) UpdateAllStatus(ctx context.Context, wallets []model.Wallet) error {
	_, err := r.DB.NewUpdate().
		Model(&wallets).
		Column("status").
		Bulk().
		Exec(ctx)

	return err
}

func (r *WalletRepository) FindAllByStatus(ctx context.Context, status model.WalletStatus) ([]model.Wallet, error) {
	wallets := make([]model.Wallet, 0)

	err := r.DB.NewSelect().
		Model(&wallets).
		Where("status = ?", status).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return wallets, nil
}
