package repositories

import (
	"booking-saas/controllers/tenant/account/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsRepository struct{ DB *pgxpool.Pool }

func NewSettingsRepository(db *pgxpool.Pool) *SettingsRepository { return &SettingsRepository{DB: db} }
func (r *SettingsRepository) Update(ctx context.Context, id string, req models.UpdateProfileRequest) (models.TenantUser, error) {
	var user models.TenantUser
	err := r.DB.QueryRow(ctx, `UPDATE users SET full_name=$1,phone=$2,updated_at=NOW() WHERE id=$3 AND system_role='user' AND status='active' RETURNING id::text,email,full_name,phone`, req.FullName, req.Phone, id).Scan(&user.ID, &user.Email, &user.FullName, &user.Phone)
	return user, err
}
