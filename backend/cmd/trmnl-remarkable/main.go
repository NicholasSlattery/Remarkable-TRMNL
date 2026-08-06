package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "golang.org/x/image/bmp"
	"trmnl-remarkable/backend/internal/batterytest"
	"trmnl-remarkable/backend/internal/brightness"
	"trmnl-remarkable/backend/internal/cache"
	"trmnl-remarkable/backend/internal/config"
	"trmnl-remarkable/backend/internal/dither"
	"trmnl-remarkable/backend/internal/power"
	"trmnl-remarkable/backend/internal/protocol"
	"trmnl-remarkable/backend/internal/trmnl"
	"trmnl-remarkable/backend/internal/update"
)

var version = "dev"

const (
	msgInitialize        uint32 = 1
	msgSaveConfig        uint32 = 2
	msgTestConnection    uint32 = 3
	msgRefreshCurrent    uint32 = 4
	msgNext              uint32 = 5
	msgSetBrightness     uint32 = 6
	msgClearCache        uint32 = 7
	msgDiagnostics       uint32 = 8
	msgResume            uint32 = 10
	msgResetSettings     uint32 = 11
	msgPrevious          uint32 = 12
	msgUpdatePreferences uint32 = 13
	msgBatteryStart      uint32 = 14
	msgBatteryStop       uint32 = 15
	msgBatteryReset      uint32 = 16
	msgState             uint32 = 101
	msgImage             uint32 = 102
	msgStatus            uint32 = 103
	msgError             uint32 = 104
	msgHistory           uint32 = 105
	msgTestResult        uint32 = 106
	msgDiagnosticsResult uint32 = 107
	msgBatteryTest       uint32 = 108
)

type historyEntry struct {
	At     time.Time `json:"at"`
	Action string    `json:"action"`
	OK     bool      `json:"ok"`
	Detail string    `json:"detail"`
}
type trigger struct {
	advance bool
	reason  string
}

type app struct {
	mu                               sync.RWMutex
	batteryMu                        sync.Mutex
	cfg                              config.Config
	configPath, dataDir, historyPath string
	batteryPath                      string
	conn                             *protocol.Connection
	client                           *trmnl.Client
	cache                            cache.Store
	light                            *brightness.Device
	wake                             *power.RTC
	history                          []historyEntry
	triggers                         chan trigger
	ctx                              context.Context
	cancel                           context.CancelFunc
	restoreOnce                      sync.Once
	guardDisarm                      string
	nextRefresh                      time.Time
	wakeAlarmError                   string
	batteryTest                      batterytest.State
	updates                          *update.Checker
	updateResult                     update.Result
}

func main() {
	if len(os.Args) == 3 && os.Args[1] == "--self-check" {
		if err := selfCheck(os.Args[2]); err != nil {
			log.Fatal(err)
		}
		fmt.Println("TRMNL AppLoad bundle is valid")
		return
	}
	if len(os.Args) < 2 {
		log.Fatal("AppLoad socket path is required")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := filepath.Join(home, ".local", "share", "trmnl-remarkable")
	configPath := filepath.Join(home, ".config", "trmnl-remarkable", "config.json")
	for _, p := range []string{dataDir, filepath.Dir(configPath), filepath.Join(home, ".cache", "trmnl-remarkable")} {
		if err := os.MkdirAll(p, 0700); err != nil {
			log.Fatal(err)
		}
	}
	logPath := filepath.Join(dataDir, "trmnl.log")
	rotateLog(logPath, 1<<20)
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err == nil {
		defer lf.Close()
		log.SetOutput(lf)
	}
	lock, err := acquireLock(filepath.Join(dataDir, "backend.lock"))
	if err != nil {
		log.Fatal(err)
	}
	defer lock()
	cleanupTemps(dataDir, filepath.Dir(configPath), filepath.Join(home, ".cache", "trmnl-remarkable"))
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("config load failed: %v", err)
		cfg = config.Defaults()
	}
	conn, err := protocol.Dial(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	a := &app{cfg: cfg, configPath: configPath, dataDir: dataDir, historyPath: filepath.Join(dataDir, "history.json"), batteryPath: filepath.Join(dataDir, "battery-test.json"), conn: conn, client: trmnl.New(), cache: cache.Store{Dir: filepath.Join(home, ".cache", "trmnl-remarkable")}, triggers: make(chan trigger, 4), ctx: ctx, cancel: cancel}
	a.client.Version = version
	a.client.Battery = readBatteryVoltage
	a.client.RSSI = readRSSI
	a.updates = update.New()
	a.loadHistory()
	if saved, e := batterytest.Load(a.batteryPath); e == nil {
		a.batteryTest = saved
	} else {
		log.Printf("battery test load failed: %v", e)
	}
	if light, e := brightness.Discover("/sys/class/backlight"); e == nil {
		a.light = light
		a.startBrightnessGuard()
		if !cfg.UseSystemBrightness {
			if err := a.light.SetPercent(cfg.BrightnessPercent); err != nil {
				log.Printf("saved brightness could not be applied: %v", err)
			}
		}
	} else {
		log.Printf("frontlight unavailable: %v", e)
	}
	if rtc, e := power.Discover("/sys/class/rtc"); e == nil {
		a.wake = rtc
	} else {
		log.Printf("scheduled wake unavailable: %v", e)
	}
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		select {
		case s := <-sigs:
			log.Printf("received %s", s)
		case <-ctx.Done():
		}
		a.cleanup()
		_ = conn.Close()
		cancel()
	}()
	go a.scheduler()
	go a.batterySampler()
	go a.updateWatcher()
	if e, ok := a.cache.Latest(); ok && cfg.StartWithCacheOffline {
		a.sendImage(e, true)
	}
	for {
		m, err := conn.Receive()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && ctx.Err() == nil {
				log.Printf("AppLoad receive ended: %v", err)
			}
			break
		}
		if m.Type == protocol.SystemTerminate {
			break
		}
		a.handle(m)
	}
	a.cleanup()
}

