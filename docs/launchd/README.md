# Launchd Service Installation (macOS)

This directory contains a launchd plist template for running the pagen sync daemon as a background service on macOS.

## Prerequisites

1. Pagen installed and available on your PATH
2. Google OAuth credentials configured (run `pagen sync init` first)
3. macOS 10.10 or later

## Installation Steps

### Step 1: Copy and Configure Service File

```bash
# 1. Create LaunchAgents directory if it doesn't exist
mkdir -p ~/Library/LaunchAgents

# 2. Copy the plist file
cp docs/launchd/com.pagen.sync.plist ~/Library/LaunchAgents/

# 3. Edit the plist file
nano ~/Library/LaunchAgents/com.pagen.sync.plist
```

### Step 2: Customize the Configuration

Edit these values in the plist file:

1. **Replace YOUR_USERNAME** with your actual macOS username:
   - `WorkingDirectory`: `/Users/YOUR_USERNAME`
   - `StandardOutPath`: `/Users/YOUR_USERNAME/Library/Logs/pagen-sync.log`
   - `StandardErrorPath`: `/Users/YOUR_USERNAME/Library/Logs/pagen-sync-error.log`

2. **Update pagen binary path** if needed:
   - Find your pagen location: `which pagen`
   - Update `ProgramArguments` first element to the actual path

3. **Add your Google OAuth credentials**:
   ```xml
   <key>GOOGLE_CLIENT_ID</key>
   <string>your-actual-client-id.apps.googleusercontent.com</string>
   <key>GOOGLE_CLIENT_SECRET</key>
   <string>your-actual-client-secret</string>
   ```

4. **Customize sync interval** (optional):
   - Change `<string>1h</string>` under `--interval` to your preferred interval
   - Examples: `15m`, `30m`, `2h`, `4h`, `24h`
   - Minimum: `5m` (to respect API rate limits)

5. **Customize services** (optional):
   - Change `<string>all</string>` under `--services` to specific services
   - Examples: `contacts`, `calendar`, `gmail`, `contacts,calendar`

### Step 3: Load and Start the Service

```bash
# Load the service (this starts it immediately and on login)
launchctl load ~/Library/LaunchAgents/com.pagen.sync.plist

# Verify it's loaded
launchctl list | grep pagen
```

## Managing the Service

### Check Service Status

```bash
# List all user agents and check if pagen is running
launchctl list | grep pagen

# Output shows: PID, Status, Label
# Example: 12345  0  com.pagen.sync
# - PID: Process ID (running)
# - Status: Exit code (0 = success)
# - Label: Service name
```

### View Logs

```bash
# View standard output (sync progress)
tail -f ~/Library/Logs/pagen-sync.log

# View error output (if any)
tail -f ~/Library/Logs/pagen-sync-error.log

# View last 50 lines
tail -n 50 ~/Library/Logs/pagen-sync.log
```

### Stop Service

```bash
# Unload (stop) the service
launchctl unload ~/Library/LaunchAgents/com.pagen.sync.plist
```

### Restart Service

```bash
# Unload and reload to pick up changes
launchctl unload ~/Library/LaunchAgents/com.pagen.sync.plist
launchctl load ~/Library/LaunchAgents/com.pagen.sync.plist
```

### Start Service (if stopped)

```bash
# Load the service
launchctl load ~/Library/LaunchAgents/com.pagen.sync.plist

# Or use bootout/bootstrap (newer macOS)
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.pagen.sync.plist
```

## Configuration Examples

### Example 1: Sync Every 15 Minutes (Contacts and Calendar Only)

```xml
<key>ProgramArguments</key>
<array>
    <string>/usr/local/bin/pagen</string>
    <string>sync</string>
    <string>daemon</string>
    <string>--interval</string>
    <string>15m</string>
    <string>--services</string>
    <string>contacts,calendar</string>
</array>
```

### Example 2: Sync Every 4 Hours (All Services)

```xml
<key>ProgramArguments</key>
<array>
    <string>/usr/local/bin/pagen</string>
    <string>sync</string>
    <string>daemon</string>
    <string>--interval</string>
    <string>4h</string>
    <string>--services</string>
    <string>all</string>
</array>
```

### Example 3: Sync Once Per Day (Calendar Only)

```xml
<key>ProgramArguments</key>
<array>
    <string>/usr/local/bin/pagen</string>
    <string>sync</string>
    <string>daemon</string>
    <string>--interval</string>
    <string>24h</string>
    <string>--services</string>
    <string>calendar</string>
</array>
```

## Troubleshooting

### Service Not Starting

1. **Check if plist is valid**:
   ```bash
   plutil -lint ~/Library/LaunchAgents/com.pagen.sync.plist
   ```
   Should output: `OK`

2. **Check system logs**:
   ```bash
   # macOS 10.12+
   log stream --predicate 'subsystem == "com.apple.launchd"' --level debug

   # Or check Console.app and filter for "pagen"
   ```

