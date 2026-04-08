//go:build windows

package platform

import "golang.org/x/sys/windows/registry"

func detectSystemColorScheme() ColorScheme {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return ColorSchemeLight
	}
	defer key.Close()

	if value, _, err := key.GetIntegerValue("AppsUseLightTheme"); err == nil {
		if value == 0 {
			return ColorSchemeDark
		}
		return ColorSchemeLight
	}
	if value, _, err := key.GetIntegerValue("SystemUsesLightTheme"); err == nil {
		if value == 0 {
			return ColorSchemeDark
		}
		return ColorSchemeLight
	}

	return ColorSchemeLight
}
