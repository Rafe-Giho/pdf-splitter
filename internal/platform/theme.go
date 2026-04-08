package platform

type ColorScheme int

const (
	ColorSchemeLight ColorScheme = iota
	ColorSchemeDark
)

func DetectSystemColorScheme() ColorScheme {
	return detectSystemColorScheme()
}
