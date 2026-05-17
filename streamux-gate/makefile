
run:
	go run *.go -c ./streamux-gate.yaml

build:
	go build -o streamux-gate

install-bin: build
	sudo rm /Users/tmuxview/streamux-gate
	sudo cp streamux-gate /Users/tmuxview/
	sudo chown tmuxview:staff /Users/tmuxview/streamux-gate
	sudo chmod 744 /Users/tmuxview/streamux-gate

install-service:
	sudo rm /Library/LaunchDaemons/net.svenxix.streamux-gate
	sudo cp net.svenxix.streamux-gate.plist /Library/LaunchDaemons/
	sudo chown root:wheel /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
	sudo chmod 644 /Library/LaunchDaemons/net.svenxix.streamux-gate.plist

start:
	sudo launchctl kickstart -k system/net.svenxix.streamux-gate

stop:
	sudo launchctl kill SIGTERM system/net.svenxix.streamux-gate

restart:
	sudo launchctl kickstart -k system/net.svenxix.streamux-gate

reload:
	sudo launchctl bootout system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
	sudo launchctl bootstrap system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist

