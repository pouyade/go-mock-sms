# ADR 0001: Platform notification backend

- Status: Accepted
- Date: 2026-10-24

## Context

Desktop Notify needs native desktop notifications on macOS, Windows, and Linux.
The service must remain small, cross-compilable, and usable without CGO or a
database. Backend failures must be visible to SDK and HTTP callers.

## Decision

Use small platform-specific implementations built on operating-system tools:

- macOS: `osascript` with a fixed AppleScript program.
- Windows: Windows PowerShell and the WinRT toast API.
- Linux: `notify-send` from `libnotify`.

User-controlled title and message values are passed as process arguments or
environment variables. They are never interpolated into shell commands or
scripts. Every process exit status is checked and converted to an error.

## Alternatives considered

### `gen2brain/beeep`

Beeep is a reasonable cross-platform Go library, but it adds a dependency and
still relies on platform-specific behavior and runtime requirements. Keeping
the adapters in this repository makes those requirements explicit and keeps
the build CGO-free.

### Native framework bindings

Native framework bindings could provide tighter integration, but they would
increase CGO and cross-compilation requirements. That complexity is not needed
for the initial local notification service.

## Consequences

The binary is easy to cross-compile with `CGO_ENABLED=0`. Linux requires
`notify-send`, a graphical session, and a notification daemon. macOS
notifications are attributed to the `osascript`/Script Editor notification
identity, and Windows toast behavior depends on Windows notification settings.
Successful process invocation means the OS accepted the request; it does not
prove that a notification was visible or read.
