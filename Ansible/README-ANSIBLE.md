# Ansible Deployment for Clip Service

This Ansible playbook automates the deployment of the Clip distributed network service to your Raspberry Pi.

## Prerequisites

1. **Ansible installed** on your local machine:
   ```bash
   # On Ubuntu/Debian
   sudo apt install ansible
   
   # On macOS
   brew install ansible
   ```

2. **SSH access** to your Raspberry Pi:
   - Ensure you can SSH to `192.168.50.200` as user `pi`
   - Set up SSH key authentication (recommended):
     ```bash
     ssh-copy-id pi@192.168.50.200
     ```

3. **Sudo privileges** on the Raspberry Pi for user `pi`

## Configuration

### 1. Update Git Repository URL

Edit `vars.yml` and update the `clip_git_repo` variable with your actual Git repository URL:

```yaml
clip_git_repo: "https://github.com/yourusername/clip.git"
```

Or if using SSH authentication:
```yaml
clip_git_repo: "git@github.com:yourusername/clip.git"
```

### 2. Optional: Customize Service Settings

In `vars.yml`, you can also customize:
- `service_port`: HTTP API port (default: 8080)
- `service_id`: Service instance identifier (default: hostname)
- `advertise_ip`: IP address to advertise (default: auto-detected)
- `seed_nodes`: Comma-separated list of seed nodes for faster discovery
- `go_version`: Go version to install (default: 1.21.5)

### 3. Update Inventory (if needed)

The `inventory.ini` file already contains your Raspberry Pi:
```ini
[myhosts]
192.168.50.200
```

You can add more hosts if deploying to multiple Raspberry Pis.

## Deployment

### Deploy the Service

Run the playbook to deploy the Clip service:

```bash
ansible-playbook deploy-clip.yml -e @vars.yml --ask-become-pass
```

Or if you have passwordless sudo configured:
```bash
ansible-playbook deploy-clip.yml -e @vars.yml
```

### First Time Deployment

If this is your first time connecting to the Raspberry Pi, you might need to provide the SSH password:

```bash
ansible-playbook deploy-clip.yml -e @vars.yml --ask-pass --ask-become-pass
```

## What the Playbook Does

1. ✅ Updates apt package cache
2. ✅ Installs required packages (git, wget, tar, ufw)
3. ✅ Downloads and installs Go (ARM64 version for Raspberry Pi)
4. ✅ Creates a dedicated system user (`clip`)
5. ✅ Clones your Git repository
6. ✅ Builds the Go application
7. ✅ Configures firewall rules:
   - TCP port 8080 (HTTP API)
   - UDP port 9999 (broadcast discovery)
8. ✅ Creates and configures systemd service
9. ✅ Starts and enables the service
10. ✅ Verifies the service is running

## Post-Deployment

### Check Service Status

```bash
# On the Raspberry Pi
systemctl status clip

# View logs
journalctl -u clip -f

# Check API status
curl http://192.168.50.200:8080/status
```

### Manage the Service

```bash
# Start service
sudo systemctl start clip

# Stop service
sudo systemctl stop clip

# Restart service
sudo systemctl restart clip

# View logs
sudo journalctl -u clip -f
```

### Service Endpoints

Once deployed, the following endpoints are available:

- `http://192.168.50.200:8080/status` - Service status and peer list
- `http://192.168.50.200:8080/peers` - List all known peers
- `http://192.168.50.200:8080/join` - Join endpoint (internal use)
- `http://192.168.50.200:8080/heartbeat` - Heartbeat endpoint (internal use)
- `http://192.168.50.200:8080/gossip` - Gossip endpoint (internal use)

## Updating the Service

To update the service with new code from Git:

```bash
ansible-playbook deploy-clip.yml -e @vars.yml --tags=update
```

Or simply re-run the full playbook:
```bash
ansible-playbook deploy-clip.yml -e @vars.yml
```

## Multi-Node Deployment

To deploy to multiple Raspberry Pis, add them to `inventory.ini`:

```ini
[myhosts]
192.168.50.200
192.168.50.201
192.168.50.202
```

Then deploy to all:
```bash
ansible-playbook deploy-clip.yml -e @vars.yml
```

The nodes will automatically discover each other via broadcast!

For faster initial discovery, you can configure seed nodes in `vars.yml`:
```yaml
seed_nodes: "http://192.168.50.200:8080"
```

## Troubleshooting

### SSH Connection Issues

```bash
# Test SSH connection
ansible myhosts -m ping

# If connection fails, check SSH manually
ssh pi@192.168.50.200
```

### Service Won't Start

```bash
# Check service logs on Raspberry Pi
sudo journalctl -u clip -n 50 --no-pager

# Check if Go is installed
/usr/local/go/bin/go version

# Verify binary exists
ls -la /opt/clip/clip
```

### Firewall Issues

```bash
# Check firewall status
sudo ufw status

# Manually allow ports if needed
sudo ufw allow 8080/tcp
sudo ufw allow 9999/udp
```

### Go Build Fails

```bash
# Check available disk space
df -h

# Check Go installation
/usr/local/go/bin/go version

# Try building manually
cd /opt/clip/src
sudo -u clip /usr/local/go/bin/go build -o /opt/clip/clip
```

## Directory Structure

After deployment, the following structure is created on the Raspberry Pi:

```
/opt/clip/
├── clip              # Compiled binary
├── src/             # Git repository
│   ├── main.go
│   ├── service.go
│   └── ...
└── go/              # Go modules cache

/etc/systemd/system/
└── clip.service     # Systemd service file
```

## Uninstalling

To remove the Clip service:

```bash
# On the Raspberry Pi
sudo systemctl stop clip
sudo systemctl disable clip
sudo rm /etc/systemd/system/clip.service
sudo systemctl daemon-reload
sudo userdel clip
sudo rm -rf /opt/clip
```

## License

Same as the Clip project (MIT)
