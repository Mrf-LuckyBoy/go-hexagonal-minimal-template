// Package service implements the driving ports: the application's business
// logic. Services depend only on domain types and port interfaces, never on
// concrete adapters, so they can be unit-tested with in-memory fakes.
//
// Example:
//
//	type orderService struct {
//		repo  port.OrderRepository
//		cache port.Cache
//	}
//
//	func NewOrderService(repo port.OrderRepository, cache port.Cache) port.OrderService {
//		return &orderService{repo: repo, cache: cache}
//	}
package service
