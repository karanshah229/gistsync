package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/natefinch/lumberjack"
)

var (
	// Log is the global logger instance
	Log *slog.Logger
	// currentWriter is the lumberjack writer, kept for closing if needed
	currentWriter *lumberjack.Logger
)

// Event types for WAL
const (
	TypeSyncStart        = "SYNC_START"
	TypeSyncSuccess      = "SYNC_SUCCESS"
	TypeSyncError        = "SYNC_ERROR"
	TypeCheckpoint       = "CHECKPOINT"
	TypeRecoveryComplete = "RECOVERY_COMPLETE"
)

// Init initializes the per-process logger
func Init(level string) error {
	logDir, err := storage.GetLogDir()
	if err != nil {
		return fmt.Errorf("failed to get log directory: %w", err)
	}

	// Create a unique filename for this process
	// Format: gistsync-YYYYMMDD-HHMMSS-PID.log
	now := time.Now()
	filename := fmt.Sprintf("gistsync-%s-%d.log", now.Format("20060102-150405"), os.Getpid())
	logPath := filepath.Join(logDir, filename)

	// Configure lumberjack for rotation
	currentWriter = &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,   // megabytes
		MaxBackups: 3,    // Not strictly needed with 30-day retention, but good for local protection
		MaxAge:     30,   // days (lumberjack's internal cleanup)
		Compress:   true, // compress rotated files
	}

	// Set log level
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create JSON handler for WAL capability
	handler := slog.NewJSONHandler(currentWriter, &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Ensure microsecond precision in timestamps
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339Nano)),
				}
			}
			return a
		},
	})

	Log = slog.New(handler).With(slog.Int("pid", os.Getpid()))
	slog.SetDefault(Log)

	return nil
}

// SyncStart logs the beginning of a sync operation
func SyncStart(localPath, remoteID string, isFolder bool) {
	Log.Info("Sync started",
		slog.String("type", TypeSyncStart),
		slog.String("local_path", localPath),
		slog.String("remote_id", remoteID),
		slog.Bool("is_folder", isFolder),
	)
}

// SyncSuccess logs a successful sync operation
func SyncSuccess(localPath, remoteID, hash string, isFolder bool, provider string, public bool) {
	Log.Info("Sync success",
		slog.String("type", TypeSyncSuccess),
		slog.String("local_path", localPath),
		slog.String("remote_id", remoteID),
		slog.String("hash", hash),
		slog.Bool("is_folder", isFolder),
		slog.String("provider", provider),
		slog.Bool("public", public),
	)
}

// SyncError logs a failed sync operation
func SyncError(localPath, err string) {
	Log.Error("Sync error",
		slog.String("type", TypeSyncError),
		slog.String("local_path", localPath),
		slog.String("error", err),
	)
}

// Checkpoint logs that a local commit has been successfully completed
func Checkpoint(msg string) {
	Log.Info(msg, slog.String("type", TypeCheckpoint))
}

// ReapLogs deletes log files older than 30 days
func ReapLogs() error {
	logDir, err := storage.GetLogDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	now := time.Now()
	retention := 30 * 24 * time.Hour

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		info, err := f.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > retention {
			_ = os.Remove(filepath.Join(logDir, f.Name()))
		}
	}

	return nil
}
