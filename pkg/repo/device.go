package repo

import (
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
)

type deviceRepo interface {
	FindDevice(where WhereCondition) (*types.Device, error)
	UpdateDevice(device *types.Device, newDevice *types.Device) error
	CreateDevice(device *types.Device) error
	ListDevices(where WhereCondition) ([]*types.Device, error)
	RemoveDevice(where WhereCondition) error
}
