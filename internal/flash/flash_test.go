package flash

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.com/calyxos/device-flasher/internal/factoryimage"
	"gitlab.com/calyxos/device-flasher/internal/flash/mocks"
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"gitlab.com/calyxos/device-flasher/internal/platformtools/fastboot"
	"testing"
)

var (
	testOS = "TestOS"
)

func TestFlash(t *testing.T) {
	ctrl := gomock.NewController(t)

	testDevice := &Device{ID: "8AAY0GK9A", Codename: "crosshatch", DiscoveryTool: platformtools.ADB}
	testDeviceFastboot := &Device{ID: "8AAY0GK9A", Codename: "crosshatch", DiscoveryTool: platformtools.Fastboot}

	tests := map[string]struct {
		device  *Device
		prepare func(*mocks.MockFactoryImageFlasher, *mocks.MockPlatformToolsFlasher,
			*mocks.MockADBFlasher, *mocks.MockFastbootFlasher)
		expectedErr error
	}{
		"happy path flash successful": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testDevice.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"device discovered through fastboot skips adb reboot": {
			device: testDeviceFastboot,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDeviceFastboot.Codename).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDeviceFastboot.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDeviceFastboot.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDeviceFastboot.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testDeviceFastboot.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"unlocked device skips unlocking step": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Unlocked, nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testDevice.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"factory image validation failure": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(factoryimage.ErrorValidation)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: factoryimage.ErrorValidation,
		},
		"adb reboot bootloader error not fatal": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(errors.New("not fatal"))
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testDevice.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"get bootloader status failure": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Unknown, fastboot.ErrorCommandFailure)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: fastboot.ErrorCommandFailure,
		},
		"set bootloader status unlock failure": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Unlocked).Return(fastboot.ErrorUnlockBootloader)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: fastboot.ErrorUnlockBootloader,
		},
		"flash all error": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(factoryimage.ErrorFailedToFlash)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: factoryimage.ErrorFailedToFlash,
		},
		"set bootloader status lock failure": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Locked).Return(fastboot.ErrorLockBootloader)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: fastboot.ErrorLockBootloader,
		},
		"reboot error is not fatal": {
			device: testDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testDevice.ID).Return(fastboot.ErrorRebootFailure)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockFactoryImage := mocks.NewMockFactoryImageFlasher(ctrl)
			mockPlatformTools := mocks.NewMockPlatformToolsFlasher(ctrl)
			mockADB := mocks.NewMockADBFlasher(ctrl)
			mockFastboot := mocks.NewMockFastbootFlasher(ctrl)

			if tc.prepare != nil {
				tc.prepare(mockFactoryImage, mockPlatformTools, mockADB, mockFastboot)
			}

			flash := New(&Config{
				HostOS:        testOS,
				FactoryImage:  mockFactoryImage,
				PlatformTools: mockPlatformTools,
				ADB:           mockADB,
				Fastboot:      mockFastboot,
				Logger:        logrus.StandardLogger(),
			})

			err := flash.Flash(tc.device)
			if tc.expectedErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.expectedErr), true)
			}
		})
	}
}

func TestDiscoverDevices(t *testing.T) {
	ctrl := gomock.NewController(t)

	testDeviceADB := &Device{ID: "serialadb", Codename: "adb", DiscoveryTool: platformtools.ADB}
	testDuplicateFastboot := &Device{ID: "serialadb", Codename: "fastboot", DiscoveryTool: platformtools.Fastboot}
	testDeviceFastboot := &Device{ID: "serialfastboot", Codename: "fastboot", DiscoveryTool: platformtools.Fastboot}

	tests := map[string]struct {
		device  *Device
		prepare func(*mocks.MockFactoryImageFlasher, *mocks.MockPlatformToolsFlasher,
			*mocks.MockADBFlasher, *mocks.MockFastbootFlasher)
		expectedErr     error
		expectedDevices map[string]*Device
	}{
		"discovery successful with adb device": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return(testDeviceADB.Codename, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return(nil, nil)
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr:     nil,
			expectedDevices: map[string]*Device{testDeviceADB.ID: testDeviceADB},
		},
		"discovery successful with fastboot device": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return(nil, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDeviceFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDeviceFastboot.ID).Return(testDeviceFastboot.Codename, nil)
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr:     nil,
			expectedDevices: map[string]*Device{testDeviceFastboot.ID: testDeviceFastboot},
		},
		"discovery successful with both adb and fastboot devices": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return(testDeviceADB.Codename, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDeviceFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDeviceFastboot.ID).Return(testDeviceFastboot.Codename, nil)
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr: nil,
			expectedDevices: map[string]*Device{
				testDeviceADB.ID:      testDeviceADB,
				testDeviceFastboot.ID: testDeviceFastboot,
			},
		},
		"discovery fails when get device returns empty for both adb and fastboot": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{}, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{}, nil)
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr:     ErrNoDevicesFound,
			expectedDevices: nil,
		},
		"discovery fails when get device fails for both adb and fastboot": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return(nil, errors.New("failed"))
				mockFastboot.EXPECT().GetDeviceIds().Return(nil, errors.New("failed"))
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr:     ErrNoDevicesFound,
			expectedDevices: nil,
		},
		"duplicate fastboot device overwrites existing adb device": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return(testDeviceADB.Codename, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDuplicateFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDuplicateFastboot.ID).Return(testDuplicateFastboot.Codename, nil)
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr:     nil,
			expectedDevices: map[string]*Device{testDuplicateFastboot.ID: testDuplicateFastboot},
		},
		"device in not added if get codename fails": {
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return("", errors.New("fail"))
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDeviceFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDeviceFastboot.ID).Return("", errors.New("fail"))
				mockADB.EXPECT().Name().Return(platformtools.ADB)
				mockFastboot.EXPECT().Name().Return(platformtools.Fastboot)
			},
			expectedErr:     ErrNoDevicesFound,
			expectedDevices: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockFactoryImage := mocks.NewMockFactoryImageFlasher(ctrl)
			mockPlatformTools := mocks.NewMockPlatformToolsFlasher(ctrl)
			mockADB := mocks.NewMockADBFlasher(ctrl)
			mockFastboot := mocks.NewMockFastbootFlasher(ctrl)

			if tc.prepare != nil {
				tc.prepare(mockFactoryImage, mockPlatformTools, mockADB, mockFastboot)
			}

			flash := New(&Config{
				HostOS:        testOS,
				FactoryImage:  mockFactoryImage,
				PlatformTools: mockPlatformTools,
				ADB:           mockADB,
				Fastboot:      mockFastboot,
				Logger:        logrus.StandardLogger(),
			})

			devices, err := flash.DiscoverDevices()
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedDevices, devices)
		})
	}
}