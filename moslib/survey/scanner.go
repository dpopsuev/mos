package survey

import "github.com/dpopsuev/mos/moslib/model"

// Scanner extracts structural metadata from source code.
type Scanner interface {
	Scan(root string) (*model.Project, error)
}
