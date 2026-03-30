// Package service use for business logic
package service

import (
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain/model"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

type bookService struct {
	repo port.BookRepository
}

func NewBookService(repo port.BookRepository) port.BookService {
	return &bookService{repo: repo}
}

func (s *bookService) Create(title, author string) (*model.Book, error) {
	return nil, nil
}

func (s *bookService) GetByID(id string) (*model.Book, error) {
	return nil, nil
}

func (s *bookService) List() ([]model.Book, error) {
	return nil, nil
}

func (s *bookService) Update(id, title, author string) (*model.Book, error) {
	return nil, nil
}

func (s *bookService) Delete(id string) error {
	return nil
}