func (a *app) handle(m protocol.Message) {
	switch m.Type {
	case msgInitialize:
		a.sendState()
		a.sendHistory()
		a.sendBatteryTest()
		a.queue(trigger{advance: true, reason: "startup"})
	case msgSaveConfig:
		var next config.Config
		if err := json.Unmarshal([]byte(m.Contents), &next); err != nil {
			a.sendError("Invalid settings", err)
			return
		}
		a.mu.RLock()
		old := a.cfg
		a.mu.RUnlock()
		if next.APIKey == "" {
			next.APIKey = old.APIKey
		}
		if err := next.Normalize(); err != nil {
			a.sendError("Invalid settings", err)
			return
		}
		if err := config.Save(a.configPath, next); err != nil {
			a.sendError("Could not save settings", err)
			return
		}
		a.mu.Lock()
		a.cfg = next
		a.mu.Unlock()
		if a.light != nil {
			if next.UseSystemBrightness {
				_ = a.light.Restore()
			} else {
				_ = a.setBrightness(next.BrightnessPercent)
			}
		}
		if e, ok := a.cache.Latest(); ok {
			a.sendImage(e, true)
		}
		a.sendStatus("Settings saved")
		a.sendState()
		if next.UpdateCheck && !old.UpdateCheck {
			go a.checkForUpdate()
		}
	case msgTestConnection:
		go a.testConnection(m.Contents)
	case msgRefreshCurrent:
		a.queue(trigger{advance: false, reason: "manual refresh"})
	case msgNext:
		a.queue(trigger{advance: true, reason: "next screen"})
	case msgSetBrightness:
		var v struct {
			Percent int `json:"percent"`
		}
		if json.Unmarshal([]byte(m.Contents), &v) == nil {
			if err := a.setBrightness(v.Percent); err != nil {
				a.sendError("Brightness change failed", err)
			} else {
				a.sendState()
			}
		}
	case msgClearCache:
		_ = a.cache.Clear()
		a.sendStatus("Cache cleared")
	case msgDiagnostics:
		a.send(msgDiagnosticsResult, a.diagnostics())
	case msgResume:
		a.recordBatteryWake()
		a.queue(trigger{advance: false, reason: "resume"})
	case msgResetSettings:
		next := config.Defaults()
		if err := config.Save(a.configPath, next); err != nil {
			a.sendError("Reset failed", err)
			return
		}
		a.mu.Lock()
		a.cfg = next
		a.mu.Unlock()
		if a.light != nil {
			_ = a.light.Restore()
		}
		if e, ok := a.cache.Latest(); ok {
			a.sendImage(e, true)
		}
		a.sendState()
		a.sendStatus("Settings reset")
	case msgPrevious:
		a.showPrevious()
	case msgUpdatePreferences:
		var p struct {
			Invert              *bool `json:"invert"`
			UseSystemBrightness *bool `json:"use_system_brightness"`
		}
		if err := json.Unmarshal([]byte(m.Contents), &p); err != nil {
			a.sendError("Invalid preference", err)
			return
		}
		a.mu.Lock()
		if p.Invert != nil {
			a.cfg.Invert = *p.Invert
		}
		if p.UseSystemBrightness != nil {
			a.cfg.UseSystemBrightness = *p.UseSystemBrightness
		}
		cfg := a.cfg
		a.mu.Unlock()
		if err := config.Save(a.configPath, cfg); err != nil {
			a.sendError("Could not save preference", err)
			return
		}
		if p.UseSystemBrightness != nil && a.light != nil {
			if *p.UseSystemBrightness {
				_ = a.light.Restore()
			} else {
				_ = a.light.SetPercent(cfg.BrightnessPercent)
			}
		}
		if e, ok := a.cache.Latest(); ok {
			a.sendImage(e, true)
		}
		a.sendState()
	case msgBatteryStart:
		a.startBatteryTest()
	case msgBatteryStop:
		a.stopBatteryTest()
	case msgBatteryReset:
		a.resetBatteryTest()
	}
}

