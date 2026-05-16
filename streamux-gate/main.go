package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/wish/v2"
	"github.com/charmbracelet/ssh"
	"github.com/creack/pty"
	"github.com/donovanhubbard/wishsplash"
	"github.com/spf13/viper"
	"github.com/superstarryeyes/bit/ansifonts"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	HOST           = "0.0.0.0"
	CTRL_B         = byte(0x02)
	CTRL_C         = byte(0x03)
	TMUX_SOCK_PATH = "/var/run/tmux/tmux.sock"
	TMUX_PATH      = "/opt/homebrew/bin/tmux"
	DISCORD_LINK   = "https://discord.gg/8DNNcujNBF"
)

type Config struct {
	Stdout   bool   `mapstructure:"stdout"`
	LogFile  string `mapstructure:"logFile"`
	LogLevel string `mapstructure:"logLevel"`
	Port     int    `mapstructure:"port"`
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		fmt.Errorf("invalid log level: %s", level)
		os.Exit(1)
	}
	return slog.LevelDebug
}

func setupOutput(cfg *Config) *lumberjack.Logger {
	logRotator := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    50,   // Max size in MB
		MaxBackups: 3,    // Number of backups
		MaxAge:     30,   // Days
		Compress:   true, // Enable compression
	}

	logLevel := parseLogLevel(cfg.LogLevel)

	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Check if the current attribute is the 'time' key
			if a.Key == slog.TimeKey {
				// Get the time value and format it using a 12-hour layout
				// "03:04:05PM" is the Go reference for 12-hour format with AM/PM
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format("2006-01-02 03:04:05PM -7:00"))
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.Stdout && cfg.LogFile != "" {
		multiWriter := io.MultiWriter(os.Stdout, logRotator)
		handler = slog.NewTextHandler(multiWriter, opts)
	} else if cfg.LogFile != "" {
		handler = slog.NewTextHandler(logRotator, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logRotator
}

func setupConfig(filePath string) *Config {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")
	v.SetDefault("Stdout", true)
	v.SetDefault("LogLevel", "info")
	err := v.ReadInConfig()

	if err != nil {
		fmt.Println(fmt.Errorf("failed to read config: %w", err))
		os.Exit(1)
	}

	if v.IsSet("port") == false {
		fmt.Println("Missing mandatory config 'port'")
		os.Exit(1)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		fmt.Println(fmt.Errorf("failed to unmarshal config: %w", err))
		os.Exit(1)
	}

	return &cfg
}

func printUsage() {
	path, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	programName := filepath.Base(path)

	fmt.Printf("usage: %s -c filePath\n", programName)
	fmt.Println("Runs a custom SSH server designed to serve tmux clients to the internet.")
	fmt.Println("  -c filePath    The path to the yaml config file.")
}

func main() {
	if len(os.Args) != 3 {
		printUsage()
		os.Exit(1)
	}

	configFilePath := os.Args[2]
	cfg := setupConfig(configFilePath)

	logRotator := setupOutput(cfg)
	defer logRotator.Close()
	slog.Info("Starting program")

	opts := wishsplash.Options{
		Font:  "8bitfortress",
		Text:  "STREAMUX",
		Delay: 3,
		RenderOptions: ansifonts.RenderOptions{
			CharSpacing:            1,
			WordSpacing:            3,
			LineSpacing:            1,
			TextColor:              "#FF0000",
			GradientColor:          "#00FF00",
			UseGradient:            true,
			GradientDirection:      ansifonts.LeftRight,
			Alignment:              ansifonts.CenterAlign,
			ScaleFactor:            2.0,
			ShadowEnabled:          false,
			ShadowHorizontalOffset: 2,
			ShadowVerticalOffset:   1,
			ShadowStyle:            ansifonts.MediumShade,
		},
	}
	// charmLogger := charmLog.New(os.Stderr)
	// charmLogger.SetLevel(charmLog.DebugLevel)

	address := fmt.Sprintf("%s:%d", HOST, cfg.Port)
	s, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),

		// Middlewares are executed in reverse order they are added
		wish.WithMiddleware(
			tmuxHandler,
			tmuxValidator,
			//wishsplash.WithLogger(opts, charmLogger),
			wishsplash.WithOptions(opts),
			slogMiddleware,
		),
	)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting SSH listener", "address", address)

	if err := s.ListenAndServe(); err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}
	slog.Info("Server terminated")

}

