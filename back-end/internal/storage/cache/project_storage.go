package cache

import (
	"context"
	"fmt"
	"mm/internal/model"

	imcache "github.com/patrickmn/go-cache"
)

var _ ProjectStorage = (*projectStorage)(nil)

type ProjectStorage interface {
	Get(ctx context.Context, id, userID uint64) (*model.ProjectWithWalletsResponse, error)
	Set(ctx context.Context, id, userID uint64, project *model.ProjectWithWalletsResponse)
}

type projectStorage struct {
	cache *imcache.Cache
}

func NewProjectStorage() ProjectStorage {
	cache := imcache.New(imcache.DefaultExpiration, imcache.NoExpiration)
	return &projectStorage{cache: cache}
}

func (s *projectStorage) Get(ctx context.Context, id, userID uint64) (*model.ProjectWithWalletsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		res, ok := s.cache.Get(fmt.Sprintf("project_%d_%d", userID, id))
		if !ok {
			return nil, fmt.Errorf("project not found")
		}

		return res.(*model.ProjectWithWalletsResponse), nil
	}
}

func (s *projectStorage) Set(ctx context.Context, id, userID uint64, project *model.ProjectWithWalletsResponse) {
	s.cache.Set(fmt.Sprintf("project_%d_%d", userID, id), project, imcache.NoExpiration)
}
