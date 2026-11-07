package repositories

import (
	"booking-saas/controllers/tenant/services/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("service or shop not found or inaccessible")

type ServiceRepository struct{ DB *pgxpool.Pool }

func NewServiceRepository(db *pgxpool.Pool) *ServiceRepository { return &ServiceRepository{DB: db} }

const columns = `s.id::text,s.tenant_id::text,s.name,s.description,s.duration_minutes,s.price_satang,s.status`

func scan(row pgx.Row) (models.Service, error) {
	var service models.Service
	err := row.Scan(&service.ID, &service.TenantID, &service.Name, &service.Description, &service.DurationMinutes, &service.PriceSatang, &service.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return service, ErrNotFound
	}
	return service, err
}
func (r *ServiceRepository) List(ctx context.Context, userID, shopID string) (models.ServiceList, error) {
	result := models.ServiceList{Items: make([]models.Service, 0)}
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.Background())
	err = tx.QueryRow(ctx, `SELECT t.name,t.status,m.role FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id WHERE t.id=$1 AND m.user_id=$2 AND u.system_role='user' AND u.status='active' AND m.status='active' AND (m.role='owner' AND t.status IN ('pending','active','inactive') OR m.role='staff' AND t.status='active')`, shopID, userID).Scan(&result.ShopName, &result.ShopStatus, &result.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrNotFound
	}
	if err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT `+columns+` FROM services s WHERE s.tenant_id=$1 AND ($2='owner' OR s.status='active') ORDER BY s.created_at,s.id`, shopID, result.Role)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		service, err := scan(rows)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, service)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, err
	}
	return result, nil
}
func (r *ServiceRepository) Save(ctx context.Context, userID, shopID, serviceID string, req models.ServiceRequest) (models.Service, error) {
	args := []interface{}{shopID, userID, req.Name, req.Description, req.DurationMinutes, req.PriceSatang, req.Status}
	// Authorization is part of each write, not just a check in the UI.
	query := `INSERT INTO services AS s (tenant_id,name,description,duration_minutes,price_satang,status)
 SELECT t.id,$3,$4,$5,$6,$7 FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id
 WHERE t.id=$1 AND m.user_id=$2 AND m.role='owner' AND m.status='active' AND u.status='active' AND u.system_role='user' AND t.status IN ('active','inactive') RETURNING ` + columns
	if serviceID != "" {
		args = append(args, serviceID)
		query = `UPDATE services s SET name=$3,description=$4,duration_minutes=$5,price_satang=$6,status=$7,updated_at=NOW()
  FROM tenants t,tenant_members m,users u
  WHERE s.id=$8 AND s.tenant_id=$1 AND t.id=s.tenant_id AND m.tenant_id=t.id AND m.user_id=$2 AND u.id=m.user_id
  AND m.role='owner' AND m.status='active' AND u.status='active' AND u.system_role='user' AND t.status IN ('active','inactive') RETURNING ` + columns
	}
	return scan(r.DB.QueryRow(ctx, query, args...))
}
