package avd

// DeviceProfile represents an Android device profile.
type DeviceProfile struct {
	Name    string
	ID      string
	Width   int
	Height  int
	Density int
}

// DefaultDeviceProfiles returns a list of common device profiles.
var DefaultDeviceProfiles = map[string]DeviceProfile{
	"pixel_6": {
		Name:    "Pixel 6",
		ID:      "pixel_6",
		Width:   1080,
		Height:  2400,
		Density: 411,
	},
	"pixel_6_pro": {
		Name:    "Pixel 6 Pro",
		ID:      "pixel_6_pro",
		Width:   1440,
		Height:  3120,
		Density: 512,
	},
	"pixel_7": {
		Name:    "Pixel 7",
		ID:      "pixel_7",
		Width:   1080,
		Height:  2400,
		Density: 416,
	},
	"pixel_7_pro": {
		Name:    "Pixel 7 Pro",
		ID:      "pixel_7_pro",
		Width:   1440,
		Height:  3120,
		Density: 512,
	},
	"pixel_5": {
		Name:    "Pixel 5",
		ID:      "pixel_5",
		Width:   1080,
		Height:  2340,
		Density: 432,
	},
	"pixel_4": {
		Name:    "Pixel 4",
		ID:      "pixel_4",
		Width:   1080,
		Height:  2280,
		Density: 440,
	},
	"pixel_4_xl": {
		Name:    "Pixel 4 XL",
		ID:      "pixel_4_xl",
		Width:   1440,
		Height:  3040,
		Density: 537,
	},
	"Nexus 5X": {
		Name:    "Nexus 5X",
		ID:      "Nexus 5X",
		Width:   1080,
		Height:  1920,
		Density: 420,
	},
	"Nexus 6P": {
		Name:    "Nexus 6P",
		ID:      "Nexus 6P",
		Width:   1440,
		Height:  2560,
		Density: 518,
	},
}

// GetDeviceProfile returns a device profile by ID.
func GetDeviceProfile(id string) (DeviceProfile, bool) {
	profile, ok := DefaultDeviceProfiles[id]
	return profile, ok
}

// GetDefaultDeviceProfile returns the default device (Pixel 6).
func GetDefaultDeviceProfile() DeviceProfile {
	return DefaultDeviceProfiles["pixel_6"]
}

// ListDeviceProfiles returns all available device profiles.
func ListDeviceProfiles() []DeviceProfile {
	profiles := make([]DeviceProfile, 0, len(DefaultDeviceProfiles))
	for _, p := range DefaultDeviceProfiles {
		profiles = append(profiles, p)
	}
	return profiles
}

// RecommendedDevices returns a list of recommended device IDs for pentesting.
func RecommendedDevices() []string {
	return []string{
		"pixel_6",  // Default recommended
		"pixel_7",  // Latest Pixel
		"pixel_5",  // Good balance
		"Nexus 5X", // Classic for testing
	}
}
