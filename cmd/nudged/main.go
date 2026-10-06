// Command nudged is the Nudge daemon: a long-running background process
// that speaks idle and dormant-project nudges through SpectreTTS. See
// nudge_service_v1.md for the full specification.
//
// Run with -setup to launch the interactive scan_roots configuration wizard.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Spectral-Kaburu/annoying-sister/internal/config"
	"github.com/Spectral-Kaburu/annoying-sister/internal/idle"
	"github.com/Spectral-Kaburu/annoying-sister/internal/logging"
	"github.com/Spectral-Kaburu/annoying-sister/internal/nudge"
	"github.com/Spectral-Kaburu/annoying-sister/internal/paths"
	"github.com/Spectral-Kaburu/annoying-sister/internal/projects"
	"github.com/Spectral-Kaburu/annoying-sister/internal/state"
	"github.com/Spectral-Kaburu/annoying-sister/internal/tts"
)

// eventChanBufferSize matches the spec's suggested buffered channel size.
const eventChanBufferSize = 8

func main() {
	setupMode := flag.Bool("setup", false, "launch the interactive scan_roots setup wizard")
	flag.Parse()

	if *setupMode {
		cfgPath, err := paths.ConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "nudged -setup: %v\n", err)
			os.Exit(1)
		}
		if err := paths.EnsureNudgeDir(); err != nil {
			fmt.Fprintf(os.Stderr, "nudged -setup: %v\n", err)
			os.Exit(1)
		}
		if err := runSetup(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "nudged -setup: %v\n", err)
			os.Exit(1)
		}
		return
	}

	logger := logging.Init()
	if err := run(logger); err != nil {
		logger.Error("fatal startup error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	mainLog := logging.For(logger, logging.ComponentMain)
	mainLog.Info("nudged starting", "pid", os.Getpid())

	if err := paths.EnsureNudgeDir(); err != nil {
		return err
	}

	cfgPath, err := paths.ConfigPath()
	if err != nil {
		return err
	}
	projPath, err := paths.ProjectsPath()
	if err != nil {
		return err
	}
	statePath, err := paths.StatePath()
	if err != nil {
		return err
	}

	cfg, err := loadConfig(logging.For(logger, logging.ComponentConfig), cfgPath)
	if err != nil {
		return err
	}

	projStore, err := loadProjects(logging.For(logger, logging.ComponentProjects), projPath)
	if err != nil {
		return err
	}

	stateStore, err := loadState(logging.For(logger, logging.ComponentState), statePath)
	if err != nil {
		return err
	}

	dbusReader, err := idle.NewDBusReader()
	if err != nil {
		mainLog.Error("failed to connect to session D-Bus for idle monitoring", "error", err)
		return err
	}
	defer func() {
		if err := dbusReader.Close(); err != nil {
			mainLog.Warn("error closing D-Bus connection", "error", err)
		}
	}()

	ttsClient := tts.NewClient(cfg.SpectreTTSSocketPath)
	mainLog.Info("configuration loaded",
		"scan_roots", cfg.ScanRoots,
		"idle_threshold_minutes", cfg.IdleThresholdMinutes,
		"dormant_threshold_days", cfg.DormantThresholdDays,
		"spectretts_socket_path", cfg.SpectreTTSSocketPath,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	events := make(chan nudge.Event, eventChanBufferSize)

	firstIdleDone := make(chan struct{})
	firstScanDone := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(4)

	go runIdleWatcher(ctx, &wg, logging.For(logger, logging.ComponentIdle), cfg, dbusReader, projStore, stateStore, events, newRand(1), firstIdleDone)
	go runProjectScanner(ctx, &wg, logging.For(logger, logging.ComponentScanner), cfg, projStore, stateStore, events, newRand(2), firstScanDone)
	go runOnStart(ctx, &wg, logging.For(logger, logging.ComponentOnStart), cfg, projStore, events, newRand(3), firstScanDone, firstIdleDone)
	go runConsumer(ctx, &wg, logging.For(logger, logging.ComponentConsumer), events, ttsClient)

	mainLog.Info("nudged started", "goroutines", 4, "event_buffer_size", eventChanBufferSize)
	wg.Wait()
	mainLog.Info("nudged shut down cleanly")
	return nil
}

func newRand(streamID int64) *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano() ^ streamID))
}