func (a *app) scheduler() {
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()
	backoff := time.Minute
	run := func(t trigger) {
		ok, refresh := a.fetch(t)
		if ok {
			backoff = time.Minute
		} else {
			if refresh <= 0 {
				refresh = backoff
			}
			if backoff < 30*time.Minute {
				backoff *= 2
			}
		}
		a.scheduleNext(timer, a.refreshInterval(refresh))
	}
	for {
		select {
		case t := <-a.triggers:
			run(t)
		case <-timer.C:
			run(trigger{advance: true, reason: "scheduled"})
		case <-a.ctx.Done():
			return
		}
	}
}

// refreshInterval applies the configured floor and, when the battery is low and
// the tablet is not charging, stretches the interval so a dashboard left
// unattended lasts materially longer.
func (a *app) refreshInterval(refresh time.Duration) time.Duration {
	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()
	if min := time.Duration(cfg.MinimumRefreshSeconds) * time.Second; refresh < min {
		refresh = min
	}
	saving, _ := a.batterySaving(cfg)
	if saving {
		refresh *= config.BatterySaverMultiplier
		if refresh > config.BatterySaverMaxRefresh {
			refresh = config.BatterySaverMaxRefresh
		}
	}
	return refresh
}

// batterySaving reports whether the low-battery interval is in force, plus the
// percentage that decided it.
func (a *app) batterySaving(cfg config.Config) (bool, int) {
	percent := readBatteryPercent()
	if cfg.BatterySaverPercent <= 0 || percent < 0 || percent > cfg.BatterySaverPercent {
		return false, percent
	}
	switch readBatteryStatus() {
	case "Charging", "Full":
		return false, percent
	}
	return true, percent
}

func (a *app) fetch(t trigger) (bool, time.Duration) {
	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()
	if cfg.APIKey == "" && cfg.DeviceID == "" {
		a.sendStatus("Open Settings to enter a TRMNL API key or BYOS device ID")
		return true, time.Hour
	}
	a.sendStatus("Refreshing…")
	r, err := a.client.Display(a.ctx, cfg, t.advance)
	if err != nil {
		a.record(t.reason, false, safeError(err))
		a.sendError("Refresh failed; cached screen remains visible", err)
		return false, retryDelay(err)
	}
	a.recordBatteryRefresh()
	refresh := time.Duration(int(r.RefreshRate)) * time.Second
	if refresh <= 0 {
		refresh = 15 * time.Minute
	}
	latest, has := a.cache.Latest()
	name := r.Filename
	if name == "" {
		name = r.ImageName
	}
	if name == "" {
		name = filepath.Base(strings.Split(r.ImageURL, "?")[0])
	}
	etag, lastModified := "", ""
	if has && r.ImageURL == latest.ImageURL {
		etag, lastModified = latest.ETag, latest.LastModified
	}
	timeout := time.Duration(int(r.ImageURLTimeout)) * time.Second
	body, h, err := a.client.Download(a.ctx, r.ImageURL, timeout, etag, lastModified)
	if errors.Is(err, trmnl.ErrNotModified) && has {
		a.sendImage(latest, false)
		a.sendState()
		a.record(t.reason, true, "unchanged; server confirmed cached image")
		a.sendStatus(fmt.Sprintf("Up to date · next refresh in %s", humanDuration(refresh)))
		return true, refresh
	}
	if err != nil {
		a.record(t.reason, false, safeError(err))
		a.sendError("Image download failed; cached screen remains visible", err)
		return false, retryDelay(err)
	}
	defer body.Close()
	const maxImageBytes = 25 << 20
	payload, err := io.ReadAll(io.LimitReader(body, maxImageBytes+1))
	if err != nil || len(payload) == 0 || len(payload) > maxImageBytes {
		if err == nil {
			err = fmt.Errorf("invalid image size: %d", len(payload))
		}
		a.record(t.reason, false, safeError(err))
		a.sendError("Image download was invalid; cached screen remains visible", err)
		return false, 0
	}
	if err := validateImagePayload(payload); err != nil {
		a.record(t.reason, false, safeError(err))
		a.sendError("Image download was invalid; cached screen remains visible", err)
		return false, 0
	}
	entry, err := a.cache.Put(name, bytes.NewReader(payload), cache.Entry{ImageURL: r.ImageURL, ImageName: r.ImageName, ETag: h.Get("ETag"), LastModified: h.Get("Last-Modified")}, maxImageBytes, 8)
	if err != nil {
		a.record(t.reason, false, safeError(err))
		a.sendError("Could not cache image", err)
		return false, 0
	}
	a.sendImage(entry, false)
	a.sendState()
	detail := "displayed " + filepath.Base(entry.Path)
	if r.SpecialFunction != "" {
		detail += "; special_function=" + r.SpecialFunction
	}
	if r.UpdateFirmware || r.ResetFirmware {
		detail += "; ignored firmware directive"
	}
	a.record(t.reason, true, detail)
	a.sendStatus(fmt.Sprintf("Updated · next refresh in %s", humanDuration(refresh)))
	return true, refresh
}

