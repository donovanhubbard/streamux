package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"charm.land/wish/v2"
	"github.com/charmbracelet/ssh"
	"github.com/donovanhubbard/wishsplash"
	"github.com/spf13/viper"
	"github.com/superstarryeyes/bit/ansifonts"
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
			activeTerm,
			slogMiddleware,
		),
	)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Starting SSH listener", "address", address)

	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			slog.Error("Failed to start listener", "error", err)
			os.Exit(1)
		} else {
			slog.Debug("Exited cleanly from go routine")
		}
	}()

	<-done

	slog.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		slog.Error("could not stop server", "error", err)
	}

	slog.Info("Server terminated")
}
