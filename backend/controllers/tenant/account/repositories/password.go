package repositories

import (
	"booking-saas/controllers/tenant/account/models"
	"context"

	"booking-saas/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordRepository struct{ DB *pgxpool.Pool }

func NewPasswordRepository(db *pgxpool.Pool) *PasswordRepository { return &PasswordRepository{DB: db} }

func (r *PasswordRepository) Reset(ctx context.Context, actor models.PasswordActor, shopID, target, hash, currentPassword string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var actorHash string
	if actor.Role == "user" {
		// Match booking/member writes: tenant first, then account rows. Also
		// authorize the owner before locking any target account.
		err = tx.QueryRow(ctx, `SELECT u.password_hash FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id
 WHERE t.id=$1 AND m.user_id=$2 AND m.role='owner' AND m.status='active' AND t.status IN ('active','inactive','pending') AND u.system_role='user' AND u.status='active' AND u.session_version=$3 AND NOT u.must_change_password FOR UPDATE OF t FOR SHARE OF m,u`, shopID, actor.ID, actor.Version).Scan(&actorHash)
	} else if actor.Role == "super_admin" {
		err = tx.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1 AND system_role='super_admin' AND status='active' AND session_version=$2 AND NOT must_change_password FOR SHARE`, actor.ID, actor.Version).Scan(&actorHash)
	} else {
		return models.ErrPasswordDenied
	}
	if err != nil {
		return models.ErrPasswordDenied
	}
	if !utils.CheckPassword(currentPassword, actorHash) {
		return models.ErrCurrentPassword
	}
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM users WHERE id=$1 AND system_role='user' AND status='active' FOR UPDATE`, target).Scan(&id)
	if err != nil {
		return models.ErrPasswordDenied
	}
	var allowed bool
	if actor.Role == "super_admin" {
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members WHERE user_id=$1 AND role='owner') OR EXISTS(SELECT 1 FROM subscriptions WHERE user_id=$1)`, target).Scan(&allowed)
	} else {
		// A password belongs to the account globally, so owners cannot reset an
		// owner account or an account with memberships in other shops.
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members a JOIN tenants t ON t.id=a.tenant_id JOIN tenant_members b ON b.tenant_id=t.id WHERE t.id=$1 AND a.user_id=$2 AND a.role='owner' AND a.status='active' AND t.status IN ('active','inactive','pending') AND b.user_id=$3 AND b.role='staff' AND b.status IN ('active','inactive') FOR SHARE OF a,t,b) AND NOT EXISTS(SELECT 1 FROM tenant_members WHERE user_id=$3 AND (role='owner' OR tenant_id<>$1::uuid)) AND NOT EXISTS(SELECT 1 FROM subscriptions WHERE user_id=$3)`, shopID, actor.ID, target).Scan(&allowed)
	}
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrPasswordDenied
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET password_hash=$2,must_change_password=TRUE,password_reset_expires_at=statement_timestamp()+INTERVAL '24 hours',session_version=session_version+1 WHERE id=$1`, target, hash); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO password_events(user_id,changed_by,event) VALUES($1,$2,'reset')`, target, actor.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PasswordRepository) Change(ctx context.Context, actor models.PasswordActor, current, next, hash string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var previous string
	err = tx.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1 AND status='active' AND system_role=$2 AND session_version=$3 AND (NOT must_change_password OR password_reset_expires_at>statement_timestamp()) FOR UPDATE`, actor.ID, actor.Role, actor.Version).Scan(&previous)
	if err != nil {
		return models.ErrPasswordDenied
	}
	if !utils.CheckPassword(current, previous) || utils.CheckPassword(next, previous) {
		return models.ErrCurrentPassword
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET password_hash=$2,must_change_password=FALSE,password_reset_expires_at=NULL,session_version=session_version+1 WHERE id=$1`, actor.ID, hash); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO password_events(user_id,changed_by,event) VALUES($1,$1,'change')`, actor.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PasswordRepository) Owners(ctx context.Context, search string) ([]models.ResettableOwner, error) {
	rows, err := r.DB.Query(ctx, `SELECT u.id::text,u.full_name,u.email,u.username FROM users u WHERE u.system_role='user' AND u.status='active' AND (EXISTS(SELECT 1 FROM tenant_members m WHERE m.user_id=u.id AND m.role='owner') OR EXISTS(SELECT 1 FROM subscriptions s WHERE s.user_id=u.id)) AND ($1='' OR u.email ILIKE '%'||$1||'%' OR u.username ILIKE '%'||$1||'%' OR u.full_name ILIKE '%'||$1||'%') ORDER BY u.full_name,u.id LIMIT 100`, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.ResettableOwner, 0)
	for rows.Next() {
		var item models.ResettableOwner
		if err = rows.Scan(&item.ID, &item.FullName, &item.Email, &item.Username); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