func (a *app) testConnection(contents string) {
	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()
	if strings.TrimSpace(contents) != "" {
		var candidate config.Config
		if json.Unmarshal([]byte(contents), &candidate) == nil {
			if candidate.APIKey == "" {
				candidate.APIKey = cfg.APIKey
			}
			if err := candidate.Normalize(); err != nil {
				a.send(msgTestResult, mustJSON(map[string]any{"ok": false, "message": safeError(err)}))
				return
			}
			cfg = candidate
		}
	}
	ctx, cancel := context.WithTimeout(a.ctx, 35*time.Second)
	defer cancel()
	_, err := a.client.Display(ctx, cfg, false)
	if err != nil {
		a.send(msgTestResult, mustJSON(map[string]any{"ok": false, "message": safeError(err)}))
		return
	}
	a.send(msgTestResult, mustJSON(map[string]any{"ok": true, "message": "Connection successful"}))
}

func (a *app) setBrightness(percent int) error {
	if a.light == nil {
		return errors.New("frontlight control is unavailable")
	}
	if err := a.light.SetPercent(percent); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.BrightnessPercent = percent
	cfg := a.cfg
	a.mu.Unlock()
	_ = config.Save(a.configPath, cfg)
	return nil
}

func (a *app) sendState() {
	a.mu.RLock()
	cfg := config.Redacted(a.cfg)
	hasKey := a.cfg.APIKey != ""
	nextRefresh := a.nextRefresh
	wakeAlarmError := a.wakeAlarmError
	a.mu.RUnlock()
	a.mu.RLock()
	full := a.cfg
	latest := a.updateResult
	a.mu.RUnlock()
	saving, percent := a.batterySaving(full)
	state := map[string]any{"config": cfg, "api_key_configured": hasKey, "version": version, "battery_percent": percent, "wake_for_refresh_available": a.wake != nil, "battery_saving": saving, "quiet_hours_active": full.InQuietHours(time.Now())}
	if latest.Latest != "" {
		state["update"] = latest
	}
	if detectedID := readDeviceID(); detectedID != "" {
		state["detected_device_id"] = detectedID
	}
	if !nextRefresh.IsZero() {
		state["next_refresh"] = nextRefresh
	}
	if wakeAlarmError != "" {
		state["wake_alarm_error"] = wakeAlarmError
	}
	if a.light != nil {
		if _, err := a.light.Read(); err == nil {
			state["brightness"] = a.light.Snapshot()
			state["brightness_percent"] = a.light.Percent()
		}
	}
	if e, ok := a.cache.Latest(); ok {
		state["last_refresh"] = e.SavedAt
	}
	a.send(msgState, mustJSON(state))
}

func (a *app) sendImage(e cache.Entry, cached bool) {
	path := e.Path
	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()
	if cfg.Invert || cfg.Dither == "auto" {
		if rendered, err := a.renderedView(e.Path, cfg); err == nil {
			path = rendered
		} else {
			log.Printf("render failed, showing the original screen: %v", err)
		}
	}
	a.send(msgImage, mustJSON(map[string]any{"path": "file://" + path, "cached": cached, "saved_at": e.SavedAt}))
}
func (a *app) sendStatus(s string) { a.send(msgStatus, mustJSON(map[string]string{"message": s})) }
func (a *app) sendError(prefix string, err error) {
	log.Printf("%s: %v", prefix, err)
	a.send(msgError, mustJSON(map[string]string{"message": prefix + ": " + safeError(err)}))
}
func (a *app) send(typ uint32, s string) {
	if err := a.conn.Send(typ, s); err != nil && a.ctx.Err() == nil {
		log.Printf("send failed: %v", err)
	}
}
func (a *app) queue(t trigger) {
	select {
	case a.triggers <- t:
	default:
		a.sendStatus("Refresh already queued")
	}
}

func (a *app) showPrevious() {
	es, err := a.cache.Entries()
	if err != nil || len(es) < 2 {
		a.sendStatus("No previous cached screen")
		return
	}
	a.sendImage(es[1], true)
	a.sendStatus("Showing previous cached screen")
}

