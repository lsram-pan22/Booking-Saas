package repositories

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrMemberQuota = errors.New("member quota reached")
var ErrMemberExists = errors.New("member already exists")
var ErrMemberAccountUnavailable = errors.New("member account unavailable")
var ErrMemberEmailExists = errors.New("member email already exists")
var ErrMemberUsernameExists = errors.New("member username already exists")
var ErrStaffShopAssigned = errors.New("staff already belongs to a shop")

func (r *ShopRepository) AddMember(ctx context.Context, ownerID, shopID string, input models.NewMember) (models.ShopMember, error) {
	var member models.ShopMember
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return member, err
	}
	defer tx.Rollback(context.Background())
	// The shop lock serializes all seat allocations for this shop.
	var shop string
	err = tx.QueryRow(ctx, `SELECT t.id::text FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id WHERE t.id=$1 AND m.user_id=$2 AND m.role='owner' AND m.status='active' AND u.status='active' AND u.system_role='user' AND t.status='active' FOR UPDATE OF t FOR SHARE OF m,u`, shopID, ownerID).Scan(&shop)
	if errors.Is(err, pgx.ErrNoRows) {
		return member, ErrShopNotFound
	}
	if err != nil {
		return member, err
	}
	var limit, count int
	err = tx.QueryRow(ctx, `SELECT user_limit FROM subscriptions WHERE user_id=$1 AND status='active' AND starts_at<=NOW() AND (ends_at IS NULL OR ends_at>NOW()) FOR SHARE`, ownerID).Scan(&limit)
	if errors.Is(err, pgx.ErrNoRows) {
		return member, ErrSubscriptionRequired
	}
	if err != nil {
		return member, err
	}
	err = tx.QueryRow(ctx, `SELECT count(*) FROM tenant_members WHERE tenant_id=$1 AND status='active'`, shopID).Scan(&count)
	if err != nil {
		return member, err
	}
	if count >= limit {
		return member, ErrMemberQuota
	}
	if input.Mode == "new" {
		err = tx.QueryRow(ctx, `INSERT INTO users(username,email,password_hash,full_name,phone,system_role,status) VALUES($1,$2,$3,$4,$5,'user','active') RETURNING id::text,full_name,email,phone,status,username`, input.Username, input.Email, input.PasswordHash, input.FullName, input.Phone).Scan(&member.UserID, &member.FullName, &member.Email, &member.Phone, &member.UserStatus, &member.Username)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_lower_unique" {
				return member, ErrMemberUsernameExists
			}
			return member, ErrMemberEmailExists
		}
	} else {
		err = tx.QueryRow(ctx, `SELECT id::text,full_name,email,COALESCE(phone,''),status,username FROM users WHERE LOWER(email)=$1 AND system_role='user' AND status='active' FOR SHARE`, input.Email).Scan(&member.UserID, &member.FullName, &member.Email, &member.Phone, &member.UserStatus, &member.Username)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return member, ErrMemberAccountUnavailable
	}
	if err != nil {
		return member, err
	}
	// Never change existing account details/password, or accept an owner role from the request.
	err = tx.QueryRow(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'staff','active') ON CONFLICT (tenant_id,user_id) DO NOTHING RETURNING id::text,role,status`, shopID, member.UserID).Scan(&member.ID, &member.Role, &member.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return member, ErrMemberExists
	}
	var membershipErr *pgconn.PgError
	if errors.As(err, &membershipErr) && membershipErr.Code == "23505" && membershipErr.ConstraintName == "tenant_members_one_staff_shop" {
		return member, ErrStaffShopAssigned
	}
	if err != nil {
		return member, err
	}
	if err = tx.Commit(ctx); err != nil {
		return models.ShopMember{}, err
	}
	return member, nil
}
