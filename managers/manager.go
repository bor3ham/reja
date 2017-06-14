package managers

import (
	"github.com/bor3ham/reja/instances"
)

type Manager interface {
	Create() instances.Instance
}
