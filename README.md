# Desktop Notify

Desktop Notify is a small, cross-platform desktop notification library and
localhost HTTP service written in Go.

It has two entry points:

1. A Go SDK for applications that want to display native notifications directly.
2. The `desktop-notify` executable for applications written in any language.

This project displays desktop notifications only. It is **not an SMS gateway**:
it does not send SMS messages, connect to a telecom network, or require a
mobile phone.

<p>
  <a href="#english">English</a> ·
  <a href="#فارسی">فارسی</a>
</p>

## English

### Features

- Native operating-system notification mechanisms.
- Small public Go API with no runtime Go dependencies.
- Local HTTP API on `127.0.0.1:4466` by default.
- macOS, Windows, and Linux build targets.
- No GUI, database, message history, persistent queue, or background worker.
- Input validation, request-size limits, timeouts, graceful shutdown, and
  structured JSON responses.
- GitHub Actions builds and publishes versioned release archives.

### Requirements

| Platform | Runtime requirement |
| --- | --- |
| macOS | `osascript`, an active graphical user session, and notification permission |
| Windows | Windows PowerShell and Windows desktop notifications enabled |
| Linux | `notify-send` from `libnotify`, a graphical session, `DBUS_SESSION_BUS_ADDRESS`, and a notification daemon |

The Go package itself is CGO-free. Native notifications still depend on the
operating-system tools and desktop session described above.

### Installation

#### Go SDK

```bash
go get github.com/pouyade/go-mock-sms
```

Import the module with the package name `notify`:

```go
import notify "github.com/pouyade/go-mock-sms"
```

#### Standalone binary

