package repo

import (
	"github.com/pkg/errors"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
)

var (
	ErrInvalidWhereCondition = errors.New("invalid where condition")
	ErrDeviceNotFound        = errors.New("device not found")
)

func IsNotExists(err error) bool {
	return errors.Is(err, ErrDeviceNotFound)
}

type deviceRepo interface {
	FindDevice(where WhereCondition) (*types.Device, error)
	UpdateDevice(device *types.Device, newDevice *types.Device) error
	CreateDevice(device *types.Device) error
	ListDevices(where WhereCondition) ([]*types.Device, error)
	RemoveDevice(where WhereCondition) error
}
