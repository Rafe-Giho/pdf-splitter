//go:build !windows

package platform

func detectSystemColorScheme() ColorScheme {
	return ColorSchemeLight
}