Download the archive for your operating system and architecture from the
[GitHub Releases](https://github.com/pouyade/go-mock-sms/releases) page.
Each release contains:

- `desktop-notify-darwin-arm64.tar.gz`
- `desktop-notify-darwin-amd64.tar.gz`
- `desktop-notify-windows-amd64.zip`
- `desktop-notify-windows-arm64.zip`
- `desktop-notify-linux-amd64.tar.gz`
- `desktop-notify-linux-arm64.tar.gz`
- `desktop-notify-macos-universal.app.zip`
- `desktop-notify-macos-universal.dmg`
- `checksums.txt`

Verify an archive on macOS or Linux:

```bash
sha256sum -c checksums.txt
```

On macOS, use `shasum -a 256 -c checksums.txt` if `sha256sum` is unavailable.
Windows users can use `Get-FileHash` in PowerShell.

### Go SDK usage

The simplest call takes a title and body:

```go
package main

import (
	"log"

	notify "github.com/pouyade/go-mock-sms"
)

func main() {
	if err := notify.Send("Verification code", "Your code is 123456"); err != nil {
		log.Fatal(err)
	}
}
```

For a reusable client:

```go
client := notify.New()

err := client.Send(notify.Message{
	Title: "ExampleApp",
	Text:  "Your verification code is 123456",
})
if err != nil {
	// Handle the backend or validation error.
}
```

`Client` is safe for concurrent use. It sends each request immediately; it
does not create a queue or persist messages.

Validation rules:

- `Text` is required and cannot be whitespace-only.
- Text is limited to `1024` Unicode code points.
- Titles are limited to `128` Unicode code points.
- Empty titles use `notify.DefaultTitle` (`Desktop Notify`).
- Newline, carriage return, and tab are allowed. Other control characters are
  rejected.

The exported sentinel errors can be inspected with `errors.Is`:

```go
if errors.Is(err, notify.ErrEmptyText) {
	// Ask the caller for a message body.
}
```

The SDK never logs notification content. A successful return means that the
operating-system mechanism accepted the request; it does not guarantee that a
notification was visible or read.

### HTTP service

Start the service:

```bash
./desktop-notify
```

The default listener is `127.0.0.1:4466`.

Send a notification:

```bash
curl -X POST http://127.0.0.1:4466/send \
  -H "Content-Type: application/json" \
  -d '{"to":"number","text":"Hello from an HTTP client"}'
```

`to` is optional and retained only for API compatibility. It is not an SMS
recipient and has no effect. `title` is optional:

```bash
curl -X POST http://127.0.0.1:4466/send \
  -H "Content-Type: application/json" \
  -d '{"title":"Build bot","text":"Build completed"}'
```

Success response:

```json
{
  "success": true,
  "message": "Notification request accepted"
}
```

Health check:

```bash
curl http://127.0.0.1:4466/health
```

Response:

```json
{"status":"ok"}
```

`/health` reports that the HTTP process is responsive. It does not test the
desktop notification backend.

#### HTTP status codes

| Status | Meaning |
| --- | --- |
| `200` | The notification mechanism accepted the request |
| `400` | Invalid JSON or invalid notification fields |
| `404` | Unknown endpoint |
| `405` | Unsupported HTTP method |
| `413` | Request body is larger than 64 KiB |
| `415` | `Content-Type` is not `application/json` |
| `500` | The notification backend failed |

Error responses have a stable JSON shape:

```json
{
  "success": false,
  "error": "notification_failed",
  "message": "The operating system notification backend failed"
}
```

The service does not return stack traces, local paths, notification bodies, or
other sensitive backend details.

### Configuration

Command-line flags:

```text
--host    HTTP listen host (default: 127.0.0.1)
--port    HTTP listen port (default: 4466)
--title   Default notification title (default: Desktop Notify)
--version Print version, commit, and build date, then exit
```

Equivalent environment variables:

```text
DESKTOP_NOTIFY_HOST
DESKTOP_NOTIFY_PORT
DESKTOP_NOTIFY_TITLE
```

Precedence is flag, then environment variable, then the built-in default.
Examples:

```bash
DESKTOP_NOTIFY_PORT=9000 ./desktop-notify
./desktop-notify --host 127.0.0.1 --port 9000 --title "My App"
./desktop-notify --version
```

The server warns when `--host` is not a recognized loopback address. Binding
to a non-loopback address allows other machines to call the endpoint and is
not protected by authentication or TLS. Do this only behind appropriate
network controls.

### Platform notes

#### macOS

The backend invokes the built-in `osascript` command with a fixed AppleScript
program. Message values are passed as arguments rather than interpolated into
the script.

The first notification may require permission in **System Settings >
Notifications**. Notification identity and presentation are controlled by
macOS and may appear as `osascript` or Script Editor rather than as a
packaged application. A successful `osascript` process means macOS accepted
the request, not that the banner was shown.

Both Intel (`amd64`) and Apple Silicon (`arm64`) release archives are provided.
The GitHub Release also includes a universal `Desktop Notify.app` archive and
an optional disk image installer. The app is a background application; it does
not provide a graphical dashboard.

#### Windows

The backend invokes Windows PowerShell and uses the Windows toast API. The
PowerShell script is fixed and title/body values are passed through environment
variables and XML-escaped before use.

Windows notification settings, Focus Assist, policy, and the current desktop
session can affect visibility. The lightweight AppUserModelID used by this
project avoids requiring an installer, but Windows packaging or policy
limitations may still affect the notification identity.

AMD64 and ARM64 builds are provided. The executable is named
`desktop-notify.exe`.

#### Linux

Install `notify-send` on Debian or Ubuntu with:

```bash
sudo apt-get install libnotify-bin
```

Other distributions provide the same command through their `libnotify`
package. Linux needs an active graphical session, a notification daemon, and
`DBUS_SESSION_BUS_ADDRESS`. Headless SSH shells, cron jobs without a desktop
environment, and containers normally cannot display a notification.

The backend invokes `notify-send` directly with separate arguments; it does
not use a shell and does not interpolate user input into a command.

Linux release archives target AMD64 and ARM64. The binary does not bundle a
notification daemon.

### Running in the background

Manual startup is the supported baseline. Native launch mechanisms are
optional conveniences.

#### macOS launchd

Create `~/Library/LaunchAgents/com.pouyade.desktop-notify.plist`, replacing
the binary path if needed. If the `.dmg` release was installed into
`/Applications`, use the app executable path below:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.pouyade.desktop-notify</string>
  <key>ProgramArguments</key>
  <array>
    <string>/Applications/Desktop Notify.app/Contents/MacOS/desktop-notify</string>
    <string>--host</string>
    <string>127.0.0.1</string>
    <string>--port</string>
    <string>4466</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
</dict>
</plist>
```

Load it with:

```bash
launchctl load -w ~/Library/LaunchAgents/com.pouyade.desktop-notify.plist
```

#### Windows

Use Task Scheduler to run `desktop-notify.exe` at user logon, or use an
organization-approved service wrapper. Configure the task to run in the
interactive user session; desktop notifications are not normally visible
from a non-interactive service session.

#### Linux systemd user service

Create `~/.config/systemd/user/desktop-notify.service`:

```ini
[Unit]
Description=Desktop Notify local HTTP service

[Service]
ExecStart=%h/bin/desktop-notify
Restart=on-failure

[Install]
WantedBy=default.target
```

Enable it from a graphical user session:

```bash
systemctl --user daemon-reload
systemctl --user enable --now desktop-notify.service
```

Ensure the service receives the graphical session's D-Bus environment.

### Development

Requirements: Go 1.27 or newer and `make`.

```bash
make fmt
make test
make vet
make check
make build
make run
```

The test suite uses mocks and `httptest`; it does not require a desktop
session. The race-enabled suite is:

```bash
go test -race ./...
```

The Makefile uses `find`, `gofmt`, and standard POSIX shell commands. It works
with GNU Make and the BSD make commonly installed on macOS; on Windows, use
the Go commands directly or run Make in a POSIX-compatible environment.

### Releases

The workflow in `.github/workflows/deploy.yml` validates formatting, modules,
vetting, race tests, static analysis, vulnerability checks, and all six
cross-build targets. A Linux AMD64 package also runs `--version` and a
`/health` smoke test. Pushes to `main` upload the six packaged binaries and
macOS packages as GitHub Actions artifacts; version tags additionally publish
a GitHub Release containing the six platform archives, the macOS `.app` zip,
the macOS `.dmg`, and checksums.

The macOS package is ad-hoc signed by default. For public Gatekeeper-friendly
distribution, configure these GitHub repository secrets:

- `MACOS_SIGNING_CERTIFICATE_BASE64`
- `MACOS_SIGNING_CERTIFICATE_PASSWORD`
- `MACOS_SIGNING_IDENTITY`
- `MACOS_KEYCHAIN_PASSWORD`
- `APPLE_ID`
- `APPLE_TEAM_ID`
- `APPLE_APP_PASSWORD`

When all secrets are present, the workflow imports the Developer ID
certificate, signs the app, notarizes the DMG, and staples the notarization
ticket. Without them, the DMG is still built but users may need to approve it
manually in macOS security settings.

Push a version tag to publish a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Manual workflow dispatch runs validation and packaging but does not publish a
release. Release publishing is restricted to `v*` tags and the publishing job
has `contents: write`; other jobs have read-only repository permissions.

### Security and privacy

- The default listener is loopback-only.
- Any local process that can access `127.0.0.1:4466` can submit notifications.
- Remote binding is intentionally not authenticated; do not expose it directly
  to an untrusted network.
- Request bodies are limited to 64 KiB and HTTP timeouts are configured.
- Messages are not persisted, queued, or logged.
- Do not assume a notification is private: operating-system settings may show
  it on a locked screen.
- Do not put secrets or OTP values in shell history or shared logs when using
  the HTTP endpoint.

### Limitations

- This is a desktop notification tool, not an SMS sender.
- Visibility depends on operating-system settings, permissions, policy, and an
  active desktop session.
- Successful SDK or HTTP responses mean that the backend accepted the request;
  they do not prove that a person saw it.
- There is no authentication, persistence, retry queue, scheduling,
  attachment support, or message history.
- Linux requires an external `notify-send` installation.
- Windows and macOS notification identity is controlled by their native
  command-line mechanisms.

## فارسی

### معرفی

Desktop Notify یک کتابخانهٔ کوچک Go و یک سرویس HTTP محلی برای نمایش اعلان‌های
بومی دسکتاپ در macOS، Windows و Linux است.

این پروژه دو روش استفاده دارد:

1. کتابخانهٔ Go برای برنامه‌هایی که مستقیماً اعلان دسکتاپ می‌فرستند.
2. فایل اجرایی `desktop-notify` برای برنامه‌هایی که با هر زبان دیگری نوشته
   شده‌اند و از HTTP استفاده می‌کنند.

این پروژه **درگاه پیامک نیست**. پیامک واقعی ارسال نمی‌کند، به شبکهٔ مخابراتی
متصل نمی‌شود و به تلفن همراه نیاز ندارد.

### قابلیت‌ها

- استفاده از سازوکار اعلان خود سیستم‌عامل.
- API کوچک Go بدون وابستگی زمان اجرا.
- سرویس HTTP فقط روی `127.0.0.1:4466` در حالت پیش‌فرض.
- هدف‌های ساخت برای macOS، Windows و Linux.
- بدون رابط گرافیکی، پایگاه داده، تاریخچه، صف پایدار یا ذخیره‌سازی پیام.
- اعتبارسنجی ورودی، محدودیت اندازه، timeout، خاموش‌شدن graceful و پاسخ JSON.
- ساخت و انتشار خودکار نسخه‌ها با GitHub Actions.

### نصب کتابخانهٔ Go

```bash
go get github.com/pouyade/go-mock-sms
```

نام پکیج پس از import، `notify` است:

```go
import notify "github.com/pouyade/go-mock-sms"
```

### استفاده از SDK

ساده‌ترین نمونه:

```go
package main

import (
	"log"

	notify "github.com/pouyade/go-mock-sms"
)

func main() {
	if err := notify.Send("کد تأیید", "کد شما ۱۲۳۴۵۶ است"); err != nil {
		log.Fatal(err)
	}
}
```

برای استفادهٔ مجدد از یک client:

```go
client := notify.New()
err := client.Send(notify.Message{
	Title: "برنامهٔ نمونه",
	Text:  "کد تأیید شما ۱۲۳۴۵۶ است",
})
```

`Client` برای استفادهٔ هم‌زمان امن است. پیام‌ها فوراً ارسال می‌شوند و در صف
قرار نمی‌گیرند و ذخیره نمی‌شوند.

قوانین اعتبارسنجی:

- فیلد `Text` اجباری است و نباید فقط شامل فاصله باشد.
- متن حداکثر ۱۰۲۴ کاراکتر Unicode دارد.
- عنوان حداکثر ۱۲۸ کاراکتر Unicode دارد.
- عنوان خالی به `Desktop Notify` تبدیل می‌شود.
- newline، carriage return و tab مجاز هستند؛ سایر کاراکترهای کنترلی رد می‌شوند.

خطاهای SDK را می‌توان با `errors.Is` بررسی کرد؛ برای نمونه
`notify.ErrEmptyText` زمانی برگردانده می‌شود که متن خالی باشد.

### استفاده از سرویس HTTP

اجرای سرویس:

```bash
./desktop-notify
```

ارسال اعلان:

```bash
curl -X POST http://127.0.0.1:4466/send \
  -H "Content-Type: application/json" \
  -d '{"to":"number","text":"سلام از یک client HTTP"}'
```

فیلد `to` اختیاری است و فقط برای سازگاری API نگه داشته شده؛ گیرندهٔ پیامک
نیست و اثری روی اعلان ندارد. فیلد `title` نیز اختیاری است:

```bash
curl -X POST http://127.0.0.1:4466/send \
  -H "Content-Type: application/json" \
  -d '{"title":"ربات ساخت","text":"ساخت پروژه تمام شد"}'
```

پاسخ موفق:

```json
{"success":true,"message":"Notification request accepted"}
```

بررسی سلامت سرویس:

```bash
curl http://127.0.0.1:4466/health
```

پاسخ:

```json
{"status":"ok"}
```

مسیر `/health` فقط پاسخ‌گو بودن فرایند HTTP را بررسی می‌کند و سلامت backend
اعلان سیستم‌عامل را تضمین نمی‌کند.

کدهای HTTP:

- `200`: سیستم‌عامل درخواست اعلان را پذیرفته است.
- `400`: JSON یا فیلدهای اعلان نامعتبر هستند.
- `404`: مسیر پیدا نشد.
- `405`: متد HTTP پشتیبانی نمی‌شود.
- `413`: اندازهٔ بدنه بیشتر از ۶۴ کیلوبایت است.
- `415`: نوع محتوا `application/json` نیست.
- `500`: backend اعلان سیستم‌عامل خطا داده است.

خطاها با JSON ساختاریافته برگردانده می‌شوند و شامل stack trace، مسیرهای داخلی
یا محتوای پیام نیستند.

### تنظیمات

پارامترهای خط فرمان:

```text
--host     میزبان HTTP، پیش‌فرض 127.0.0.1
--port     پورت HTTP، پیش‌فرض 4466
--title    عنوان پیش‌فرض اعلان، پیش‌فرض Desktop Notify
--version  چاپ نسخه و خروج
```

متغیرهای محیطی متناظر:

```text
DESKTOP_NOTIFY_HOST
DESKTOP_NOTIFY_PORT
DESKTOP_NOTIFY_TITLE
```

اولویت تنظیمات به ترتیب flag، متغیر محیطی و سپس مقدار پیش‌فرض است. اگر میزبان
غیر loopback انتخاب شود، برنامه هشدار امنیتی چاپ می‌کند. برای دسترسی شبکه‌ای
احراز هویت یا TLS داخلی وجود ندارد.

### نکات سیستم‌عامل

#### macOS

از `osascript` داخلی macOS استفاده می‌شود. ممکن است اولین اعلان به اجازه در
**System Settings > Notifications** نیاز داشته باشد. نام برنامهٔ نمایش‌دهنده
را macOS تعیین می‌کند و ممکن است `osascript` یا Script Editor دیده شود.

هر دو معماری Intel (`amd64`) و Apple Silicon (`arm64`) ارائه می‌شوند.

#### Windows

از Windows PowerShell و Windows toast API استفاده می‌شود. فعال بودن اعلان‌ها،
Focus Assist، سیاست‌های سازمانی و وجود session گرافیکی روی دیده شدن اعلان اثر
دارند. فایل اجرایی با نام `desktop-notify.exe` و برای AMD64 و ARM64 منتشر می‌شود.

#### Linux

در Debian یا Ubuntu نصب کنید:

```bash
sudo apt-get install libnotify-bin
```

وجود یک session گرافیکی فعال، notification daemon و متغیر
`DBUS_SESSION_BUS_ADDRESS` لازم است. اجرای سرویس در SSH بدون محیط دسکتاپ،
cron یا container معمولاً اعلان قابل مشاهده تولید نمی‌کند.

### اجرای پس‌زمینه

اجرای دستی روش پایه و پشتیبانی‌شده است. برای macOS می‌توان از launchd، برای
Windows از Task Scheduler در session تعاملی و برای Linux از systemd user
service استفاده کرد. نمونه‌های کامل این تنظیمات در بخش انگلیسی همین README
آمده‌اند.

### توسعه و انتشار

دستورات توسعه:

```bash
make fmt
make test
make vet
make check
make build
make run
go test -race ./...
```

تست‌ها به desktop واقعی نیاز ندارند و از mock و `httptest` استفاده می‌کنند.
برای انتشار:

```bash
git tag v0.1.0
git push origin v0.1.0
```

workflow ابتدا format، test، vet، race test، تحلیل ایستا، بررسی آسیب‌پذیری و
ساخت هر شش معماری را اجرا می‌کند؛ سپس برای tagهای `v*` فایل‌های release و
`checksums.txt` را منتشر می‌کند. Push روی شاخهٔ `main` نیز هر شش فایل binary را
به‌عنوان artifact در GitHub Actions بارگذاری می‌کند. برای macOS علاوه بر archive،
یک فایل universal با پسوند `.app.zip` و یک فایل `.dmg` نیز ساخته می‌شود.

به‌صورت پیش‌فرض امضای macOS به‌صورت ad-hoc است. برای امضای رسمی Developer ID و
notarization باید secretهای امضای Apple را در تنظیمات GitHub repository قرار دهید.
در صورت نبود این secretها، فایل DMG ساخته می‌شود اما ممکن است macOS برای اجرای
آن تأیید دستی بخواهد.

### امنیت و محدودیت‌ها

- listener پیش‌فرض فقط روی loopback است.
- هر فرایند محلی دارای دسترسی به پورت می‌تواند اعلان ارسال کند.
- پیام‌ها ذخیره یا log نمی‌شوند.
- اعلان ممکن است طبق تنظیمات سیستم‌عامل روی صفحهٔ قفل دیده شود.
- پاسخ موفق فقط پذیرش درخواست توسط backend را نشان می‌دهد، نه دیده شدن آن.
- احراز هویت، صف retry، زمان‌بندی، attachment و تاریخچهٔ پیام وجود ندارد.
- Linux به نصب خارجی `notify-send` نیاز دارد و هر سه سیستم‌عامل به session
  دسکتاپ فعال و تنظیمات اعلان خود وابسته‌اند.
