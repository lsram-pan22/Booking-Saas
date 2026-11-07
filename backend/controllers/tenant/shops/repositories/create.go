package repositories

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountUnavailable = errors.New("account unavailable")
var ErrSubscriptionRequired = errors.New("active subscription required")
var ErrShopLimit = errors.New("shop quota reached")
var ErrOwnerAccountRequired = errors.New("staff accounts cannot create shops")

type ShopRepository struct{ DB *pgxpool.Pool }

func NewShopRepository(db *pgxpool.Pool) *ShopRepository { return &ShopRepository{DB: db} }

func (r *ShopRepository) Create(ctx context.Context, userID string, input models.NewShop) (models.Shop, error) {
	var shop models.Shop
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return shop, err
	}
	defer tx.Rollback(context.Background())
	// Serialize shop creation for an account, including simultaneous requests.
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM users WHERE id=$1 AND status='active' AND system_role='user' FOR UPDATE`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return shop, ErrAccountUnavailable
	}
	if err != nil {
		return shop, err
	}
	var shopLimit int
	var staffOnly bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members WHERE user_id=$1 AND role='staff') AND NOT EXISTS(SELECT 1 FROM tenant_members WHERE user_id=$1 AND role='owner')`, userID).Scan(&staffOnly)
	if err != nil {
		return shop, err
	}
	if staffOnly {
		return shop, ErrOwnerAccountRequired
	}
	err = tx.QueryRow(ctx, `SELECT shop_limit FROM subscriptions WHERE user_id=$1 AND status='active' AND starts_at<=NOW() AND (ends_at IS NULL OR ends_at>NOW()) AND user_limit>=1 FOR SHARE`, userID).Scan(&shopLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return shop, ErrSubscriptionRequired
	}
	if err != nil {
		return shop, err
	}
	var ownedShopCount int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM tenant_members WHERE user_id=$1 AND role='owner'`, userID).Scan(&ownedShopCount)
	if err != nil {
		return shop, err
	}
	if ownedShopCount >= shopLimit {
		return shop, ErrShopLimit
	}
	err = tx.QueryRow(ctx, `INSERT INTO tenants(name,slug,business_type,timezone,status) VALUES($1,$2,$3,'Asia/Bangkok','active') RETURNING id::text,name,business_type,slug,status`, input.Name, input.Slug, input.BusinessType).Scan(&shop.ID, &shop.Name, &shop.BusinessType, &shop.Slug, &shop.Status)
	if err != nil {
		return shop, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'owner','active')`, shop.ID, userID)
	if err != nil {
		return shop, err
	}
	if err = tx.Commit(ctx); err != nil {
		return models.Shop{}, err
	}
	shop.Role = "owner"
	return shop, nil
}
