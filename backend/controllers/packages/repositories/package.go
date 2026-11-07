package repositories

import (
	"booking-saas/controllers/packages/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("package not found")
var ErrDuplicateCode = errors.New("package code already exists")

type PackageRepository struct{ DB *pgxpool.Pool }

func NewPackageRepository(db *pgxpool.Pool) *PackageRepository { return &PackageRepository{DB: db} }

const columns = `id::text, code, name, description, monthly_price_satang, branch_limit, user_limit, staff_management_level, report_level, export_excel, multi_branch, api_webhook, (status = 'active'), sort_order`

func scan(row pgx.Row) (models.Package, error) {
	var p models.Package
	err := row.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.MonthlyPriceSatang, &p.BranchLimit, &p.UserLimit, &p.StaffManagement, &p.Reports, &p.ExportExcel, &p.MultiBranch, &p.APIWebhook, &p.Active, &p.SortOrder)
	return p, err
}

func (r *PackageRepository) List(ctx context.Context, activeOnly bool) ([]models.Package, error) {
	query := "SELECT " + columns + " FROM packages"
	if activeOnly {
		query += " WHERE status = 'active'"
	}
	rows, err := r.DB.Query(ctx, query+" ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.Package, 0)
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *PackageRepository) Save(ctx context.Context, p models.Package, id string) (models.Package, error) {
	status := "inactive"
	if p.Active {
		status = "active"
	}
	args := []interface{}{p.Code, p.Name, p.Description, p.MonthlyPriceSatang, p.BranchLimit, p.UserLimit, p.StaffManagement, p.Reports, p.ExportExcel, p.MultiBranch, p.APIWebhook, status, p.SortOrder}
	query := `INSERT INTO packages (code,name,description,monthly_price_satang,branch_limit,user_limit,staff_management_level,report_level,export_excel,multi_branch,api_webhook,status,sort_order) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING ` + columns
	if id != "" {
		args = append(args, id)
		query = `UPDATE packages SET code=$1,name=$2,description=$3,monthly_price_satang=$4,branch_limit=$5,user_limit=$6,staff_management_level=$7,report_level=$8,export_excel=$9,multi_branch=$10,api_webhook=$11,status=$12,sort_order=$13,updated_at=now() WHERE id=$14 RETURNING ` + columns
	}
	saved, err := scan(r.DB.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return saved, ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return saved, ErrDuplicateCode
	}
	return saved, err
}
