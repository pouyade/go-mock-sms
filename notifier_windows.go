//go:build windows

package notify

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"unicode/utf16"
)

type windowsNotifier struct{}

const windowsToastScript = `
$ErrorActionPreference = "Stop"
$title = [System.Security.SecurityElement]::Escape($env:DESKTOP_NOTIFY_TITLE)
$text = [System.Security.SecurityElement]::Escape($env:DESKTOP_NOTIFY_TEXT)
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml("<toast><visual><binding template='ToastGeneric'><text>$title</text><text>$text</text></binding></visual></toast>")
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Desktop Notify").Show($toast)
`

func newPlatformNotifier() notifier {
	return windowsNotifier{}
}

func (windowsNotifier) send(ctx context.Context, message Message) error {
	if !commandExists("powershell.exe") {
		return fmt.Errorf("%w: powershell.exe was not found", ErrBackendUnavailable)
	}
	encoded := base64.StdEncoding.EncodeToString(utf16Bytes(windowsToastScript))
	commandCtx, cancel := context.WithTimeout(ctx, backendTimeout)
	defer cancel()
	command := exec.CommandContext(commandCtx, "powershell.exe", "-NoProfile", "-NonInteractive",
		"-EncodedCommand", encoded)
	command.Env = append(os.Environ(),
		"DESKTOP_NOTIFY_TITLE="+message.Title,
		"DESKTOP_NOTIFY_TEXT="+message.Text,
	)
	if output, err := command.CombinedOutput(); err != nil {
		if commandCtx.Err() != nil {
			return fmt.Errorf("%w: powershell.exe timed out", ErrBackendUnavailable)
		}
		if len(output) > 0 {
			return fmt.Errorf("%w: %s", ErrBackendUnavailable, output)
		}
		return fmt.Errorf("%w: powershell.exe failed", ErrBackendUnavailable)
	}
	return nil
}

func utf16Bytes(value string) []byte {
	encoded := utf16.Encode([]rune(value))
	result := make([]byte, len(encoded)*2)
	for index, word := range encoded {
		result[index*2] = byte(word)
		result[index*2+1] = byte(word >> 8)
	}
	return result
}
