package models

import "time"

func (m *Models) InsertTenant(name string) (int64, error) {
	query := "INSERT INTO tenants (name, is_active, created_at) VALUES (?, 1, ?)"
	res, err := m.DB.Exec(query, name, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