3. **Verify pagen path**:
   ```bash
   which pagen
   # Update ProgramArguments in plist if different
   ```

4. **Check permissions**:
   ```bash
   chmod 644 ~/Library/LaunchAgents/com.pagen.sync.plist
   ```

### Service Crashes Immediately

Check error log:
```bash
tail -n 50 ~/Library/Logs/pagen-sync-error.log
```

Common issues:
- **"no authentication token found"**: Run `pagen sync init` first
- **"invalid interval format"**: Check `--interval` value in plist
- **"GOOGLE_CLIENT_ID not set"**: Add credentials to `EnvironmentVariables` section

### OAuth Token Expired

The service will automatically handle token refresh. If authentication fails:

```bash
# Stop service
launchctl unload ~/Library/LaunchAgents/com.pagen.sync.plist

# Re-authenticate
pagen sync init

# Start service
launchctl load ~/Library/LaunchAgents/com.pagen.sync.plist
```

### Update Configuration

After editing the plist:

```bash
# Reload the service
launchctl unload ~/Library/LaunchAgents/com.pagen.sync.plist
launchctl load ~/Library/LaunchAgents/com.pagen.sync.plist
```

### Disable Auto-Start on Login

Edit the plist and set:
```xml
<key>RunAtLoad</key>
<false/>
```

Then reload:
```bash
launchctl unload ~/Library/LaunchAgents/com.pagen.sync.plist
launchctl load ~/Library/LaunchAgents/com.pagen.sync.plist
```

## Uninstallation

```bash
# 1. Unload the service
launchctl unload ~/Library/LaunchAgents/com.pagen.sync.plist

# 2. Remove the plist file
rm ~/Library/LaunchAgents/com.pagen.sync.plist

# 3. Remove logs (optional)
rm ~/Library/Logs/pagen-sync.log
rm ~/Library/Logs/pagen-sync-error.log
```

## Advanced Configuration

### Reduce System Impact

Lower the process priority:

```xml
<!-- Higher nice value = lower priority (0-20) -->
<key>Nice</key>
<integer>15</integer>
```

### Prevent Frequent Restarts

Increase throttle interval if service crashes:

```xml
<!-- Wait 5 minutes before restarting -->
<key>ThrottleInterval</key>
<integer>300</integer>
```

### Run Only When Idle

Add conditions to run only when system is idle:

```xml
<!-- Only run when on AC power -->
<key>StartOnMount</key>
<false/>

<!-- Only run when user is logged in -->
<key>LimitLoadToSessionType</key>
<array>
    <string>Aqua</string>
</array>
```

### Set Resource Limits

Limit CPU usage:

```xml
<!-- Soft/Hard CPU limit (percentage) -->
<key>SoftResourceLimits</key>
<dict>
    <key>CPU</key>
    <integer>50</integer>
</dict>
```

## Security Notes

1. **Credentials in plist**: The plist file contains your Google OAuth credentials
   - Ensure proper permissions: `chmod 600 ~/Library/LaunchAgents/com.pagen.sync.plist`
   - macOS LaunchAgents directory is user-specific and protected

2. **Alternative credential storage**: Use environment file instead:
   ```bash
   # Create credentials file
   cat > ~/.config/pagen/sync-credentials.env << EOF
   export GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   export GOOGLE_CLIENT_SECRET=your-client-secret
   EOF

   chmod 600 ~/.config/pagen/sync-credentials.env
   ```

   Then modify plist to source it:
   ```xml
   <key>ProgramArguments</key>
   <array>
       <string>/bin/sh</string>
       <string>-c</string>
       <string>source ~/.config/pagen/sync-credentials.env && /usr/local/bin/pagen sync daemon --interval 1h --services all</string>
   </array>
   ```

## Monitoring

### Watch Live Sync Activity

```bash
# Tail sync log in real-time
tail -f ~/Library/Logs/pagen-sync.log | grep -E "(Starting|completed|failed)"
```

### Check Recent Sync Results

```bash
# Last 20 sync cycles
grep "Sync cycle completed" ~/Library/Logs/pagen-sync.log | tail -20
```

### Set Up Notifications

Use a monitoring tool like `fswatch` to get desktop notifications:

```bash
# Install fswatch via Homebrew
brew install fswatch

# Watch for errors
fswatch ~/Library/Logs/pagen-sync-error.log | while read f; do
    osascript -e 'display notification "Pagen sync error detected" with title "Pagen Sync"'
done
```

## Helpful Commands

```bash
# List all user agents
launchctl list

# Print service details
launchctl print gui/$(id -u)/com.pagen.sync

# Check if service is loaded
launchctl list | grep com.pagen.sync

# View service output in real-time
tail -f ~/Library/Logs/pagen-sync.log

# Search logs for errors
grep -i error ~/Library/Logs/pagen-sync-error.log

# Count successful syncs
grep "sync completed" ~/Library/Logs/pagen-sync.log | wc -l
```
