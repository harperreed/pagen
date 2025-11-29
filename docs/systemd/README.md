# Systemd Service Installation (Linux)

This directory contains a systemd service template for running the pagen sync daemon as a background service on Linux systems.

## Prerequisites

1. Pagen installed and available on your PATH
2. Google OAuth credentials configured (run `pagen sync init` first)
3. Systemd-based Linux distribution (Ubuntu, Debian, Fedora, Arch, etc.)

## Installation Steps

### Option 1: User Service (Recommended)

Run the daemon as your user account (no root required):

```bash
# 1. Create user systemd directory
mkdir -p ~/.config/systemd/user

# 2. Copy service file
cp docs/systemd/pagen-sync.service ~/.config/systemd/user/

# 3. Edit the service file
nano ~/.config/systemd/user/pagen-sync.service

# Update these values:
#   - User: YOUR_USERNAME (your actual username)
#   - ExecStart: /path/to/pagen (find with: which pagen)
#   - Customize --interval and --services flags as needed

# 4. Create environment file for OAuth credentials
mkdir -p ~/.config/pagen
cat > ~/.config/pagen/sync-credentials.env << EOF
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
EOF

chmod 600 ~/.config/pagen/sync-credentials.env

# 5. Reload systemd and enable service
systemctl --user daemon-reload
systemctl --user enable pagen-sync.service
systemctl --user start pagen-sync.service

# 6. Check status
systemctl --user status pagen-sync.service
```

### Option 2: System Service

Run the daemon system-wide (requires root):

```bash
# 1. Copy service file to system directory
sudo cp docs/systemd/pagen-sync.service /etc/systemd/system/

# 2. Edit the service file
sudo nano /etc/systemd/system/pagen-sync.service

# Update these values:
#   - User: your-username
#   - ExecStart: /usr/local/bin/pagen (or wherever pagen is installed)
#   - Customize --interval and --services flags as needed

# 3. Create environment file
sudo mkdir -p /etc/pagen
sudo nano /etc/pagen/sync-credentials.env

# Add:
# GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
# GOOGLE_CLIENT_SECRET=your-client-secret

sudo chmod 600 /etc/pagen/sync-credentials.env

# Update EnvironmentFile path in service file if using /etc/pagen

# 4. Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable pagen-sync.service
sudo systemctl start pagen-sync.service

# 5. Check status
sudo systemctl status pagen-sync.service
```

## Managing the Service

### User Service Commands

```bash
# Start service
systemctl --user start pagen-sync.service

# Stop service
systemctl --user stop pagen-sync.service

# Restart service
systemctl --user restart pagen-sync.service

# View status
systemctl --user status pagen-sync.service

# View logs (live tail)
journalctl --user -u pagen-sync.service -f

# View logs (last 100 lines)
journalctl --user -u pagen-sync.service -n 100

# Disable service (stop running on boot)
systemctl --user disable pagen-sync.service

# Enable service (run on boot)
systemctl --user enable pagen-sync.service
```

### System Service Commands

Same as above but with `sudo` and without `--user` flag:

```bash
sudo systemctl start pagen-sync.service
sudo systemctl status pagen-sync.service
sudo journalctl -u pagen-sync.service -f
```

## Customization

### Sync Interval

Edit the `ExecStart` line in the service file:

```ini
# Sync every 15 minutes
ExecStart=/usr/local/bin/pagen sync daemon --interval 15m --services all

# Sync every 4 hours
ExecStart=/usr/local/bin/pagen sync daemon --interval 4h --services all

# Sync once per day
ExecStart=/usr/local/bin/pagen sync daemon --interval 24h --services all
```

**Note:** Minimum interval is 5 minutes to respect Google API rate limits.

### Services Selection

Sync only specific services:

```ini
# Only sync contacts and calendar (skip Gmail)
ExecStart=/usr/local/bin/pagen sync daemon --interval 1h --services contacts,calendar

# Only sync calendar
ExecStart=/usr/local/bin/pagen sync daemon --interval 30m --services calendar

# Sync all services (default)
ExecStart=/usr/local/bin/pagen sync daemon --interval 1h --services all
```

## Troubleshooting

### Service fails to start

Check logs for errors:
```bash
journalctl --user -u pagen-sync.service -n 50
```

Common issues:
- **"no authentication token found"**: Run `pagen sync init` to authenticate with Google
- **"invalid interval format"**: Check that your `--interval` value is valid (e.g., `15m`, `1h`, `4h`)
- **"permission denied"**: Ensure the service user has access to `~/.local/share/pagen/`

### OAuth token expired

The service will automatically handle token refresh. If it fails:

```bash
# Stop service
systemctl --user stop pagen-sync.service

# Re-authenticate
pagen sync init

# Start service
systemctl --user start pagen-sync.service
```

### High memory or CPU usage

Increase the sync interval to reduce API calls:

```ini
# Change from 15m to 1h
ExecStart=/usr/local/bin/pagen sync daemon --interval 1h --services all
```

Then reload and restart:
```bash
systemctl --user daemon-reload
systemctl --user restart pagen-sync.service
```

## Uninstallation

### User Service

```bash
# Stop and disable service
systemctl --user stop pagen-sync.service
systemctl --user disable pagen-sync.service

# Remove service file
rm ~/.config/systemd/user/pagen-sync.service

# Remove credentials (optional)
rm ~/.config/pagen/sync-credentials.env

# Reload systemd
systemctl --user daemon-reload
```

### System Service

```bash
# Stop and disable service
sudo systemctl stop pagen-sync.service
sudo systemctl disable pagen-sync.service

# Remove service file
sudo rm /etc/systemd/system/pagen-sync.service

# Remove credentials (optional)
sudo rm /etc/pagen/sync-credentials.env

# Reload systemd
sudo systemctl daemon-reload
```

## Security Notes

1. **Protect credentials file**: Always set permissions to `600` on the credentials file
2. **Use user service when possible**: Runs with your user permissions, not root
3. **Environment file recommended**: More secure than hardcoding credentials in service file
4. **ReadWritePaths restriction**: Service can only write to pagen data directory
5. **NoNewPrivileges**: Prevents privilege escalation

## Monitoring

### Set up alerts for failures

Create a systemd unit override to send email on failure:

```bash
# Create override directory
mkdir -p ~/.config/systemd/user/pagen-sync.service.d

# Create override file
cat > ~/.config/systemd/user/pagen-sync.service.d/email-on-failure.conf << EOF
[Unit]
OnFailure=status-email@%n.service
EOF

systemctl --user daemon-reload
```

### Monitor sync success

Check recent logs:
```bash
journalctl --user -u pagen-sync.service --since "1 hour ago" | grep "completed"
```

Look for:
- `✓ contacts sync completed`
- `✓ calendar sync completed`
- `✓ gmail sync completed`