// loadConfig loads config.json and logs, distinctly, whether it already
// existed or had to be created with defaults on this run — that
// distinction matters operationally (a fresh default config on what the
// user thinks is an established install is worth noticing).
func loadConfig(logger *slog.Logger, path string) (config.Config, error) {
	_, statErr := os.Stat(path)
	existedBefore := statErr == nil

	cfg, err := config.Load(path)
	if err != nil {
		logger.Error("failed to load config.json", "path", path, "error", err)
		return config.Config{}, err
	}

	if existedBefore {
		logger.Debug("config.json loaded", "path", path)
	} else {
		logger.Info("config.json not found, created with defaults", "path", path)
	}
	return cfg, nil
}

func loadProjects(logger *slog.Logger, path string) (*projects.Store, error) {
	_, statErr := os.Stat(path)
	existedBefore := statErr == nil

	store, err := projects.Load(path)
	if err != nil {
		logger.Error("failed to load projects.json", "path", path, "error", err)
		return nil, err
	}

	if existedBefore {
		logger.Debug("projects.json loaded", "path", path, "project_count", len(store.Snapshot()))
	} else {
		logger.Info("projects.json not found, starting with empty registry", "path", path)
	}
	return store, nil
}

func loadState(logger *slog.Logger, path string) (*state.Store, error) {
	_, statErr := os.Stat(path)
	existedBefore := statErr == nil

	store, err := state.Load(path)
	if err != nil {
		logger.Error("failed to load state.json", "path", path, "error", err)
		return nil, err
	}

	if existedBefore {
		logger.Debug("state.json loaded", "path", path)
	} else {
		logger.Info("state.json not found, starting with empty cooldown state", "path", path)
	}
	return store, nil
}

// runIdleWatcher polls idle time every cfg.IdlePollIntervalSeconds, decides
// whether to fire an idle nudge via idle.Watcher, and closes firstDone
// after its first poll completes.
func runIdleWatcher(
	ctx context.Context,
	wg *sync.WaitGroup,
	logger *slog.Logger,
	cfg config.Config,
	reader idle.Reader,
	projStore *projects.Store,
	stateStore *state.Store,
	out chan<- nudge.Event,
	rng *rand.Rand,
	firstDone chan<- struct{},
) {
	defer wg.Done()

	threshold := time.Duration(cfg.IdleThresholdMinutes) * time.Minute
	repeat := time.Duration(cfg.IdleRepeatIntervalMinutes) * time.Minute
	pollInterval := time.Duration(cfg.IdlePollIntervalSeconds) * time.Second
	dormantThreshold := time.Duration(cfg.DormantThresholdDays) * 24 * time.Hour

	logger.Info("starting", "threshold_minutes", cfg.IdleThresholdMinutes,
		"repeat_interval_minutes", cfg.IdleRepeatIntervalMinutes,
		"poll_interval_seconds", cfg.IdlePollIntervalSeconds)

	watcher := idle.NewWatcher(threshold, repeat)
	var mediaReader idle.MediaPlayerReader
	if dbusRdr, ok := reader.(*idle.DBusReader); ok && dbusRdr != nil && dbusRdr.Conn() != nil {
		mediaReader = idle.NewDBusMediaPlayerReader(dbusRdr.Conn())
	}
	roaster := idle.NewRoaster(rng.Int63())
	engine := idle.NewNudgeEngine(watcher, mediaReader, roaster)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	first := true
	poll := func() {
		idleTime, err := reader.GetIdleTime()
		if err != nil {
			logger.Warn("failed to read idle time from D-Bus", "error", err)
		} else {
			logger.Debug("polled idle time", "idle_minutes", idleTime.Minutes())
			now := time.Now()
			lastNudged := stateStore.IdleLastNudged()

			var dormantProject string
			var dormantFor time.Duration
			if projStore != nil {
				reg := projStore.Snapshot()
				if name, p, ok := projects.MostDormant(reg); ok && projects.IsDormant(p, dormantThreshold, now) {
					dormantProject = name
					dormantFor = now.Sub(p.LastActive)
				}
			}

			if fired, roastMsg := engine.Poll(idleTime, lastNudged, now, dormantProject, dormantFor); fired {
				idleMinutes := int(idleTime.Minutes())
				var text string
				if roastMsg != "" {
					text = roastMsg
				} else {
					text = nudge.RenderIdle(rng, nudge.Slots{IdleMinutes: idleMinutes})
				}
				select {
				case out <- nudge.Event{Category: nudge.CategoryIdle, Text: text}:
					logger.Info("idle nudge fired", "idle_minutes", idleMinutes)
					if err := stateStore.SetIdleLastNudged(now); err != nil {
						logger.Error("failed to persist idle cooldown state", "error", err)
					}
				case <-ctx.Done():
					logger.Info("shutdown signal received mid-enqueue, exiting")
					return
				}
			}
		}
		if first {
			close(firstDone)
			first = false
			logger.Debug("first idle read complete")
		}
	}

	poll() // immediate first read, so firstDone can close promptly at boot
	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received, exiting")
			return
		case <-ticker.C:
			poll()
		}
	}
}

