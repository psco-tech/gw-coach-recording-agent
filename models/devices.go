package models

import (
	"log"

	"github.com/psco-tech/gw-coach-recording-agent/pbx"
)

var devices = &DevicesRepository{}

type DevicesRepository struct {
	reloadHandlers []func() error
	devices        []pbx.Device
}

func GetDevicesRepository() *DevicesRepository {
	return devices
}

func (d *DevicesRepository) OnReload(handler func() error) {
	d.reloadHandlers = append(d.reloadHandlers, handler)
}

func (d *DevicesRepository) Reload() error {
	log.Printf("Reloading devices\n")
	for _, h := range d.reloadHandlers {
		err := h()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DevicesRepository) Clear() {
	log.Printf("Clearing device repository\n")
	d.devices = make([]pbx.Device, 0)
}

func (d *DevicesRepository) Add(device pbx.Device) {
	log.Printf("Adding device %s to repository\n", device)
	d.devices = append(d.devices, device)
}
