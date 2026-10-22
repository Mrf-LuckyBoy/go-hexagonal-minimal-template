// Package adapters contains the concrete implementations that connect the
// core to frameworks and infrastructure. Put each adapter in its own
// sub-package, grouped by role, for example:
//
//	adapters/
//	├── handler/            driving adapters: call core services
//	│   ├── http/gin/       Gin, Echo, Fiber, net/http, ...
//	│   └── grpc/
//	├── repository/         driven adapters: implement repository ports
//	│   └── postgres/
//	├── cache/redis/        driven adapter: implements port.Cache
//	└── storage/minio/      driven adapter: implements port.ObjectStorage
//
// Rules:
//   - Adapters may import core packages; core packages never import adapters.
//   - Adapters never import each other; cmd/di.go wires them together.
//   - Translate infrastructure errors into domain errors at the boundary
//     (e.g. redis.Nil -> domain.ErrNotFound).
//
// See the README "Recipes" section for ready-to-copy adapters.
package adapters
