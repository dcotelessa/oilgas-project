package repository

import (
	"database/sql"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *sql.DB
}

func NewBaseRepository(db *sql.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) GetDB() *sql.DB {
	return r.db
}
