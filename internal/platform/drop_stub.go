//go:build !windows

package platform

func AttachFileDrop(_ any, _ func(string)) error {
	return nil
}
