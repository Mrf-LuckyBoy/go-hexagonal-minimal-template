// Package port defines the interfaces (ports) between the core and the
// outside world. The core owns these contracts; adapters fulfil them.
//
//   - Driving ports (inbound): use cases the core offers. Implemented by
//     services, called by driving adapters (HTTP, gRPC, CLI, consumers).
//   - Driven ports (outbound): what the core needs. Implemented by driven
//     adapters (databases, Redis, MinIO, external APIs), called by services.
//
// Every method takes context.Context first and uses only domain types.
//
// Example:
//
//	// Driving port
//	type OrderService interface {
//		PlaceOrder(ctx context.Context, items []domain.OrderItem) (*domain.Order, error)
//	}
//
//	// Driven ports
//	type OrderRepository interface {
//		Save(ctx context.Context, o *domain.Order) error
//		GetByID(ctx context.Context, id string) (*domain.Order, error) // domain.ErrNotFound if missing
//	}
//	type Cache interface {
//		Get(ctx context.Context, key string) ([]byte, error) // domain.ErrNotFound on miss
//		Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
//	}
package port
