package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/cameronsralla/culdechat/server/internal/db/dbq"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

// Unit is an address in the community. Residents are assigned to it.
// One of them is the primary; more residents can be added later.
type Unit struct {
	ID        uuid.UUID      `json:"id"`
	Number    string         `json:"number"`
	CreatedAt time.Time      `json:"created_at"`
	Residents []UnitResident `json:"residents"`
}

type UnitResident struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	IsPrimary   bool      `json:"is_primary"`
}

type UnitPage struct {
	Items    []Unit `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type UnitParams struct {
	Q        string
	Dir      string
	Vacant   *bool
	Page     int
	PageSize int
}

type UnitInput struct {
	Number string `json:"number"`
}

func (in *UnitInput) Validate() error {
	var f httpx.Fields
	in.Number = strings.TrimSpace(in.Number)
	if in.Number == "" || len(in.Number) > 32 {
		f.Add("number", "required, max 32 characters")
	}
	return f.Err()
}

func (s *Service) ListUnits(ctx context.Context, p UnitParams) (UnitPage, error) {
	p.Page, p.PageSize = clampPage(p.Page, p.PageSize)
	dirSQL, _ := sortSQL(p.Dir)
	where, args := unitWhere(p)
	clause := ""
	if len(where) > 0 {
		clause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM units "+clause, args...).Scan(&total); err != nil {
		return UnitPage{}, err
	}

	limit := len(args) + 1
	offset := len(args) + 2
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)
	rows, err := s.pool.Query(ctx, fmt.Sprintf(`SELECT id, number, created_at FROM units %s ORDER BY number %s, id ASC LIMIT $%d OFFSET $%d`, clause, dirSQL, limit, offset), args...)
	if err != nil {
		return UnitPage{}, err
	}
	defer rows.Close()

	items := []Unit{}
	ids := []uuid.UUID{}
	index := map[uuid.UUID]int{}
	for rows.Next() {
		var u Unit
		if err := rows.Scan(&u.ID, &u.Number, &u.CreatedAt); err != nil {
			return UnitPage{}, err
		}
		u.Residents = []UnitResident{}
		index[u.ID] = len(items)
		ids = append(ids, u.ID)
		items = append(items, u)
	}
	if err := rows.Err(); err != nil {
		return UnitPage{}, err
	}
	if len(ids) == 0 {
		return UnitPage{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
	}

	people, err := s.pool.Query(ctx, `
SELECT id, email, display_name, status, is_primary, unit_id
FROM users
WHERE unit_id = ANY($1)
ORDER BY is_primary DESC, display_name ASC, id ASC`, ids)
	if err != nil {
		return UnitPage{}, err
	}
	defer people.Close()
	for people.Next() {
		var r UnitResident
		var unitID uuid.UUID
		if err := people.Scan(&r.ID, &r.Email, &r.DisplayName, &r.Status, &r.IsPrimary, &unitID); err != nil {
			return UnitPage{}, err
		}
		if i, ok := index[unitID]; ok {
			items[i].Residents = append(items[i].Residents, r)
		}
	}
	if err := people.Err(); err != nil {
		return UnitPage{}, err
	}
	return UnitPage{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

func (s *Service) CreateUnit(ctx context.Context, in UnitInput) (Unit, error) {
	row, err := s.q.CreateUnit(ctx, in.Number)
	if constraint(err) == "units_number_key" {
		return Unit{}, httpx.ErrConflict.WithMessage("that unit already exists")
	}
	if err != nil {
		return Unit{}, err
	}
	return Unit{ID: row.ID, Number: row.Number, CreatedAt: row.CreatedAt, Residents: []UnitResident{}}, nil
}

func (s *Service) RenameUnit(ctx context.Context, id uuid.UUID, in UnitInput) (Unit, error) {
	row, err := s.q.UpdateUnitNumber(ctx, dbq.UpdateUnitNumberParams{ID: id, Number: in.Number})
	if errors.Is(err, pgx.ErrNoRows) {
		return Unit{}, httpx.ErrNotFound
	}
	if constraint(err) == "units_number_key" {
		return Unit{}, httpx.ErrConflict.WithMessage("that unit already exists")
	}
	if err != nil {
		return Unit{}, err
	}
	return s.unitWithResidents(ctx, row)
}

func (s *Service) DeleteUnit(ctx context.Context, id uuid.UUID) error {
	if _, err := s.q.GetUnit(ctx, id); errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	} else if err != nil {
		return err
	}
	n, err := s.q.CountUnitResidents(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return httpx.ErrConflict.WithMessage("remove the residents before deleting this unit")
	}
	return s.q.DeleteUnit(ctx, id)
}

func (s *Service) unitWithResidents(ctx context.Context, row dbq.Unit) (Unit, error) {
	page, err := s.ListUnits(ctx, UnitParams{Q: row.Number, Page: 1, PageSize: 100})
	if err != nil {
		return Unit{}, err
	}
	for _, item := range page.Items {
		if item.ID == row.ID {
			return item, nil
		}
	}
	return Unit{ID: row.ID, Number: row.Number, CreatedAt: row.CreatedAt, Residents: []UnitResident{}}, nil
}

func unitWhere(p UnitParams) ([]string, []any) {
	var where []string
	var args []any
	if q := strings.TrimSpace(p.Q); q != "" {
		args = append(args, "%"+escapeLike(q)+"%")
		where = append(where, fmt.Sprintf("number ILIKE $%d ESCAPE '\\'", len(args)))
	}
	if p.Vacant != nil && *p.Vacant {
		where = append(where, "NOT EXISTS (SELECT 1 FROM users WHERE users.unit_id = units.id AND users.is_primary)")
	}
	return where, args
}
