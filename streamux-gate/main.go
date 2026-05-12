package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"

	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/accesscontrol"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	"github.com/creack/pty"
)

const (
	HOST   = "0.0.0.0"
	PORT   = 2222
	CTRL_B = byte(0x02)
	CTRL_C = byte(0x03)
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("Staring program")
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", HOST, PORT)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),

		wish.WithMiddleware(
			tmuxHandler,
			logging.Middleware(),
			accesscontrol.Middleware("tmux"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("starting SSH server on :2222")

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	log.Info("Server terminated")

}
func tmuxHandler(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		log.Debug("Started tmux handler")
		ptyReq, winCh, ok := sess.Pty()
		if !ok {
			sess.Exit(1)
			return
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

		//Create PTY for tmux
		log.Debug("Starting cmd")
		// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
		// and c.Stderr, calls c.Start, and returns the File of the tty's
		// corresponding pty.
		ptmx, err := pty.Start(cmd)
		if err != nil {
			fmt.Fprintf(sess, "failed to start tmux: %v\r\n", err)
			sess.Exit(1)
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
		log.Info("Before io.copy go routine")

		go func() {
			escape_sequence := []byte{CTRL_B, byte('d')}
			bPressed := false
			for {
				buffer := make([]byte, 1024)
				n, err := sess.Read(buffer)

				if err != nil {
					if !errors.Is(err, io.EOF) {
						log.Error("Failed to read from TTY")
						log.Error(err)
						return
					}
				}

				if bPressed && buffer[0] == byte('d') {
					log.Info("You hit ctrl+b d")
					buffer = escape_sequence
					ptmx.Write(buffer[:2])
					continue
				}

				if buffer[0] == CTRL_B {
					log.Info("You hit ctrl+b")
					bPressed = true
				} else {
					bPressed = false
				}

				if slices.Index(buffer[:n], CTRL_C) >= 0 {
					log.Info("You hit ctrl+c")
					buffer = escape_sequence
					ptmx.Write(buffer[:2])
					continue
				}

				// log.Infof("%q", buffer[:n])

				// ptmx.Write(buffer[:n])
			}
		}()

		// go io.Copy(ptmx, sess)
		io.Copy(sess, ptmx)

		log.Debug("Starting cmd.Wait()")
		_ = cmd.Wait()
		log.Info("After cmd.Wait()")
		next(sess)
	}
}