func (a *app) record(action string, ok bool, detail string) {
	h := historyEntry{At: time.Now().UTC(), Action: action, OK: ok, Detail: detail}
	a.mu.Lock()
	a.history = append([]historyEntry{h}, a.history...)
	limit := a.cfg.HistoryLimit
	if len(a.history) > limit {
		a.history = a.history[:limit]
	}
	copyH := append([]historyEntry(nil), a.history...)
	a.mu.Unlock()
	atomicJSON(a.historyPath, copyH, 0600)
	a.sendHistory()
}
func (a *app) loadHistory() {
	b, err := os.ReadFile(a.historyPath)
	if err == nil {
		_ = json.Unmarshal(b, &a.history)
	}
}
func (a *app) sendHistory() {
	a.mu.RLock()
	// Preserve an empty JSON array. A nil slice serializes as null, which is not
	// a valid model payload for the QML history view on first launch.
	h := append([]historyEntry{}, a.history...)
	a.mu.RUnlock()
	a.send(msgHistory, mustJSON(h))
}

// updateWatcher polls the release feed only while the owner has opted in. The
// checker caches for a day, so the six-hour tick simply notices a setting change
// without waiting for a restart.
func (a *app) updateWatcher() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	a.checkForUpdate()
	for {
		select {
		case <-ticker.C:
			a.checkForUpdate()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *app) checkForUpdate() {
	a.mu.RLock()
	enabled := a.cfg.UpdateCheck
	a.mu.RUnlock()
	if !enabled || a.updates == nil {
		return
	}
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	result, err := a.updates.Check(ctx, version, time.Now().UTC())
	if err != nil {
		log.Printf("update check failed: %v", safeError(err))
		return
	}
	a.mu.Lock()
	a.updateResult = result
	a.mu.Unlock()
	a.sendState()
}

func (a *app) batterySampler() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.recordBatterySample(false)
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *app) startBatteryTest() {
	percent := readBatteryPercent()
	if percent < 0 {
		a.sendError("Battery test could not start", errors.New("battery percentage is unavailable"))
		return
	}
	a.batteryMu.Lock()
	err := a.batteryTest.Start(time.Now().UTC(), percent)
	state := a.batteryTest.Clone()
	a.batteryMu.Unlock()
	if err != nil {
		a.sendError("Battery test could not start", err)
		return
	}
	atomicJSON(a.batteryPath, state, 0600)
	a.sendBatteryTest()
	a.sendStatus("Battery test started · unplug the charger for a useful estimate")
}

func (a *app) stopBatteryTest() {
	a.batteryMu.Lock()
	if !a.batteryTest.Active {
		a.batteryMu.Unlock()
		return
	}
	a.batteryTest.StopWithStatus(time.Now().UTC(), readBatteryPercent(), readBatteryStatus())
	state := a.batteryTest.Clone()
	a.batteryMu.Unlock()
	atomicJSON(a.batteryPath, state, 0600)
	a.sendBatteryTest()
	a.sendStatus("Battery test stopped")
}

func (a *app) resetBatteryTest() {
	a.batteryMu.Lock()
	a.batteryTest.Reset()
	state := a.batteryTest.Clone()
	a.batteryMu.Unlock()
	atomicJSON(a.batteryPath, state, 0600)
	a.sendBatteryTest()
	a.sendStatus("Battery test data reset")
}

func (a *app) recordBatterySample(force bool) {
	percent := readBatteryPercent()
	a.batteryMu.Lock()
	if !a.batteryTest.Active {
		a.batteryMu.Unlock()
		return
	}
	a.batteryTest.AddSampleWithStatus(time.Now().UTC(), percent, force, readBatteryStatus())
	state := a.batteryTest.Clone()
	a.batteryMu.Unlock()
	atomicJSON(a.batteryPath, state, 0600)
	a.sendBatteryTest()
}

func (a *app) recordBatteryRefresh() {
	percent := readBatteryPercent()
	a.batteryMu.Lock()
	if !a.batteryTest.Active {
		a.batteryMu.Unlock()
		return
	}
	a.batteryTest.Refreshes++
	a.batteryTest.AddSampleWithStatus(time.Now().UTC(), percent, false, readBatteryStatus())
	state := a.batteryTest.Clone()
	a.batteryMu.Unlock()
	atomicJSON(a.batteryPath, state, 0600)
	a.sendBatteryTest()
}

func (a *app) recordBatteryWake() {
	percent := readBatteryPercent()
	a.batteryMu.Lock()
	if !a.batteryTest.Active {
		a.batteryMu.Unlock()
		return
	}
	a.batteryTest.Wakes++
	a.batteryTest.AddSampleWithStatus(time.Now().UTC(), percent, false, readBatteryStatus())
	state := a.batteryTest.Clone()
	a.batteryMu.Unlock()
	atomicJSON(a.batteryPath, state, 0600)
	a.sendBatteryTest()
}

func (a *app) sendBatteryTest() {
	a.batteryMu.Lock()
	snapshot := a.batteryTest.Snapshot(time.Now().UTC(), readBatteryStatus())
	a.batteryMu.Unlock()
	a.send(msgBatteryTest, mustJSON(snapshot))
}

