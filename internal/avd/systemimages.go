package avd

// SystemImage represents an Android system image.
type SystemImage struct {
	APILevel     int
	AndroidVer   string // Human-readable version (e.g., "12")
	Target       string // "google_apis", "google_apis_playstore", "default"
	Architecture string // "x86_64", "arm64-v8a", "x86"
	DisplayName  string
	Package      string // Full package name for sdkmanager
}

// DefaultSystemImages returns a list of recommended system images.
func DefaultSystemImages() []SystemImage {
	return []SystemImage{
		// Android 12 (API 31) - Default recommended
		{
			APILevel:     31,
			AndroidVer:   "12",
			Target:       "google_apis",
			Architecture: "x86_64",
			DisplayName:  "Android 12 (Google APIs) - Intel x86_64 [Recommended]",
			Package:      "system-images;android-31;google_apis;x86_64",
		},
		{
			APILevel:     31,
			AndroidVer:   "12",
			Target:       "google_apis",
			Architecture: "arm64-v8a",
			DisplayName:  "Android 12 (Google APIs) - ARM64 v8a",
			Package:      "system-images;android-31;google_apis;arm64-v8a",
		},
		{
			APILevel:     31,
			AndroidVer:   "12",
			Target:       "google_apis_playstore",
			Architecture: "x86_64",
			DisplayName:  "Android 12 (Google Play) - Intel x86_64",
			Package:      "system-images;android-31;google_apis_playstore;x86_64",
		},

		// Android 13 (API 33)
		{
			APILevel:     33,
			AndroidVer:   "13",
			Target:       "google_apis",
			Architecture: "x86_64",
			DisplayName:  "Android 13 (Google APIs) - Intel x86_64",
			Package:      "system-images;android-33;google_apis;x86_64",
		},
		{
			APILevel:     33,
			AndroidVer:   "13",
			Target:       "google_apis",
			Architecture: "arm64-v8a",
			DisplayName:  "Android 13 (Google APIs) - ARM64 v8a",
			Package:      "system-images;android-33;google_apis;arm64-v8a",
		},
		{
			APILevel:     33,
			AndroidVer:   "13",
			Target:       "google_apis_playstore",
			Architecture: "x86_64",
			DisplayName:  "Android 13 (Google Play) - Intel x86_64",
			Package:      "system-images;android-33;google_apis_playstore;x86_64",
		},

		// Android 14 (API 34)
		{
			APILevel:     34,
			AndroidVer:   "14",
			Target:       "google_apis",
			Architecture: "x86_64",
			DisplayName:  "Android 14 (Google APIs) - Intel x86_64",
			Package:      "system-images;android-34;google_apis;x86_64",
		},
		{
			APILevel:     34,
			AndroidVer:   "14",
			Target:       "google_apis",
			Architecture: "arm64-v8a",
			DisplayName:  "Android 14 (Google APIs) - ARM64 v8a",
			Package:      "system-images;android-34;google_apis;arm64-v8a",
		},
		{
			APILevel:     34,
			AndroidVer:   "14",
			Target:       "google_apis_playstore",
			Architecture: "x86_64",
			DisplayName:  "Android 14 (Google Play) - Intel x86_64",
			Package:      "system-images;android-34;google_apis_playstore;x86_64",
		},

		// Android 11 (API 30) - For older app compatibility
		{
			APILevel:     30,
			AndroidVer:   "11",
			Target:       "google_apis",
			Architecture: "x86_64",
			DisplayName:  "Android 11 (Google APIs) - Intel x86_64",
			Package:      "system-images;android-30;google_apis;x86_64",
		},
		{
			APILevel:     30,
			AndroidVer:   "11",
			Target:       "google_apis",
			Architecture: "arm64-v8a",
			DisplayName:  "Android 11 (Google APIs) - ARM64 v8a",
			Package:      "system-images;android-30;google_apis;arm64-v8a",
		},
	}
}

// GetSystemImagesByArch filters system images by architecture.
func GetSystemImagesByArch(arch string) []SystemImage {
	var filtered []SystemImage
	for _, img := range DefaultSystemImages() {
		if img.Architecture == arch {
			filtered = append(filtered, img)
		}
	}
	return filtered
}

// GetSystemImagesByAPI filters system images by API level.
func GetSystemImagesByAPI(apiLevel int) []SystemImage {
	var filtered []SystemImage
	for _, img := range DefaultSystemImages() {
		if img.APILevel == apiLevel {
			filtered = append(filtered, img)
		}
	}
	return filtered
}

// GetDefaultImage returns the default recommended system image.
func GetDefaultImage() SystemImage {
	images := DefaultSystemImages()
	if len(images) > 0 {
		return images[0]
	}
	return SystemImage{}
}

// GetDefaultImageForArch returns the default image for a specific architecture.
func GetDefaultImageForArch(arch string) SystemImage {
	for _, img := range DefaultSystemImages() {
		if img.Architecture == arch && img.APILevel == 31 && img.Target == "google_apis" {
			return img
		}
	}
	// Fallback to first matching
	for _, img := range DefaultSystemImages() {
		if img.Architecture == arch {
			return img
		}
	}
	return GetDefaultImage()
}
