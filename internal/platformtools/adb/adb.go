package adb

import (
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	adbExecutable = platformtools.ADB
)

type Tool struct {
	executable string
	hostOS string
}

func New(path platformtools.PlatformToolsPath, hostOS string) (*Tool, error) {
	executable := filepath.Join(string(path), string(adbExecutable))
	if hostOS == "windows" {
		executable = executable + ".exe"
	}
	if _, err := os.Stat(executable); os.IsNotExist(err) {
		return nil, err
	}
	return &Tool{
		executable: executable,
		hostOS: hostOS,
	}, nil
}

func (t *Tool) GetDeviceIds() ([]string, error) {
	resp, err := t.command([]string{"devices"})
	if err != nil {
		return nil, err
	}
	devices := strings.Split(string(resp), "\n")
	devices = devices[1 : len(devices)-2]
	for i, device := range devices {
		devices[i] = strings.Split(device, "\t")[0]
	}
	return devices, nil
}

func (t *Tool) GetDeviceCodename(deviceId string) (string, error) {
	return t.getProp("ro.product.device", deviceId)
}

func (t *Tool) RebootIntoBootloader(deviceId string) error {
	_, err := t.command([]string{"-s", deviceId, "reboot", "bootloader"})
	if err != nil {
		return err
	}
	return nil
}

func (t *Tool) KillServer() error {
	_, err := t.command([]string{"kill-server"})
	if err != nil {
		return err
	}
	return nil
}

func (t *Tool) Name() platformtools.ToolName {
	return adbExecutable
}

func (t *Tool) command(args []string) ([]byte, error) {
	cmd := exec.Command(t.executable, args...)
	return cmd.CombinedOutput()
}

func (t *Tool) getProp(prop, deviceId string) (string, error) {
	resp, err := t.command([]string{"-s", deviceId, "shell", "getprop", prop})
	if err != nil {
		return "", err
	}
	return strings.Trim(string(resp), "[]\n\r"), nil
}