func (a *app) cleanup() {
	a.restoreOnce.Do(func() {
		a.mu.RLock()
		restore := a.cfg.RestoreBrightnessOnExit
		a.mu.RUnlock()
		if restore && a.light != nil {
			if err := a.light.Restore(); err != nil {
				log.Printf("brightness restore failed: %v", err)
			}
		}
		if a.guardDisarm != "" {
			_ = os.WriteFile(a.guardDisarm, []byte("ok\n"), 0600)
		}
		if a.wake != nil {
			if err := a.wake.Clear(); err != nil {
				log.Printf("RTC wake alarm clear failed: %v", err)
			}
		}
		a.cancel()
	})
}

func (a *app) startBrightnessGuard() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	script := filepath.Join(filepath.Dir(filepath.Dir(exe)), "scripts", "brightness_guard.sh")
	if _, err := os.Stat(script); err != nil {
		return
	}
	disarm := filepath.Join(a.dataDir, fmt.Sprintf("guard-%d.disarm", os.Getpid()))
	_ = os.Remove(disarm)
	a.guardDisarm = disarm
	cmd := exec.Command("/bin/sh", script, strconv.Itoa(os.Getpid()), a.light.Path, strconv.Itoa(a.light.Original), disarm)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		log.Printf("brightness guard failed: %v", err)
		a.guardDisarm = ""
		return
	}
	_ = cmd.Process.Release()
}

func (a *app) diagnostics() string {
	a.mu.RLock()
	cfg := config.Redacted(a.cfg)
	a.mu.RUnlock()
	m := map[string]any{"version": version, "config": cfg, "cache_dir": a.cache.Dir, "data_dir": a.dataDir, "battery_percent": readBatteryPercent(), "display": readTrimmed("/sys/class/graphics/fb0/virtual_size"), "os": readTrimmed("/etc/os-release")}
	a.batteryMu.Lock()
	m["battery_test"] = a.batteryTest.Snapshot(time.Now().UTC(), readBatteryStatus())
	a.batteryMu.Unlock()
	if a.light != nil {
		m["brightness"] = a.light.Snapshot()
	}
	return mustJSON(m)
}

// renderedView applies the display transforms the panel benefits from. Both are
// optional; when neither changes the image the caller shows the cached file.
func (a *app) renderedView(source string, cfg config.Config) (string, error) {
	f, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer f.Close()
	decoded, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	var src image.Image = decoded
	suffix := ""
	if cfg.Dither == "auto" {
		palette, paletteErr := dither.NewPalette(cfg.Palette(), config.ParseHexColor)
		if paletteErr != nil {
			log.Printf("dither palette unusable: %v", paletteErr)
		} else if dither.Needed(src, palette) {
			src = dither.Apply(src, palette)
			suffix += "-d"
		}
	}
	if cfg.Invert {
		src = invertImage(src)
		suffix += "-i"
	}
	if suffix == "" {
		return source, nil
	}

	// Derive the render name from its source and the transforms applied. QML's
	// Image ignores a source assignment that does not change the URL, so a
	// single fixed filename would freeze the dashboard on the first render and
	// make Previous a no-op.
	final := filepath.Join(a.dataDir, renderName(source, suffix))
	tmp, err := os.CreateTemp(a.dataDir, ".invert-*.tmp")
	if err != nil {
		return "", err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := png.Encode(tmp, src); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(name, final); err != nil {
		return "", err
	}
	pruneInverted(a.dataDir, 3)
	return final, nil
}

func invertImage(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
			dst.SetNRGBA(x, y, color.NRGBA{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B, A: c.A})
		}
	}
	return dst
}

func renderName(source, suffix string) string {
	base := filepath.Base(source)
	return "inverted-" + strings.TrimSuffix(base, filepath.Ext(base)) + suffix + ".png"
}

// pruneInverted keeps the newest renders so that toggling between the current
// and previous screen does not re-encode every time, and drops the rest.
func pruneInverted(dir string, keep int) {
	// "inverted.png" is the pre-2.1 fixed name and is always removed.
	_ = os.Remove(filepath.Join(dir, "inverted.png"))
	matches, _ := filepath.Glob(filepath.Join(dir, "inverted-*.png"))
	if len(matches) <= keep {
		return
	}
	type rendered struct {
		path string
		at   time.Time
	}
	var all []rendered
	for _, p := range matches {
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		all = append(all, rendered{path: p, at: st.ModTime()})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].at.After(all[j].at) })
	for _, old := range all[min(keep, len(all)):] {
		_ = os.Remove(old.path)
	}
}

func validateImagePayload(payload []byte) error {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("decode image header: %w", err)
	}
	if format != "png" && format != "jpeg" && format != "bmp" {
		return fmt.Errorf("unsupported image format %q", format)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 10000 || cfg.Height > 10000 || int64(cfg.Width)*int64(cfg.Height) > 40_000_000 {
		return fmt.Errorf("unsafe image dimensions: %dx%d", cfg.Width, cfg.Height)
	}
	return nil
}

