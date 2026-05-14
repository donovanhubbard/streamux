
build:
	go build -o streamux-gate

install-bin: build
	-rm /Users/tmuxview/streamux-gate
	cp streamux-gate /Users/tmuxview/
	chown tmuxview:staff /Users/tmuxview/streamux-gate
	chmod 744 /Users/tmuxview/streamux-gate

install-service:
	-rm /Library/LaunchDaemons/net.svenxix.streamux-gate
	cp net.svenxix.streamux-gate.plist /Library/LaunchDaemons/
	chown root:wheel /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
	chmod 644 /Library/LaunchDaemons/net.svenxix.streamux-gate.plist

start:
	sudo launchctl kickstart -k system/net.svenxix.streamux-gate

stop:
	sudo launchctl kill SIGTERM system/net.svenxix.streamux-gate

restart:
	sudo launchctl kickstart -k system/net.svenxix.streamux-gate

reload:
	sudo launchctl bootout system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
	sudo launchctl bootstrap system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist

