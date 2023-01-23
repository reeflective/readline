//go:build windows
// +build windows

package editor

// EditBuffer is currently not supported on Windows operating systems.
func EditBuffer(buf []rune, filename, filetype string) ([]rune, error) {
	return buf, errors.New("Not currently supported on Windows")
}
