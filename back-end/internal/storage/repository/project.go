package repository

import (
	"context"
	"database/sql"
	"fmt"
	"mm/internal/model"
	"mm/pkg/repository"
	"time"

	"github.com/uptrace/bun"
)

type ProjectRepository struct {
	repository.Generic[model.Project, uint64]
}

func NewProjectRepository(genericRepository repository.Generic[model.Project, uint64]) *ProjectRepository {
	return &ProjectRepository{Generic: genericRepository}
}

func (r *ProjectRepository) FetchProjectWithWalletsByID(ctx context.Context, id, userID uint64) (*model.ProjectWithWallets, error) {
	project := new(model.ProjectWithWallets)
	err := r.DB.NewSelect().
		Model(project).
		Where("id = ? AND user_id = ?", id, userID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	wallets := make([]model.Wallet, 0)
	err = r.DB.NewSelect().
		Table("wallets").
		Column("wallets.*").
		Join("JOIN project_wallets pw ON pw.wallet_id = wallets.id").
		Where("pw.project_id = ?", id).
		Where("wallets.status = ?", model.WalletStatusSuccess).
		Order("wallets.id ASC").
		Scan(ctx, &wallets)

	if err != nil {
		return nil, fmt.Errorf("get wallets for project %d: %w", id, err)
	}
	project.Wallets = wallets

	return project, nil
}

func (r *ProjectRepository) DeleteProjectByIDAndUserID(ctx context.Context, id, userID uint64) error {
	res, err := r.DB.NewDelete().
		Model((*model.Project)(nil)).
		Where("id = ? AND user_id = ?", id, userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ProjectRepository) FindAllByUserID(ctx context.Context, userID uint64, page, pageSize int, sortDesc bool) ([]model.ProjectWithWallets, int, error) {
	var rows []struct {
		ProjectID   uint64    `bun:"project_id"`
		ProjectName string    `bun:"project_name"`
		ProjectUser uint64    `bun:"user_id"`
		WalletID    uint64    `bun:"wallet_id"`
		PublicKey   string    `bun:"public_key"`
		CreatedAt   time.Time `bun:"created_at"`
		CreatedAtP  time.Time `bun:"project_created_at"`
	}

	orderExpr := "p.id"

	if sortDesc {
		orderExpr += " DESC"
	} else {
		orderExpr += " ASC"
	}

	pidsQuery := r.DB.NewSelect().
		Model((*model.Project)(nil)).
		Column("id").
		Where("user_id = ?", userID).
		OrderExpr(orderExpr)

	totalCount, err := pidsQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if page > 0 && pageSize > 0 {
		pidsQuery.Limit(pageSize).Offset(pageSize * (page - 1))
	}

	query := r.DB.NewSelect().
		ColumnExpr("p.id AS project_id, p.name AS project_name, p.user_id, p.created_at AS project_created_at").
		ColumnExpr("w.id AS wallet_id, w.public_key, w.created_at").
		Model((*model.ProjectWithWallets)(nil)).
		Join("LEFT JOIN project_wallets pw ON pw.project_id = p.id").
		Join("LEFT JOIN wallets w ON w.id = pw.wallet_id").
		Where("p.id IN (?)", pidsQuery).
		Where("w.status = ? OR w.status is NULL", model.WalletStatusSuccess).
		OrderExpr(orderExpr).
		Order("w.id ASC")

	err = query.Scan(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	projects := make([]model.ProjectWithWallets, 0)

	for _, row := range rows {
		if len(projects) == 0 || projects[len(projects)-1].ID != row.ProjectID {
			projects = append(projects, model.ProjectWithWallets{
				Project: model.Project{
					ID:        row.ProjectID,
					Name:      row.ProjectName,
					UserID:    row.ProjectUser,
					CreatedAt: row.CreatedAtP,
				},
				Wallets: []model.Wallet{},
			})
		}

		if row.WalletID != 0 {
			currentProject := &projects[len(projects)-1]

			currentProject.Wallets = append(currentProject.Wallets, model.Wallet{
				ID:        row.WalletID,
				PublicKey: row.PublicKey,
				CreatedAt: row.CreatedAt,
			})
		}
	}

	for i := range projects {
		projects[i].WalletCount = uint64(len(projects[i].Wallets))
	}

	return projects, totalCount, nil
}

func (r *ProjectRepository) FindAllByUserID1(ctx context.Context, userID uint64, page, pageSize int) ([]model.ProjectWithWallets, error) {
	var rows []struct {
		ProjectID   uint64    `bun:"project_id"`
		ProjectName string    `bun:"project_name"`
		ProjectUser uint64    `bun:"user_id"`
		WalletID    uint64    `bun:"wallet_id"`
		PublicKey   string    `bun:"public_key"`
		CreatedAt   time.Time `bun:"created_at"`
	}

	query := r.DB.NewSelect().
		ColumnExpr("p.id AS project_id, p.name AS project_name, p.user_id").
		ColumnExpr("w.id AS wallet_id, w.public_key, w.created_at").
		Model((*model.ProjectWithWallets)(nil)).
		Join("LEFT JOIN project_wallets pw ON pw.project_id = p.id").
		Join("LEFT JOIN wallets w ON w.id = pw.wallet_id").
		Where("p.user_id = ?", userID).
		Where("w.status = ? OR w.status is NULL", model.WalletStatusSuccess).
		Order("p.id")

	if page != 0 && pageSize != 0 {
		query.Limit(pageSize).Offset(pageSize * (page - 1))
	}

	err := query.
		Scan(ctx, &rows)

	if err != nil {
		return nil, err
	}

	projectsMap := make(map[uint64]*model.ProjectWithWallets)
	for _, row := range rows {
		proj, exists := projectsMap[row.ProjectID]
		if !exists {
			proj = &model.ProjectWithWallets{
				Project: model.Project{
					ID:     row.ProjectID,
					Name:   row.ProjectName,
					UserID: row.ProjectUser,
				},
			}
			projectsMap[row.ProjectID] = proj
		}
		if row.WalletID != 0 {
			proj.Wallets = append(proj.Wallets, model.Wallet{
				ID:        row.WalletID,
				PublicKey: row.PublicKey,
				CreatedAt: row.CreatedAt,
			})
		}
	}

	projects := make([]model.ProjectWithWallets, 0, len(projectsMap))
	for _, p := range projectsMap {
		p.WalletCount = uint64(len(p.Wallets))
		projects = append(projects, *p)
	}

	return projects, nil
}

func (r *ProjectRepository) FindAllOlderThanWithWalletsByStatus(ctx context.Context, t time.Duration, status model.WalletStatus) ([]model.ProjectWithWallets, error) {
	var rows []struct {
		ProjectID   uint64    `bun:"project_id"`
		ProjectName string    `bun:"project_name"`
		ProjectUser uint64    `bun:"user_id"`
		WalletID    uint64    `bun:"wallet_id"`
		PublicKey   string    `bun:"public_key"`
		CreatedAt   time.Time `bun:"created_at"`
	}

	threshold := time.Now().UTC().Add(-t)

	err := r.DB.NewSelect().
		ColumnExpr("p.id AS project_id, p.name AS project_name, p.user_id").
		ColumnExpr("w.id AS wallet_id, w.public_key, w.created_at").
		Model((*model.ProjectWithWallets)(nil)).
		Join("LEFT JOIN project_wallets pw ON pw.project_id = p.id").
		Join("LEFT JOIN wallets w ON w.id = pw.wallet_id").
		Where("w.created_at < ?", threshold).
		Where("w.status = ?", status).
		Scan(ctx, &rows)

	if err != nil {
		return nil, err
	}

	projectsMap := make(map[uint64]*model.ProjectWithWallets)
	for _, row := range rows {
		proj, exists := projectsMap[row.ProjectID]
		if !exists {
			proj = &model.ProjectWithWallets{
				Project: model.Project{
					ID:     row.ProjectID,
					Name:   row.ProjectName,
					UserID: row.ProjectUser,
				},
			}
			projectsMap[row.ProjectID] = proj
		}
		if row.WalletID != 0 {
			proj.Wallets = append(proj.Wallets, model.Wallet{
				ID:        row.WalletID,
				PublicKey: row.PublicKey,
				CreatedAt: row.CreatedAt,
			})
		}
	}

	projects := make([]model.ProjectWithWallets, 0, len(projectsMap))
	for _, p := range projectsMap {
		p.WalletCount = uint64(len(p.Wallets))
		projects = append(projects, *p)
	}

	return projects, nil
}

func (r *ProjectRepository) FetchAllByIDs(ctx context.Context, ids []uint64, userID uint64) ([]model.Project, error) {
	projects := make([]model.Project, 0, len(ids))

	err := r.DB.NewSelect().
		Model(&projects).
		Where("id IN (?) AND user_id = ?", bun.In(ids), userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepository) FetchAllByUserID(ctx context.Context, userID uint64) ([]model.Project, error) {
	projects := make([]model.Project, 0)

	err := r.DB.NewSelect().
		Model(&projects).
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepository) UpdateByID(ctx context.Context, id, userID uint64, req model.EditProjectReq) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Project)(nil)).
		Where("id = ? AND user_id = ?", id, userID).
		Set("name = ?", req.Name).
		Exec(ctx)
	return err
}
