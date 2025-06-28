package repo

import (
	"sync"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
)

type InMemoryRepository struct {
	devices sync.Map // Using sync.Map for concurrent access
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		devices: sync.Map{},
	}
}

func (r *InMemoryRepository) FindDevice(where WhereCondition) (*types.Device, error) {
	deviceId, ok := where["device_id"].(string)
	if !ok {
		return nil, ErrInvalidWhereCondition
	}

	obj, ok := r.devices.Load(deviceId)
	if !ok {
		return nil, ErrDeviceNotFound
	}

	dev, ok := obj.(*types.Device)
	if !ok {
		return nil, ErrDeviceNotFound
	}

	return dev, nil
}

func (r *InMemoryRepository) UpdateDevice(device *types.Device, newDevice *types.Device) error {
	r.devices.CompareAndSwap(device.DeviceId, device, newDevice)
	return nil
}

func (r *InMemoryRepository) CreateDevice(device *types.Device) error {
	r.devices.Store(device.DeviceId, device)
	return nil
}

func (r *InMemoryRepository) ListDevices(where WhereCondition) ([]*types.Device, error) {
	devices := make([]*types.Device, 0)

	r.devices.Range(func(key, value any) bool {
		dev, ok := value.(*types.Device)
		if !ok {
			return false // Skip if the value is not a Device
		}

		if where.MatchDevice(dev) {
			devices = append(devices, dev)
		}
		return true
	})

	return devices, nil
}

func (r *InMemoryRepository) RemoveDevice(where WhereCondition) error {
	devices, err := r.ListDevices(where)
	if err != nil {
		return err
	}

	for _, device := range devices {
		r.devices.Delete(device.DeviceId)
	}
	return nil
}