const powerSupplyRoot = "/sys/class/power_supply"

func readBatteryPercent() int { return readBatteryPercentAt(powerSupplyRoot) }
func readBatteryStatus() string {
	dir, _ := systemBattery(powerSupplyRoot)
	if dir == "" {
		return ""
	}
	return readTrimmed(filepath.Join(dir, "status"))
}
func readBatteryVoltage() string { return readBatteryVoltageAt(powerSupplyRoot) }

// systemBattery picks the supply that reports the tablet's own charge. The
// Paper Pro also exposes marker accessories as Wireless supplies and a charger
// IC as a second "Battery" that has no capacity file, so percentage, status and
// voltage must all come from the one supply that reports a usable capacity.
func systemBattery(root string) (string, int) {
	paths, _ := filepath.Glob(filepath.Join(root, "*", "capacity"))
	sort.Strings(paths)
	var fallbackDir string
	fallbackPercent := -1
	for _, p := range paths {
		dir := filepath.Dir(p)
		v, err := strconv.Atoi(readTrimmed(p))
		if err != nil || v < 0 || v > 100 {
			continue
		}
		if readTrimmed(filepath.Join(dir, "type")) == "Battery" {
			return dir, v
		}
		if fallbackDir == "" {
			fallbackDir, fallbackPercent = dir, v
		}
	}
	return fallbackDir, fallbackPercent
}

func readBatteryPercentAt(root string) int {
	_, percent := systemBattery(root)
	return percent
}

func readBatteryVoltageAt(root string) string {
	dir, _ := systemBattery(root)
	if dir == "" {
		return ""
	}
	microvolts, err := strconv.ParseInt(readTrimmed(filepath.Join(dir, "voltage_now")), 10, 64)
	if err != nil || microvolts <= 0 {
		return ""
	}
	return strconv.FormatFloat(float64(microvolts)/1_000_000, 'f', 3, 64)
}
func readDeviceID() string {
	return readDeviceIDAt("/sys/class/net")
}

func readDeviceIDAt(root string) string {
	paths, _ := filepath.Glob(filepath.Join(root, "*", "address"))
	sort.SliceStable(paths, func(i, j int) bool {
		return filepath.Base(filepath.Dir(paths[i])) == "wlan0" && filepath.Base(filepath.Dir(paths[j])) != "wlan0"
	})
	for _, p := range paths {
		name := filepath.Base(filepath.Dir(p))
		if name == "lo" {
			continue
		}
		value := strings.ToLower(readTrimmed(p))
		hardware, err := net.ParseMAC(value)
		if err == nil && len(hardware) == 6 && value != "00:00:00:00:00:00" {
			return value
		}
	}
	return ""
}

var (
	rssiMu        sync.Mutex
	rssiValue     string
	rssiFetchedAt time.Time
	rssiPattern   = regexp.MustCompile(`signal:\s*(-?\d+)\s*dBm`)
)

// readRSSI reports the Wi-Fi signal strength TRMNL records alongside battery
// voltage. This firmware leaves /proc/net/wireless and the sysfs wireless
// directory empty, so the value comes from iw. It is cached because a refresh
// interval is minutes long but state updates are frequent.
func readRSSI() string {
	rssiMu.Lock()
	defer rssiMu.Unlock()
	if time.Since(rssiFetchedAt) < time.Minute {
		return rssiValue
	}
	rssiFetchedAt = time.Now()
	rssiValue = ""
	if readTrimmed("/sys/class/net/wlan0/operstate") != "up" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/sbin/iw", "dev", "wlan0", "link").Output()
	if err != nil {
		return ""
	}
	if m := rssiPattern.FindSubmatch(out); len(m) == 2 {
		rssiValue = string(m[1])
	}
	return rssiValue
}
func readTrimmed(p string) string {
	b, e := os.ReadFile(p)
	if e != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// urlPattern matches the absolute URLs that net/http embeds in *url.Error.
var urlPattern = regexp.MustCompile(`https?://[^\s"']+`)

// safeError prepares an error for the tablet screen, the history list and the
// log. TRMNL image URLs carry signed query parameters, so only the origin is
// kept; anything naming the access token is dropped entirely.
func safeError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if i := strings.Index(strings.ToLower(s), "access-token"); i >= 0 {
		s = s[:i] + "credentials redacted"
	}
	return urlPattern.ReplaceAllStringFunc(s, func(match string) string {
		u, parseErr := url.Parse(match)
		if parseErr != nil || u.Host == "" {
			return "[redacted URL]"
		}
		if u.RawQuery == "" && u.User == nil && u.Fragment == "" {
			return match
		}
		return u.Scheme + "://" + u.Host + u.EscapedPath() + " [query redacted]"
	})
}
func retryDelay(err error) time.Duration {
	var httpErr *trmnl.HTTPError
	if errors.As(err, &httpErr) && httpErr.RetryAfter > 0 {
		if httpErr.RetryAfter > 24*time.Hour {
			return 24 * time.Hour
		}
		return httpErr.RetryAfter
	}
	return 0
}
func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}

