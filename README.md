# ops-ctrl
WIP minimalistic init system

# Usage
- Run the daemon
```
go run daemon/main.go
```
- Run the cli tool
```
go run cli/main.go help
```

# Plans
- Get this program running as PID1 in dev environment
- Add other signals and poweroff
- Improve configuration possibilities
- Systemd service file converter
- Add tests