// Package port interface of BookRepository
package port

import "github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain/model"

type BookRepository interface {
	Create(book *model.Book) error
	GetByID(id string) (*model.Book, error)
	List() ([]model.Book, error)
	Update(book *model.Book) error
	Delete(id string) error
}
