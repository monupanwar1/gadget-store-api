package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewSQLite(path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}