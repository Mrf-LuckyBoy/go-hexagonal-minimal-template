package port

import "github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain/model"

type BookService interface {
	Create(title, author string) (*model.Book, error)
	GetByID(id string) (*model.Book, error)
	List() ([]model.Book, error)
	Update(id, title, author string) (*model.Book, error)
	Delete(id string) error
}
