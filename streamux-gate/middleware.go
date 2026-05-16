package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/creack/pty"
)

func activeTerm(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		_, _, containsPty := sess.Pty()
		if !containsPty {
			slog.Error(
				"A session did not have an acceptable tty",
				"source",
				sess.RemoteAddr().String(),
			)
			sess.Exit(1)
		} else {
			slog.Debug("Contains an active PTY session", "source", sess.RemoteAddr().String())
			next(sess)
		}
	}
}

func tmuxValidator(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		slog.Debug("Checking for socket.", "path", TMUX_SOCK_PATH, "source", sess.RemoteAddr().String())
		_, err := os.Stat(TMUX_SOCK_PATH)
		if errors.Is(err, os.ErrNotExist) {
			slog.Error("tmux socket is missing.", "path", TMUX_SOCK_PATH, "source", sess.RemoteAddr().String())
			sess.Write([]byte("We're sorry but the streaming hasn't started yet.\r\n"))
			sess.Exit(1)
		}

		slog.Debug("Socket found", "source", sess.RemoteAddr().String())
		cmdString := []string{TMUX_PATH, "-S", TMUX_SOCK_PATH, "list-clients"}
		slog.Debug("executing", "command", strings.Join(cmdString, " "), "source", sess.RemoteAddr().String())

		cmd := exec.Command(cmdString[0], cmdString[1:]...)
		out, err := cmd.Output()
		slog.Debug("Command ran.", "STDOUT", out, "source", sess.RemoteAddr().String())
		if err != nil {
			slog.Error("Error", "error", err, "source", sess.RemoteAddr().String())
			sess.Write([]byte("We're sorry but the streaming hasn't started yet.\r\n"))
			sess.Exit(1)
		}

		exitCode := cmd.ProcessState.ExitCode()
		slog.Debug("Exit code", "code", exitCode, "source", sess.RemoteAddr().String())

		if exitCode != 0 {
			slog.Error("Tmux server not running on socket", "path", TMUX_SOCK_PATH, "source", sess.RemoteAddr().String())
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
			"source", sess.RemoteAddr().String(),
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
			"source", sess.RemoteAddr().String(),
			"duration", time.Since(ct),
		)
	}
}

func tmuxHandler(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		ptyReq, winCh, _ := sess.Pty()

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
			slog.Error("Failed to start command on ssh session.", err, err, "source", sess.RemoteAddr().String())
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
						slog.Error("Failed to read from TTY", "error", err, "source", sess.RemoteAddr().String())
						return
					}
				}

				if bPressed && buffer[0] == byte('d') {
					slog.Info(
						"User ended session by pressing escape sequence",
						"source",
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
						"source",
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

func setWindowSize(win ssh.Window, ptmx *os.File) {
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(win.Height),
		Cols: uint16(win.Width),
	})
}
