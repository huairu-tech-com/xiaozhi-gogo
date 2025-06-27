package repo

import (
	"testing"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func randomDevice() *types.Device {
	return &types.Device{
		DeviceId: uuid.New().String(),
		ClientId: uuid.New().String(),
	}
}

func memoryRepository() Respository {
	m := NewInMemoryRepository()
	return m
}

func TestNewMemoryRepository(t *testing.T) {
	m := memoryRepository()
	assert.NotNil(t, m, "Expected non-nil repository")
}

func TestMemoryCreateDevice(t *testing.T) {
	m := memoryRepository()
	device := randomDevice()

	err := m.CreateDevice(device)
	assert.NoError(t, err, "Expected no error when creating device")

	// Verify that the device was created
	foundDevice, err := m.FindDevice(WhereCondition{"device_id": device.DeviceId})
	assert.NoError(t, err, "Expected no error when finding device")
	assert.Equal(t, device.DeviceId, foundDevice.DeviceId, "Expected found device to match created device")
}

func TestMemoryFindDeviceWithNilWhereCondition(t *testing.T) {
	m := memoryRepository()
	device := randomDevice()
	m.CreateDevice(device)

	where := WhereCondition(nil)
	fondDevice, err := m.FindDevice(where)
	assert.Error(t, err, "Expected error when finding device with nil where condition")
	assert.Nil(t, fondDevice, "Expected no device found with nil where condition")
}

func TestMemoryFindDeviceWithInvalidWhereCondition(t *testing.T) {
	m := memoryRepository()
	device := randomDevice()
	m.CreateDevice(device)

	where := WhereCondition{"invalid_key": "value"}
	fondDevice, err := m.FindDevice(where)
	assert.Error(t, err, "Expected error when finding device with invalid where condition")
	assert.Nil(t, fondDevice, "Expected no device found with invalid where condition")
}

func TestMemoryFindDevice(t *testing.T) {
	m := memoryRepository()
	device1 := randomDevice()
	m.CreateDevice(device1)
	device2 := randomDevice()
	m.CreateDevice(device2)

	where := WhereCondition{"device_id": device1.DeviceId}
	list, err := m.ListDevices(where)
	assert.NoError(t, err, "Expected no error when listing devices")
	assert.Len(t, list, 1, "Expected to find one device")
}

func TestMemoryUpdateDevice(t *testing.T) {
	m := memoryRepository()
	device := randomDevice()
	m.CreateDevice(device)

	newDevice := randomDevice()
	newDevice.DeviceId = device.DeviceId // Keep the same DeviceId for update

	assert.NotEqual(t, device.ClientId, newDevice.ClientId, "ClientId should be different for update")
	m.UpdateDevice(device, newDevice)

	foundDevice, err := m.FindDevice(WhereCondition{"device_id": device.DeviceId})
	assert.NoError(t, err, "Expected no error when finding updated device")
	assert.Equal(t, newDevice.ClientId, foundDevice.ClientId, "Expected updated device to have new ClientId")
}

func TestMemoryRemoveDevice(t *testing.T) {
	m := memoryRepository()
	device := randomDevice()
	m.CreateDevice(device)

	m.RemoveDevice(WhereCondition{"device_id": device.DeviceId})
	fountDevice, err := m.FindDevice(WhereCondition{"device_id": device.DeviceId})
	assert.Error(t, err, "Expected error when finding removed device")
	assert.Nil(t, fountDevice, "Expected no device found after removal")
}
