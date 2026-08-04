package main

import (
	"crypto/rand"
	"crypto/sha256"
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
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var version = "dev"

//go:embed web/index.html web/icon.svg
var webFiles embed.FS

type server struct {
	csrf       string
	payloadDir string
}

type credentials struct {
	Host        string `json:"host"`
	Password    string `json:"password"`
	Fingerprint string `json:"fingerprint"`
}

type response struct {
	OK          bool   `json:"ok"`
	Message     string `json:"message,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Model       string `json:"model,omitempty"`
	OSVersion   string `json:"os_version,omitempty"`
	Compatible  bool   `json:"compatible,omitempty"`
	Installed   bool   `json:"installed,omitempty"`
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
	s := &server{csrf: randomToken(), payloadDir: filepath.Join(filepath.Dir(exe), "payload")}
	listenAddress := os.Getenv("TRMNL_INSTALLER_ADDR")
	if listenAddress == "" {
		listenAddress = "127.0.0.1:0"
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
	mux.HandleFunc("/api/recover", s.recover)
	mux.HandleFunc("/api/uninstall", s.uninstall)
	mux.HandleFunc("/api/quit", s.quit)
	httpServer := &http.Server{Handler: securityHeaders(mux), ReadHeaderTimeout: 5 * time.Second}
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

func (s *server) icon(w http.ResponseWriter, _ *http.Request) {
	b, err := webFiles.ReadFile("web/icon.svg")
	if err != nil {
		http.NotFound(w, nil)
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
	message := "Device found. Confirm the fingerprint, then click Install."
	if !compatible {
		message = "This installer supports only reMarkable Paper Pro firmware 3.26 or 3.27. Nothing was changed."
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: message, Fingerprint: fingerprint, Model: values["model"], OSVersion: values["os"], Compatible: compatible, Installed: values["installed"] == "yes"})
}

func (s *server) install(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !s.decode(w, r, &c) {
		return
	}
	manifest, err := verifyPayload(s.payloadDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Message: "Release payload is incomplete: " + err.Error()})
		return
	}
	client, _, err := connect(c, c.Fingerprint)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: friendlySSHError(err)})
		return
	}
	defer client.Close()
	if _, compatible, inspectErr := inspectDevice(client); inspectErr != nil || !compatible {
		writeJSON(w, http.StatusBadRequest, response{Message: "The model or firmware is not supported. Nothing was changed."})
		return
	}
	if _, err := run(client, "rm -rf /tmp/trmnl-install && mkdir -p /tmp/trmnl-install/appload /tmp/trmnl-install/licenses"); err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: err.Error()})
		return
	}
	for name := range manifest.Files {
		local := filepath.Join(s.payloadDir, filepath.FromSlash(name))
		remote := "/tmp/trmnl-install/" + filepath.ToSlash(name)
		if err := upload(client, local, remote); err != nil {
			writeJSON(w, http.StatusBadGateway, response{Message: "Upload failed: " + err.Error()})
			return
		}
	}
	out, err := run(client, "/bin/sh /tmp/trmnl-install/install-device-runtime.sh")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: "Installation failed and was rolled back. " + strings.TrimSpace(out)})
		return
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: "TRMNL is installed. On the tablet, open AppLoad and tap TRMNL."})
}

func (s *server) recover(w http.ResponseWriter, r *http.Request) {
	s.deviceAction(w, r, "/home/root/trmnl-remarkable/recover-stock.sh", "Stock reMarkable interface restored.")
}

func (s *server) uninstall(w http.ResponseWriter, r *http.Request) {
	s.deviceAction(w, r, "/home/root/trmnl-remarkable/uninstall.sh", "TRMNL was removed. Settings and the shared extension runtime were preserved.")
}

func (s *server) quit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("X-TRMNL-CSRF") != s.csrf {
		writeJSON(w, http.StatusForbidden, response{Message: "Invalid installer session"})
		return
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: "Installer closed."})
	go func() {
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}()
}

func (s *server) deviceAction(w http.ResponseWriter, r *http.Request, script, success string) {
	var c credentials
	if !s.decode(w, r, &c) {
		return
	}
	client, _, err := connect(c, c.Fingerprint)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: friendlySSHError(err)})
		return
	}
	defer client.Close()
	out, err := run(client, "test -x "+script+" && "+script)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, response{Message: strings.TrimSpace(out)})
		return
	}
	writeJSON(w, http.StatusOK, response{OK: true, Message: success})
}

func (s *server) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{Message: "POST required"})
		return false
	}
	if r.Header.Get("X-TRMNL-CSRF") != s.csrf {
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

func run(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	b, err := session.CombinedOutput(command)
	return string(b), err
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
	return session.Run("umask 077; cat > " + shellQuote(remote))
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
	out, err := run(client, `printf 'model='; tr -d '\000' </proc/device-tree/model 2>/dev/null || true; printf '\nos='; sed -n 's/^IMG_VERSION="\{0,1\}\([^" ]*\)"\{0,1\}$/\1/p' /etc/os-release | head -n1; printf '\narch='; uname -m; printf '\ninstalled='; test -f /home/root/xovi/exthome/appload/trmnl-remarkable/manifest.json && echo yes || echo no`)
	if err != nil {
		return nil, false, err
	}
	values := parseKeyValues(out)
	compatible := values["model"] == "reMarkable Ferrari" && values["arch"] == "aarch64" && (strings.HasPrefix(values["os"], "3.26.") || strings.HasPrefix(values["os"], "3.27."))
	return values, compatible, nil
}

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
