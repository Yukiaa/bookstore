package main

import (
	"context"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Shelf{}, &Book{})
	return db, nil
}

type Shelf struct {
	ID        int64     `gorm:"primaryKey"`
	Theme     string    `gorm:"column:theme"`
	Size      int64     `gorm:"column:size"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Book struct {
	ID        int64     `gorm:"primaryKey"`
	Author    string    `gorm:"column:author"`
	Title     string    `gorm:"column:title"`
	ShelfID   int64     `gorm:"column:shelf_id"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateShelf(ctx context.Context, shelf *Shelf) (*Shelf, error) {
	if err := s.db.WithContext(ctx).Create(shelf).Error; err != nil {
		return nil, err
	}
	return shelf, nil
}

func (s *Store) ListShelves(ctx context.Context) ([]Shelf, error) {
	var shelves []Shelf
	if err := s.db.WithContext(ctx).Find(&shelves).Error; err != nil {
		return nil, err
	}
	return shelves, nil
}

func (s *Store) GetShelf(ctx context.Context, id int64) (*Shelf, error) {
	var shelf Shelf
	if err := s.db.WithContext(ctx).First(&shelf, id).Error; err != nil {
		return nil, err
	}
	return &shelf, nil
}

func (s *Store) DeleteShelf(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&Shelf{}, id).Error
}

func (s *Store) GetBookListByShelfID(ctx context.Context, shelfID int64, cursor string, pageSize int) ([]Book, error) {
	var books []Book
	if err := s.db.WithContext(ctx).Where("shelf_id = ?", shelfID).
		Where("id > ?", cursor).
		Order("id").
		Limit(pageSize).
		Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}
