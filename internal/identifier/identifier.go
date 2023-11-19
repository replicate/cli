package identifier

import (
	"fmt"
	"strings"
)

// Identifier is a model identifier

type Identifier struct {
	// Owner
	Owner string

	// Name
	Name string

	// Version (optional)
	Version string
}

func ParseIdentifier(s string) (*Identifier, error) {
	identifier := &Identifier{}

	// TODO validate owner, name, version formats

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid model identifier: %s", s)
	}

	identifier.Owner = parts[0]
	parts = strings.Split(parts[1], ":")
	switch len(parts) {
	case 1:
		identifier.Name = parts[0]
	case 2:
		identifier.Name = parts[0]
		identifier.Version = parts[1]
	default:
		return nil, fmt.Errorf("invalid model identifier: %s", s)
	}

	return identifier, nil
}

func (i *Identifier) String() string {
	if i.Version == "" {
		return fmt.Sprintf("%s/%s", i.Owner, i.Name)
	}

	return fmt.Sprintf("%s/%s:%s", i.Owner, i.Name, i.Version)
}

func (i *Identifier) Validate() error {
	if i.Owner == "" {
		return fmt.Errorf("owner must be set")
	}
	if i.Name == "" {
		return fmt.Errorf("name must be set")
	}
	return nil
}
