package tangent_sdk

import "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/mapper"

// Metadata describes a plugin's identifying information.
type Metadata struct {
	Name    string
	Version string
}

func (m Metadata) ToMapper() mapper.Meta {
	return mapper.Meta{
		Name:    m.Name,
		Version: m.Version,
	}
}