func tmuxValidator(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		slog.Debug("Checking for socket.", "path", TMUX_SOCK_PATH)
		_, err := os.Stat(TMUX_SOCK_PATH)
		if errors.Is(err, os.ErrNotExist) {
			slog.Error("tmux socket is missing.", "path", TMUX_SOCK_PATH)
			sess.Write([]byte("We're sorry but the streaming hasn't started yet.\r\n"))
			sess.Exit(1)
		}

		slog.Debug("Socket found")
		cmdString := []string{TMUX_PATH, "-S", TMUX_SOCK_PATH, "list-clients"}
		slog.Debug("executing", "command", strings.Join(cmdString, " "))

		cmd := exec.Command(cmdString[0], cmdString[1:]...)
		out, err := cmd.Output()
		slog.Debug("Command ran.", "STDOUT", out)
		if err != nil {
			slog.Error("Error", "error", err)
			sess.Write([]byte("We're sorry but the streaming hasn't started yet.\r\n"))
			sess.Exit(1)
		}

		exitCode := cmd.ProcessState.ExitCode()
		slog.Debug("Exit code", "code", exitCode)

		if exitCode != 0 {
			slog.Error("Tmux server not running on socket", "path", TMUX_SOCK_PATH)
			sess.Write([]byte("We're sorry but the streaming hasn't started yet.\r\n"))
			sess.Exit(1)
		}
		next(sess)
	}
}

func slogMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		ct := time.Now()
		hpk := sess.PublicKey() != nil
		pty, _, _ := sess.Pty()
		slog.Info(
			"User connected",
			"user", sess.User(),
			"remote-addr", sess.RemoteAddr().String(),
			"public-key", hpk,
			"command", sess.Command(),
			"term", pty.Term,
			"width", pty.Window.Width,
			"height", pty.Window.Height,
			"client-version", sess.Context().ClientVersion(),
		)

		next(sess)
		slog.Info(
			"User disconnected",
			"user", sess.User(),
			"remote-addr", sess.RemoteAddr().String(),
			"duration", time.Since(ct),
		)
	}
}

func setWindowSize(win ssh.Window, ptmx *os.File) {
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(win.Height),
		Cols: uint16(win.Width),
	})
}

func tmuxHandler(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		ptyReq, winCh, ok := sess.Pty()
		if !ok {
			slog.Error(
				"A session did not have an acceptable tty",
				"source address",
				sess.RemoteAddr().String(),
			)
			sess.Exit(1)
		}

		// Start or attach to a tmux session
		cmd := exec.CommandContext(
			context.Background(),
			TMUX_PATH,
			"-S",
			TMUX_SOCK_PATH,
			"attach-session",
			"-r",
		)

		// TERM handling matters for tmux
		cmd.Env = append(cmd.Env,
			"TERM="+ptyReq.Term,
			"LANG=en_US.UTF-8",
		)

		// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
		// and c.Stderr, calls c.Start, and returns the File of the tty's
		// corresponding pty.
		ptmx, err := pty.Start(cmd)
		if err != nil {
			slog.Error("Failed to start command on ssh session.", err, err)
			return
		}
		defer ptmx.Close()

		// Handle window resize
		go func() {
			for win := range winCh {
				setWindowSize(win, ptmx)
			}
		}()

		go func() {
			escape_sequence := []byte{CTRL_B, byte('d')}
			bPressed := false
			for {
				buffer := make([]byte, 1024)
				n, err := sess.Read(buffer)

				if err != nil {
					if !errors.Is(err, io.EOF) {
						slog.Error("Failed to read from TTY", "error", err)
						return
					}
				}

				if bPressed && buffer[0] == byte('d') {
					slog.Info(
						"User ended session by pressing escape sequence",
						"address",
						sess.RemoteAddr().String())
					buffer = escape_sequence
					ptmx.Write(buffer[:2])
					continue
				}

				if buffer[0] == CTRL_B {
					bPressed = true
				} else {
					bPressed = false
				}

				if slices.Index(buffer[:n], CTRL_C) >= 0 {
					slog.Info(
						"User ended session by pressing escape sequence",
						"address",
						sess.RemoteAddr().String())
					buffer = escape_sequence
					ptmx.Write(buffer[:2])
					continue
				}
			}
		}()

		// Set initial window size
		setWindowSize(ptyReq.Window, ptmx)

		io.Copy(sess, ptmx)

		_ = cmd.Wait()

		sess.Write([]byte(DISCORD_LINK + "\r\n"))
		next(sess)
	}
}
