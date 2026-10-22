// Package domain holds the core business entities, value objects and
// domain errors.
//
// Rules:
//   - Import only the standard library (and other core packages).
//   - No framework, database, or transport concerns: no GORM/JSON/BSON tags.
//     Adapters map these types to their own persistence models and DTOs.
//
// Example:
//
//	type Order struct {
//		ID        string
//		Items     []OrderItem
//		CreatedAt time.Time
//	}
package domain
