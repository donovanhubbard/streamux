# Streamux-Gate

Features
* Readme
* New logo
* robust, configurable logging
* Configurable socket address
* launchd service information plist file and install instructions
Bugs
* the pane background color is too light?

```
sudo chown root:wheel /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
sudo chmod 644 /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
```

sudo launchctl bootstrap system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
sudo launchctl enable system/net.svenxix.streamux-gate.plist

// Start the service
sudo launchctl kickstart -k system/net.svenxix.streamux-gate

// To reload the config
sudo launchctl bootout system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist
sudo launchctl bootstrap system /Library/LaunchDaemons/net.svenxix.streamux-gate.plist

// To restart the service
sudo launchctl kickstart -k system/net.svenxix.streamux-gate

// To stop the service entirely
sudo launchctl kill SIGTERM system/net.svenxix.streamux-gate
