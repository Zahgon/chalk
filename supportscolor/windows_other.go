//go:build !windows

package supportscolor

// windowsLevel never applies off Windows, so detection continues with the
// terminal checks.
func windowsLevel() (int, bool) {
	return 0, false
}
