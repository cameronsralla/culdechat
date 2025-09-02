package models

import (
	"context"
	"errors"

	"github.com/cameronsralla/culdechat/connectors/postgres"
)

// DirectoryUser is a lightweight projection for the public directory.
type DirectoryUser struct {
	ID                string
	UnitNumber        string
	ProfilePictureURL *string
}

// ListDirectoryUsers returns active users who opted-in to the directory.
func ListDirectoryUsers(ctx context.Context) ([]DirectoryUser, error) {
	const q = `
SELECT id::text, unit_number, profile_picture_url
FROM users
WHERE is_directory_opt_in = TRUE AND status = 'active'
ORDER BY unit_number ASC;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DirectoryUser
	for rows.Next() {
		var du DirectoryUser
		var profile *string
		if err := rows.Scan(&du.ID, &du.UnitNumber, &profile); err != nil {
			return nil, err
		}
		du.ProfilePictureURL = profile
		out = append(out, du)
	}
	return out, rows.Err()
}
