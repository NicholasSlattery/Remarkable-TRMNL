package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var version = "dev"

//go:embed web/index.html web/icon.svg
var webFiles embed.FS

type server struct {
	csrf        string
	payloadDir  string
	operationMu sync.Mutex
	known       *knownHosts
}

type credentials struct {
	Host        string `json:"host"`
	Password    string `json:"password"`
	Fingerprint string `json:"fingerprint"`
	// AcceptNewKey lets the user replace a remembered key after confirming the
	// change out of band. Developer Mode factory-resets the tablet and issues a
	// new host key, so without this a legitimate reset would lock them out.
	AcceptNewKey bool `json:"accept_new_key"`
}

// expectedFingerprint prefers the key remembered from a previous successful
// install. The key echoed back by the page is only trusted when the user has
// explicitly accepted a change.
func (s *server) expectedFingerprint(c credentials) string {
	if c.AcceptNewKey {
		return c.Fingerprint
	}
	if stored, ok := s.known.Lookup(c.Host); ok {
		return stored
	}
	return c.Fingerprint
}

type response struct {
	OK          bool   `json:"ok"`
	Message     string `json:"message,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Model       string `json:"model,omitempty"`
	OSVersion   string `json:"os_version,omitempty"`
	Compatible  bool   `json:"compatible,omitempty"`
	Installed   bool   `json:"installed,omitempty"`
	Active      bool   `json:"active,omitempty"`
	// KnownHost is true when this tablet's key matches the one recorded on a
	// previous run. Changed reports a key that no longer matches.
	KnownHost bool `json:"known_host,omitempty"`
	Changed   bool `json:"fingerprint_changed,omitempty"`
}

type payloadManifest struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		fatalDialog(err.Error())
		return
	}
	s := &server{
		csrf:       randomToken(),
		payloadDir: filepath.Join(filepath.Dir(exe), "payload"),
		known:      loadKnownHosts(defaultKnownHostsPath()),
	}
	listenAddress := os.Getenv("TRMNL_INSTALLER_ADDR")
	if listenAddress == "" {
		listenAddress = "127.0.0.1:0"
	}
	// The installer executes commands on the tablet over SSH. Binding anywhere
	// but loopback would expose that to the network, so an override that is not
	// loopback is refused rather than honored.
	if !isLoopbackHost(listenAddress) {
		fatalDialog("TRMNL_INSTALLER_ADDR must be a loopback address; refusing to listen on " + listenAddress)
		return
	}
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fatalDialog(err.Error())
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.index)
	mux.HandleFunc("/icon.png", s.icon)
	mux.HandleFunc("/api/preflight", s.preflight)
	mux.HandleFunc("/api/install", s.install)
	mux.HandleFunc("/api/reactivate", s.reactivate)
	mux.HandleFunc("/api/recover", s.recover)
	mux.HandleFunc("/api/uninstall", s.uninstall)
	mux.HandleFunc("/api/purge", s.purge)
	mux.HandleFunc("/api/quit", s.quit)
	httpServer := &http.Server{Handler: loopbackOnly(securityHeaders(mux)), ReadHeaderTimeout: 5 * time.Second}
	url := "http://" + ln.Addr().String() + "/"
	if os.Getenv("TRMNL_INSTALLER_NO_BROWSER") != "1" {
		go func() {
			time.Sleep(150 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				log.Printf("open browser: %v", err)
			}
		}()
	}
	if err := httpServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		fatalDialog(err.Error())
	}
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	b, err := webFiles.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "installer UI unavailable", http.StatusInternalServerError)
		return
	}
	html := strings.ReplaceAll(string(b), "{{CSRF}}", s.csrf)
	html = strings.ReplaceAll(html, "{{VERSION}}", version)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, html)
}

func (s *server) icon(w http.ResponseWriter, r *http.Request) {
	b, err := webFiles.ReadFile("web/icon.svg")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write(b)
}

func (s *server) preflight(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !s.decode(w, r, &c) {
		return
	}
	// Preflight deliberately accepts whatever key the tablet offers so the user
	// can see it. It is compared against the stored key immediately afterwards.
	client, fingerprint, err := connect(c, "")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: friendlySSHError(err)})
		return
	}
	defer client.Close()
	values, compatible, err := inspectDevice(client)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: "Connected, but device inspection failed: " + err.Error()})
		return
	}
	stored, seenBefore := s.known.Lookup(c.Host)
	changed := seenBefore && stored != fingerprint
	message := "Device found. Confirm the fingerprint, then click Install."
	switch {
	case changed:
		message = "This tablet's SSH key is not the one recorded last time. Reconnect over USB and confirm you are talking to your own tablet before continuing."
	case seenBefore:
		message = "Device found, and its SSH key matches the one recorded last time."
	}
	if !compatible {
		message = "This installer supports only reMarkable Paper Pro firmware 3.26 or 3.27. Nothing was changed."
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: message, Fingerprint: fingerprint, Model: values["model"], OSVersion: values["os"], Compatible: compatible, Installed: values["installed"] == "yes", Active: values["active"] == "yes", KnownHost: seenBefore && !changed, Changed: changed})
}

// install streams newline-delimited JSON progress events. An install uploads
// several megabytes and then runs a multi-minute device script, so a single
// blocking response would leave the page with no way to tell work from a hang.
func (s *server) install(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !s.decode(w, r, &c) {
		return
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	encoder := json.NewEncoder(w)
	step := 0
	emit := func(v progress) {
		_ = encoder.Encode(v)
		if flusher != nil {
			flusher.Flush()
		}
	}
	stage := func(message string) {
		step++
		emit(progress{Stage: message, Step: step, Total: installSteps})
	}
	fail := func(message string) { emit(progress{Error: message, Step: step, Total: installSteps}) }

	stage("Verifying the release payload")
	manifest, err := verifyPayload(s.payloadDir)
	if err != nil {
		fail("Release payload is incomplete: " + err.Error())
		return
	}

	stage("Connecting to the tablet")
	client, fingerprint, err := connect(c, s.expectedFingerprint(c))
	if err != nil {
		fail(friendlySSHError(err))
		return
	}
	defer client.Close()

	stage("Checking the model and firmware")
	if _, compatible, inspectErr := inspectDevice(client); inspectErr != nil || !compatible {
		fail("The model or firmware is not supported. Nothing was changed.")
		return
	}

	stage("Preparing the staging directory")
	if _, err := run(client, "rm -rf /tmp/trmnl-install && mkdir -p /tmp/trmnl-install/appload /tmp/trmnl-install/licenses"); err != nil {
		fail(err.Error())
		return
	}

	names := make([]string, 0, len(manifest.Files))
	for name := range manifest.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for i, name := range names {
		emit(progress{Stage: fmt.Sprintf("Uploading %s (%d of %d)", name, i+1, len(names)), Step: step, Total: installSteps})
		local := filepath.Join(s.payloadDir, filepath.FromSlash(name))
		remote := "/tmp/trmnl-install/" + filepath.ToSlash(name)
		if err := upload(client, local, remote); err != nil {
			fail("Upload failed: " + err.Error())
			return
		}
	}
	stage("Uploaded the payload")

	stage("Installing on the tablet")
	out, err := runWithin(client, "/bin/sh /tmp/trmnl-install/install-device-runtime.sh", installTimeout)
	if err != nil {
		fail("Installation failed and was rolled back. " + deviceFailure(out, err))
		return
	}

	// Only record the key once an install has actually succeeded with it.
	s.known.Remember(c.Host, fingerprint)
	message := "TRMNL is installed. On the tablet, open AppLoad and tap TRMNL."
	if strings.Contains(out, "TRMNL_ACTIVATION_PENDING") {
		// The files are installed and verified but the runtime did not start.
		message = "TRMNL is installed, but the extension runtime did not start. Click Reactivate after reboot, or restart the tablet and try again."
	}
	emit(progress{OK: true, Step: installSteps, Total: installSteps, Stage: message})
}

const installSteps = 6

// progress is one newline-delimited event from a long-running operation.
type progress struct {
	Stage string `json:"stage,omitempty"`
	Step  int    `json:"step"`
	Total int    `json:"total"`
	OK    bool   `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`
}

// reactivateCommand detaches the XOVI start script with setsid so restarting
// Xochitl cannot take the SSH session with it, then waits for the injection to
// appear. An earlier version backgrounded the job and returned immediately,
// which let the session teardown kill the child before it ran while still
// reporting success.
const reactivateCommand = `set -e
test -x /home/root/xovi/start
test -f /home/root/xovi/exthome/appload/trmnl-remarkable/manifest.json
rm -f /tmp/trmnl-xovi-start.log
setsid sh -c 'exec /home/root/xovi/start' >/tmp/trmnl-xovi-start.log 2>&1 </dev/null &
i=0
while [ "$i" -lt 30 ]; do
  sleep 1
  i=$((i + 1))
  pid=$(pidof xochitl 2>/dev/null | awk '{print $1}')
  if [ -n "$pid" ] && [ -r "/proc/$pid/environ" ] &&
     tr '\000' '\n' <"/proc/$pid/environ" | grep -q '^LD_PRELOAD=/home/root/xovi/xovi.so$'; then
    echo REACTIVATED
    exit 0
  fi
done
echo "XOVI did not start within 30 seconds." >&2
tail -n 5 /tmp/trmnl-xovi-start.log 2>/dev/null >&2
exit 1`

func (s *server) reactivate(w http.ResponseWriter, r *http.Request) {
	s.deviceAction(w, r, reactivateCommand, "TRMNL is active again. Open AppLoad on the tablet.")
}

func (s *server) recover(w http.ResponseWriter, r *http.Request) {
	s.deviceAction(w, r, "test -x /home/root/trmnl-remarkable/recover-stock.sh && /home/root/trmnl-remarkable/recover-stock.sh", "The normal reMarkable interface was restored. Developer Mode remains enabled.")
}

func (s *server) uninstall(w http.ResponseWriter, r *http.Request) {
	s.deviceAction(w, r, "test -x /home/root/trmnl-remarkable/uninstall.sh && /home/root/trmnl-remarkable/uninstall.sh", "TRMNL was removed. Settings and the shared extension runtime were preserved.")
}

func (s *server) purge(w http.ResponseWriter, r *http.Request) {
	s.deviceAction(w, r, "if test -x /home/root/trmnl-remarkable/uninstall.sh; then /home/root/trmnl-remarkable/uninstall.sh --purge; else rm -rf -- /home/root/.config/trmnl-remarkable /home/root/.cache/trmnl-remarkable /home/root/.local/share/trmnl-remarkable /home/root/xovi/exthome/appload/trmnl-remarkable; fi", "TRMNL and its settings, API key, cache, logs, history, and battery-test data were removed. The shared extension runtime was preserved.")
}

func (s *server) quit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !s.validCSRF(r) {
		writeJSON(w, http.StatusForbidden, response{Message: "Invalid installer session"})
		return
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: "Installer closed."})
	go func() {
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}()
}

func (s *server) deviceAction(w http.ResponseWriter, r *http.Request, command, success string) {
	var c credentials
	if !s.decode(w, r, &c) {
		return
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	client, fingerprint, err := connect(c, s.expectedFingerprint(c))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: friendlySSHError(err)})
		return
	}
	s.known.Remember(c.Host, fingerprint)
	defer client.Close()
	out, err := run(client, command)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: deviceFailure(out, err)})
		return
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: success})
}

// deviceFailure prefers the tablet's own output. Guard scripts exit silently
// when a required file is absent, so fall back to the transport error rather
// than returning an empty message the page cannot explain.
func deviceFailure(out string, err error) string {
	if trimmed := strings.TrimSpace(out); trimmed != "" {
		return trimmed
	}
	return "The tablet did not complete the request (" + err.Error() + "). Nothing was changed."
}

func (s *server) validCSRF(r *http.Request) bool {
	return subtle.ConstantTimeCompare([]byte(r.Header.Get("X-TRMNL-CSRF")), []byte(s.csrf)) == 1
}

func (s *server) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{Message: "POST required"})
		return false
	}
	if !s.validCSRF(r) {
		writeJSON(w, http.StatusForbidden, response{Message: "Invalid installer session"})
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Message: "Invalid request"})
		return false
	}
	return true
}

func connect(c credentials, expectedFingerprint string) (*ssh.Client, string, error) {
	host := strings.TrimSpace(c.Host)
	if host == "" {
		host = "10.11.99.1"
	}
	if !strings.Contains(host, ":") {
		host += ":22"
	}
	var fingerprint string
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password(c.Password)},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			fingerprint = ssh.FingerprintSHA256(key)
			if expectedFingerprint != "" && fingerprint != expectedFingerprint {
				return fmt.Errorf("device fingerprint changed: expected %s, received %s", expectedFingerprint, fingerprint)
			}
			return nil
		},
		Timeout: 12 * time.Second,
	}
	client, err := ssh.Dial("tcp", host, config)
	return client, fingerprint, err
}

// Timeouts bound every remote operation. golang.org/x/crypto/ssh has no
// context-aware session API, so a wedged tablet would otherwise block the
// installer and its operation lock indefinitely.
const (
	inspectTimeout = 30 * time.Second
	actionTimeout  = 3 * time.Minute
	uploadTimeout  = 5 * time.Minute
	installTimeout = 15 * time.Minute
)

func run(client *ssh.Client, command string) (string, error) {
	return runWithin(client, command, actionTimeout)
}

func runWithin(client *ssh.Client, command string, timeout time.Duration) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	type outcome struct {
		out []byte
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		b, err := session.CombinedOutput(command)
		done <- outcome{out: b, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case r := <-done:
		return string(r.out), r.err
	case <-timer.C:
		// Closing the session unblocks the reader goroutine.
		_ = session.Close()
		return "", fmt.Errorf("the tablet stopped responding after %s; nothing further was sent", timeout)
	}
}

func upload(client *ssh.Client, local, remote string) error {
	f, err := os.Open(local)
	if err != nil {
		return err
	}
	defer f.Close()
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	session.Stdin = f
	done := make(chan error, 1)
	go func() { done <- session.Run("umask 077; cat > " + shellQuote(remote)) }()
	timer := time.NewTimer(uploadTimeout)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		_ = session.Close()
		return fmt.Errorf("upload of %s stalled for %s", filepath.Base(local), uploadTimeout)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func verifyPayload(dir string) (payloadManifest, error) {
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return payloadManifest{}, err
	}
	var manifest payloadManifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		return manifest, err
	}
	if len(manifest.Files) == 0 {
		return manifest, errors.New("empty payload manifest")
	}
	if version != "dev" && manifest.Version != version {
		return manifest, fmt.Errorf("installer version %s does not match payload version %s", version, manifest.Version)
	}
	for name, expected := range manifest.Files {
		clean := filepath.Clean(filepath.FromSlash(name))
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			return manifest, fmt.Errorf("unsafe payload path %q", name)
		}
		f, err := os.Open(filepath.Join(dir, clean))
		if err != nil {
			return manifest, err
		}
		h := sha256.New()
		_, copyErr := io.Copy(h, f)
		closeErr := f.Close()
		if copyErr != nil {
			return manifest, copyErr
		}
		if closeErr != nil {
			return manifest, closeErr
		}
		actual := hex.EncodeToString(h.Sum(nil))
		if !strings.EqualFold(actual, expected) {
			return manifest, fmt.Errorf("checksum mismatch for %s", name)
		}
	}
	return manifest, nil
}

func parseKeyValues(s string) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(strings.ReplaceAll(s, "\r", ""), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			out[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return out
}

func inspectDevice(client *ssh.Client) (map[string]string, bool, error) {
	out, err := runWithin(client, inspectCommand, inspectTimeout)
	if err != nil {
		return nil, false, err
	}
	values := parseKeyValues(out)
	compatible := values["model"] == "reMarkable Ferrari" && values["arch"] == "aarch64" && (strings.HasPrefix(values["os"], "3.26.") || strings.HasPrefix(values["os"], "3.27."))
	return values, compatible, nil
}

const inspectCommand = `printf 'model='; tr -d '\000' </proc/device-tree/model 2>/dev/null || true; printf '\nos='; sed -n 's/^IMG_VERSION="\{0,1\}\([^" ]*\)"\{0,1\}$/\1/p' /etc/os-release | head -n1; printf '\narch='; uname -m; printf '\ninstalled='; test -f /home/root/xovi/exthome/appload/trmnl-remarkable/manifest.json && echo yes || echo no; printf 'active='; pid=$(pidof xochitl | awk '{print $1}'); if [ -n "$pid" ] && tr '\000' '\n' <"/proc/$pid/environ" 2>/dev/null | grep -q '^LD_PRELOAD=/home/root/xovi/xovi.so$'; then echo yes; else echo no; fi`

func writeJSON(w http.ResponseWriter, status int, v response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func randomToken() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func friendlySSHError(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "unable to authenticate"):
		return "The tablet rejected the password. Copy the SSH password from the tablet settings and try again."
	case strings.Contains(s, "connection refused"):
		return "The tablet is reachable, but SSH is disabled. Enable developer mode/SSH in tablet settings."
	case strings.Contains(s, "i/o timeout"), strings.Contains(s, "no route to host"):
		return "Could not reach the tablet. Connect its USB cable and leave the address set to 10.11.99.1."
	default:
		return s
	}
}

// loopbackOnly rejects requests whose Host header is not a loopback name. The
// listener already binds 127.0.0.1, but without this a DNS-rebinding page could
// reach the installer as a same-origin document and read the CSRF token.
func loopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "the installer answers loopback requests only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLoopbackHost(value string) bool {
	host := value
	if h, _, err := net.SplitHostPort(value); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func fatalDialog(message string) {
	if runtime.GOOS == "windows" {
		_ = exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", "Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show($args[0], 'TRMNL Installer')", message).Run()
		return
	}
	fmt.Fprintln(os.Stderr, message)
}
