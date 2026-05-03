package contracts

import "embed"

// Schemas exposes the executable JSON Schema contracts shipped with OrchestraOS.
//
//go:embed schemas
var Schemas embed.FS

const SchemaRoot = "schemas"

func ReadSchema(path string) ([]byte, error) {
	return Schemas.ReadFile(SchemaRoot + "/" + path)
}