// runProjectScanner re-walks scan_roots every
// cfg.ProjectScanIntervalMinutes, merges discoveries into the registry,
// and fires dormant-project nudges under the once-per-calendar-day
// cooldown. Closes firstDone after its first scan completes.
func runProjectScanner(
	ctx context.Context,
	wg *sync.WaitGroup,
	logger *slog.Logger,
	cfg config.Config,
	projStore *projects.Store,
	stateStore *state.Store,
	out chan<- nudge.Event,
	rng *rand.Rand,
	firstDone chan<- struct{},
) {
	defer wg.Done()

	logger.Info("starting", "scan_roots", cfg.ScanRoots, "ignored_paths", cfg.IgnoredPaths,
		"scan_interval_minutes", cfg.ProjectScanIntervalMinutes,
		"dormant_threshold_days", cfg.DormantThresholdDays)

	ticker := time.NewTicker(time.Duration(cfg.ProjectScanIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	dormantThreshold := time.Duration(cfg.DormantThresholdDays) * 24 * time.Hour

	first := true
	scan := func() {
		start := time.Now()
		discovered, err := projects.Discover(cfg.ScanRoots, cfg.IgnoredPaths)
		if err != nil {
			logger.Error("discovery failed", "error", err)
			if first {
				close(firstDone)
				first = false
			}
			return
		}
		logger.Debug("discovery complete", "discovered_count", len(discovered))

		existing := projStore.Snapshot()
		lastActiveFailures := 0
		merged := projects.Merge(existing, discovered, func(path string) projects.LastActiveResult {
			t, err := projects.ComputeLastActive(path)
			if err != nil {
				lastActiveFailures++
				logger.Warn("failed to compute last_active", "path", path, "error", err)
				return projects.LastActiveResult{OK: false}
			}
			return projects.LastActiveResult{LastActive: t, OK: true}
		})

		if err := projStore.Replace(merged); err != nil {
			logger.Error("failed to persist projects.json", "error", err)
		}

		logger.Info("scan complete",
			"duration_ms", time.Since(start).Milliseconds(),
			"discovered_count", len(discovered),
			"registry_size", len(merged),
			"last_active_failures", lastActiveFailures,
		)

		now := time.Now()
		today := now.Format("2006-01-02")

		// Find the most dormant project that hasn't been nudged today.
		// Nudging at most one project per scan cycle spaces them out naturally.
		var targetName string
		var targetProject projects.Project
		foundDormant := false

		for name, p := range merged {
			if !projects.IsDormant(p, dormantThreshold, now) {
				continue
			}
			if lastDate, ok := stateStore.ProjectLastNudgedDate(name); ok && lastDate == today {
				logger.Debug("dormant project skipped, already nudged today", "project", name)
				continue
			}
			if !foundDormant || p.LastActive.Before(targetProject.LastActive) {
				targetName = name
				targetProject = p
				foundDormant = true
			}
		}

		if foundDormant {
			days := int(now.Sub(targetProject.LastActive).Hours() / 24)
			text := nudge.RenderDormant(rng, nudge.Slots{ProjectName: targetName, DormantDays: days})
			select {
			case out <- nudge.Event{Category: nudge.CategoryDormantProject, Text: text}:
				logger.Info("dormant project nudge fired", "project", targetName, "dormant_days", days)
				if err := stateStore.SetProjectLastNudgedDate(targetName, today); err != nil {
					logger.Error("failed to persist dormant cooldown state", "project", targetName, "error", err)
				}
			case <-ctx.Done():
				logger.Info("shutdown signal received mid-enqueue, exiting")
				return
			}
		}

		if first {
			close(firstDone)
			first = false
			logger.Debug("first project scan complete")
		}
	}

	scan()
	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received, exiting")
			return
		case <-ticker.C:
			scan()
		}
	}
}