func (a *app) scheduleNext(timer *time.Timer, d time.Duration) {
	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()

	next := time.Now().Add(d).Truncate(time.Second)
	// A refresh that would land inside the quiet window is deferred until it
	// closes, so the tablet stays asleep overnight instead of waking hourly.
	if resumed := cfg.NextActiveTime(next); resumed.After(next) {
		next = resumed.Truncate(time.Second)
		d = time.Until(next)
	}
	resetTimer(timer, d)
	a.mu.Lock()
	a.nextRefresh = next
	wakeEnabled := a.cfg.WakeForRefresh
	a.wakeAlarmError = ""
	a.mu.Unlock()

	if a.wake != nil {
		var err error
		if wakeEnabled {
			err = a.wake.Set(next)
		} else {
			err = a.wake.Clear()
		}
		if err != nil {
			log.Printf("RTC wake alarm update failed: %v", err)
			a.mu.Lock()
			a.wakeAlarmError = safeError(err)
			a.mu.Unlock()
		}
	} else if wakeEnabled {
		a.mu.Lock()
		a.wakeAlarmError = "This firmware does not expose a writable RTC wake alarm"
		a.mu.Unlock()
	}
	a.sendState()
}

func atomicJSON(path string, v any, mode os.FileMode) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	f, e := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if e != nil {
		return
	}
	n := f.Name()
	defer os.Remove(n)
	_ = f.Chmod(mode)
	_, _ = f.Write(append(b, '\n'))
	_ = f.Sync()
	_ = f.Close()
	_ = os.Rename(n, path)
}
func rotateLog(path string, max int64) {
	if st, e := os.Stat(path); e == nil && st.Size() > max {
		_ = os.Rename(path, path+".1")
	}
}

func cleanupTemps(dirs ...string) {
	// Inverted renders are regenerated on demand, so nothing carries over a
	// restart; dropping them reclaims space when inversion is turned off.
	patterns := []string{".download-*.tmp", ".index-*.tmp", ".config-*.tmp", ".invert-*.tmp", ".tmp-*", "inverted.png", "inverted-*.png"}
	for _, dir := range dirs {
		for _, pattern := range patterns {
			matches, _ := filepath.Glob(filepath.Join(dir, pattern))
			for _, path := range matches {
				_ = os.Remove(path)
			}
		}
	}
}

func acquireLock(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if errors.Is(err, os.ErrExist) {
		b, _ := os.ReadFile(path)
		pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		if pid > 0 {
			if p, e := os.FindProcess(pid); e == nil && p.Signal(syscall.Signal(0)) == nil {
				return nil, fmt.Errorf("TRMNL backend is already running as PID %d", pid)
			}
		}
		_ = os.Remove(path)
		f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	}
	if err != nil {
		return nil, err
	}
	_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
	_ = f.Close()
	return func() { _ = os.Remove(path) }, nil
}

func selfCheck(dir string) error {
	type manifest struct {
		Name                     string `json:"name"`
		ID                       string `json:"id"`
		Entry                    string `json:"entry"`
		LoadsBackend             bool   `json:"loadsBackend"`
		CanHaveMultipleFrontends bool   `json:"canHaveMultipleFrontends"`
	}
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("manifest JSON: %w", err)
	}
	if m.Name != "TRMNL" || m.ID != "trmnl.remarkable" || m.Entry != "/ui/TRMNL.qml" || !m.LoadsBackend || m.CanHaveMultipleFrontends {
		return fmt.Errorf("manifest fields are invalid")
	}
	rcc, err := os.ReadFile(filepath.Join(dir, "resources.rcc"))
	if err != nil {
		return fmt.Errorf("resources: %w", err)
	}
	if len(rcc) < 16 || string(rcc[:4]) != "qres" {
		return fmt.Errorf("resources.rcc is not a Qt binary resource")
	}
	iconFile, err := os.Open(filepath.Join(dir, "icon.png"))
	if err != nil {
		return fmt.Errorf("icon: %w", err)
	}
	defer iconFile.Close()
	cfg, format, err := image.DecodeConfig(iconFile)
	if err != nil || format != "png" || cfg.Width < 64 || cfg.Height < 64 {
		return fmt.Errorf("icon.png is invalid")
	}
	entry, err := os.Stat(filepath.Join(dir, "backend", "entry"))
	if err != nil {
		return fmt.Errorf("backend: %w", err)
	}
	if entry.Mode()&0111 == 0 {
		return fmt.Errorf("backend entry is not executable")
	}
	return nil
}
