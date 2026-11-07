package usecases

import (
	"booking-saas/controllers/packages/models"
	"context"
	"errors"
	"time"
)

var ErrInvalidPackage = errors.New("invalid package data")

type PackageRepository interface {
	List(context.Context, bool) ([]models.Package, error)
	Save(context.Context, models.Package, string) (models.Package, error)
}

type PackageUsecase struct{ Repo PackageRepository }

func NewPackageUsecase(repo PackageRepository) *PackageUsecase { return &PackageUsecase{Repo: repo} }

func (u *PackageUsecase) List(ctx context.Context, activeOnly bool) ([]models.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.List(ctx, activeOnly)
}

// An empty ID creates a package; a UUID updates an existing package.
func (u *PackageUsecase) Save(ctx context.Context, p models.Package, id string) (models.Package, error) {
	if (id != "" && !models.ValidID(id)) || !p.Valid() {
		return models.Package{}, ErrInvalidPackage
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Save(ctx, p, id)
}
