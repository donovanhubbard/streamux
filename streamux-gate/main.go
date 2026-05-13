package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"time"

	"charm.land/wish/v2"
	"github.com/charmbracelet/ssh"
	"github.com/creack/pty"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	HOST   = "0.0.0.0"
	PORT   = 2222
	CTRL_B = byte(0x02)
	CTRL_C = byte(0x03)
)

func setupLogging() *lumberjack.Logger {
	logRotator := &lumberjack.Logger{
		Filename:   "./app.log",
		MaxSize:    50,   // Max size in MB
		MaxBackups: 3,    // Number of backups
		MaxAge:     30,   // Days
		Compress:   true, // Enable compression
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	multiWriter := io.MultiWriter(os.Stdout, logRotator)
	handler := slog.NewTextHandler(multiWriter, opts)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logRotator
}

func main() {
	logRotator := setupLogging()
	defer logRotator.Close()
	slog.Info("Starting program")

	address := fmt.Sprintf("%s:%d", HOST, PORT)
	s, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),

		// Middlewares are executed in reverse order they are added
		wish.WithMiddleware(
			tmuxHandler,
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
			"tmux",
			"-S",
			"/var/run/tmux/tmux.sock",
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
				_ = pty.Setsize(ptmx, &pty.Winsize{
					Rows: uint16(win.Height),
					Cols: uint16(win.Width),
				})
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

		io.Copy(sess, ptmx)

		_ = cmd.Wait()
		next(sess)
	}
}
