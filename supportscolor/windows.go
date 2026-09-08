//go:build windows

package supportscolor

import "golang.org/x/sys/windows"

// windowsLevel reports the color level implied by the Windows version.
//
// Windows 10 build 10586 added 256-color support and build 14931 added 24-bit
// color. Earlier versions get the basic palette. On Windows this always has an
// answer, so the check never falls through to terminal sniffing.
//
// The JavaScript reads os.release(), which on Windows is derived from the same
// RtlGetVersion call used here.
func windowsLevel() (int, bool) {
	version := windows.RtlGetVersion()

	if version.MajorVersion >= 10 && version.BuildNumber >= 10586 {
		if version.BuildNumber >= 14931 {
			return LevelTrueColor, true
		}

		return LevelAnsi256, true
	}

	return LevelBasic, true
}