// runOnStart waits for the first project scan and first idle read to both
// complete, then fires exactly one on-start NudgeEvent — announcing
// startup and, if a project is already dormant, mentioning the
// most-dormant one.
func runOnStart(
	ctx context.Context,
	wg *sync.WaitGroup,
	logger *slog.Logger,
	cfg config.Config,
	projStore *projects.Store,
	out chan<- nudge.Event,
	rng *rand.Rand,
	firstScanDone <-chan struct{},
	firstIdleDone <-chan struct{},
) {
	defer wg.Done()
	logger.Debug("waiting for first scan and first idle read")

	select {
	case <-firstScanDone:
	case <-ctx.Done():
		logger.Info("shutdown signal received while waiting, exiting without firing")
		return
	}
	select {
	case <-firstIdleDone:
	case <-ctx.Done():
		logger.Info("shutdown signal received while waiting, exiting without firing")
		return
	}

	reg := projStore.Snapshot()
	dormantThreshold := time.Duration(cfg.DormantThresholdDays) * 24 * time.Hour
	now := time.Now()

	name, p, ok := projects.MostDormant(reg)
	hasDormant := ok && projects.IsDormant(p, dormantThreshold, now)

	var slots nudge.Slots
	logAttrs := []any{"has_dormant_project", hasDormant}
	if hasDormant {
		slots.ProjectName = name
		slots.DormantDays = int(now.Sub(p.LastActive).Hours() / 24)
		logAttrs = append(logAttrs, "project", name, "dormant_days", slots.DormantDays)
	}

	text := nudge.RenderOnStart(rng, hasDormant, slots)
	select {
	case out <- nudge.Event{Category: nudge.CategoryOnStart, Text: text}:
		logger.Info("on-start nudge fired", logAttrs...)
	case <-ctx.Done():
		logger.Info("shutdown signal received mid-enqueue, exiting without firing")
	}
}

// runConsumer is the single goroutine that ever writes to the SpectreTTS
// socket, guaranteeing nudges are never interleaved on the wire. It checks
// ctx.Done() only between messages, never mid-write, per the spec's
// shutdown requirements.
func runConsumer(ctx context.Context, wg *sync.WaitGroup, logger *slog.Logger, in <-chan nudge.Event, client *tts.Client) {
	defer wg.Done()
	logger.Info("starting", "socket_path", client.SocketPath)

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received, exiting")
			return
		case ev, ok := <-in:
			if !ok {
				logger.Info("event channel closed, exiting")
				return
			}
			start := time.Now()
			if err := client.Speak(ev.Text); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					logger.Warn("spectretts socket not found, dropping nudge",
						"category", ev.Category, "error", err)
				} else {
					logger.Warn("failed to send nudge to spectretts",
						"category", ev.Category, "error", err)
				}
				continue
			}
			logger.Debug("nudge spoken",
				"category", ev.Category,
				"text_length", len(ev.Text),
				"duration_ms", time.Since(start).Milliseconds(),
			)

			// Breathing space: give SpectreTTS time to speak the text and pause before next message
			delay := estimateSpeakDelay(ev.Text)
			logger.Debug("pacing consumer queue", "delay_ms", delay.Milliseconds())
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				logger.Info("shutdown signal received during pacing delay, exiting")
				return
			}
		}
	}
}

// estimateSpeakDelay calculates a pacing duration based on sentence length
// plus breathing room so consecutive speeches in the queue do not speak over each other.
func estimateSpeakDelay(text string) time.Duration {
	words := len(strings.Fields(text))
	if words == 0 {
		return 3 * time.Second
	}
	// Average speech rate is ~130-150 words/min (~400ms per word) + 2.5s breathing pause
	dur := time.Duration(words)*400*time.Millisecond + 2500*time.Millisecond
	if dur < 3*time.Second {
		dur = 3 * time.Second
	}
	if dur > 15*time.Second {
		dur = 15 * time.Second
	}
	return dur
}
