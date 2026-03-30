// Package sqliter for repository of book
package sqliter

import (
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain/model"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
	"gorm.io/gorm"
)

var _ port.BookRepository = (*BookRepositorySqlite)(nil)

type BookRepositorySqlite struct {
	db *gorm.DB
}

func NewBookRepositorySqlite(db *gorm.DB) *BookRepositorySqlite {
	return &BookRepositorySqlite{db: db}
}

func (r *BookRepositorySqlite) Create(book *model.Book) error {
	return nil
}

func (r *BookRepositorySqlite) GetByID(id string) (*model.Book, error) {
	return nil, nil
}

func (r *BookRepositorySqlite) List() ([]model.Book, error) {
	var books []model.Book
	err := r.db.Find(&books).Error
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (r *BookRepositorySqlite) Update(book *model.Book) error {
	return r.db.Save(book).Error
}

func (r *BookRepositorySqlite) Delete(id string) error {
	return nil
}
