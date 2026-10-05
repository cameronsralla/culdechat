package users

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/cameronsralla/culdechat/server/internal/db/dbq"
)

// ListParams is the admin roster query. Search and column filters apply to the
// full table; page/page_size only slice the already-filtered result.
type ListParams struct {
	Q         string
	Name      string
	Email     string
	Unit      string
	Status    string
	Admin     *bool
	Directory *bool
	Sort      string
	Dir       string
	Page      int
	PageSize  int
}

// UserPage is one page of the roster plus the unpaged match count.
type UserPage struct {
	Items    []User `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

var userSortColumns = map[string]string{
	"display_name":     "display_name",
	"email":            "email",
	"unit_number":      "unit_number",
	"status":           "status",
	"is_admin":         "is_admin",
	"directory_opt_in": "directory_opt_in",
	"created_at":       "created_at",
}

// DirectoryParams is the People directory query. The listing is always
// active residents who opted in; search and column filters apply before paging.
type DirectoryParams struct {
	Q        string
	Name     string
	Email    string
	Unit     string
	Sort     string
	Dir      string
	Page     int
	PageSize int
}

// DirectoryResult is one page of the directory plus the unpaged match count.
type DirectoryResult struct {
	Items    []DirectoryEntry `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

var directorySortColumns = map[string]string{
	"display_name": "display_name",
	"email":        "email",
	"unit_number":  "unit_number",
}

// ListPage returns a filtered, sorted, paginated roster.
func (s *Service) ListPage(ctx context.Context, p ListParams) (UserPage, error) {
	p.Page, p.PageSize = clampPage(p.Page, p.PageSize)
	col, ok := userSortColumns[p.Sort]
	if !ok {
		col = "unit_number"
	}
	dirSQL, _ := sortSQL(p.Dir)

	where, args := userWhere(p)
	clause := ""
	if len(where) > 0 {
		clause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	countSQL := "SELECT count(*) FROM residents " + clause
	if err := s.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return UserPage{}, err
	}

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)
	listSQL := fmt.Sprintf(`SELECT id, email, unit_number, unit_id, is_primary, display_name, password_hash, is_admin, status, directory_opt_in, created_at, updated_at, deactivated_at
FROM residents %s
ORDER BY %s %s, id ASC
LIMIT $%d OFFSET $%d`, clause, col, dirSQL, limitArg, offsetArg)

	rows, err := s.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return UserPage{}, err
	}
	defer rows.Close()

	items := []User{}
	for rows.Next() {
		var u dbq.Resident
		if err := rows.Scan(
			&u.ID, &u.Email, &u.UnitNumber, &u.UnitID, &u.IsPrimary, &u.DisplayName, &u.PasswordHash,
			&u.IsAdmin, &u.Status, &u.DirectoryOptIn, &u.CreatedAt, &u.UpdatedAt, &u.DeactivatedAt,
		); err != nil {
			return UserPage{}, err
		}
		items = append(items, ToUser(u))
	}
	if err := rows.Err(); err != nil {
		return UserPage{}, err
	}
	return UserPage{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// DirectoryPage returns a filtered, sorted, paginated People listing.
func (s *Service) DirectoryPage(ctx context.Context, p DirectoryParams, viewer uuid.UUID) (DirectoryResult, error) {
	p.Page, p.PageSize = clampPage(p.Page, p.PageSize)
	col, ok := directorySortColumns[p.Sort]
	if !ok {
		col = "unit_number"
	}
	dirSQL, _ := sortSQL(p.Dir)

	personWhere, unitWhere, args := directoryFilters(p)
	union := fmt.Sprintf(`
SELECT 'person'::text AS kind, id, unit_number, display_name, email
FROM residents
WHERE %s
UNION ALL
SELECT 'unit'::text, NULL::uuid, unit_number, ''::text, ''::text
FROM residents
WHERE %s
GROUP BY unit_number`, strings.Join(personWhere, " AND "), strings.Join(unitWhere, " AND "))

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM ("+union+") AS directory_rows", args...).Scan(&total); err != nil {
		return DirectoryResult{}, err
	}

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)
	listSQL := fmt.Sprintf(`SELECT kind, id, unit_number, display_name, email FROM (%s) AS directory_rows
ORDER BY %s %s, unit_number ASC, kind ASC
LIMIT $%d OFFSET $%d`, union, col, dirSQL, limitArg, offsetArg)

	rows, err := s.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return DirectoryResult{}, err
	}
	defer rows.Close()

	items := []DirectoryEntry{}
	for rows.Next() {
		var e DirectoryEntry
		if err := rows.Scan(&e.Kind, &e.ID, &e.UnitNumber, &e.DisplayName, &e.Email); err != nil {
			return DirectoryResult{}, err
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return DirectoryResult{}, err
	}
	if err := s.markOwnUnit(ctx, viewer, items); err != nil {
		return DirectoryResult{}, err
	}
	return DirectoryResult{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// markOwnUnit flags a hidden unit when the viewer is the resident a message
// to that unit would reach, so the client can hide the button.
func (s *Service) markOwnUnit(ctx context.Context, viewer uuid.UUID, items []DirectoryEntry) error {
	rows, err := s.pool.Query(ctx, `
SELECT unit_number, id
FROM residents
WHERE status = 'active' AND directory_opt_in = FALSE AND is_primary = TRUE`)
	if err != nil {
		return err
	}
	defer rows.Close()
	own := map[string]bool{}
	for rows.Next() {
		var unit string
		var id uuid.UUID
		if err := rows.Scan(&unit, &id); err != nil {
			return err
		}
		if id == viewer {
			own[unit] = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range items {
		if items[i].Kind == "unit" && own[items[i].UnitNumber] {
			items[i].Self = true
		}
	}
	return nil
}

// directoryFilters splits the People query. Name and email match listed
// residents only, so a search cannot reveal someone who opted out. Unit
// matches both listed people and occupied hidden units.
func directoryFilters(p DirectoryParams) (person, unit []string, args []any) {
	person = []string{"status = 'active'", "directory_opt_in = TRUE"}
	unit = []string{"status = 'active'", "directory_opt_in = FALSE", "is_primary = TRUE"}
	like := func(dest *[]string, column, raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		args = append(args, "%"+escapeLike(raw)+"%")
		*dest = append(*dest, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", column, len(args)))
	}
	if q := strings.TrimSpace(p.Q); q != "" {
		args = append(args, "%"+escapeLike(q)+"%")
		n := len(args)
		person = append(person, fmt.Sprintf("(email ILIKE $%d ESCAPE '\\' OR display_name ILIKE $%d ESCAPE '\\' OR unit_number ILIKE $%d ESCAPE '\\')", n, n, n))
		unit = append(unit, fmt.Sprintf("unit_number ILIKE $%d ESCAPE '\\'", n))
	}
	like(&person, "display_name", p.Name)
	like(&person, "email", p.Email)
	like(&person, "unit_number", p.Unit)
	like(&unit, "unit_number", p.Unit)
	if strings.TrimSpace(p.Name) != "" || strings.TrimSpace(p.Email) != "" {
		unit = append(unit, "FALSE")
	}
	return person, unit, args
}

func clampPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func sortSQL(dir string) (string, string) {
	if strings.EqualFold(dir, "desc") {
		return "DESC", "desc"
	}
	return "ASC", "asc"
}

func userWhere(p ListParams) ([]string, []any) {
	var where []string
	var args []any
	like := func(column, raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		args = append(args, "%"+escapeLike(raw)+"%")
		where = append(where, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", column, len(args)))
	}
	if q := strings.TrimSpace(p.Q); q != "" {
		args = append(args, "%"+escapeLike(q)+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(email ILIKE $%d ESCAPE '\\' OR display_name ILIKE $%d ESCAPE '\\' OR unit_number ILIKE $%d ESCAPE '\\' OR status ILIKE $%d ESCAPE '\\')", n, n, n, n))
	}
	like("display_name", p.Name)
	like("email", p.Email)
	like("unit_number", p.Unit)
	if p.Status == "invited" || p.Status == "active" || p.Status == "inactive" {
		args = append(args, p.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if p.Admin != nil {
		args = append(args, *p.Admin)
		where = append(where, fmt.Sprintf("is_admin = $%d", len(args)))
	}
	if p.Directory != nil {
		args = append(args, *p.Directory)
		where = append(where, fmt.Sprintf("directory_opt_in = $%d", len(args)))
	}
	return where, args
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
