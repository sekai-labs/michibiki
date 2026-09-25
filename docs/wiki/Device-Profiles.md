# Device Profiles & Credential Vault

Michibiki stores device connection profiles in `~/.config/michibiki/config.yaml`. Passwords and API secrets are never stored in plaintext.

## Adding Device Profiles

```bash
# OPNsense Firewall
michibiki device add core-fw --provider opnsense --address https://192.168.1.1 --insecure

# MikroTik RouterOS v7
michibiki device add edge-router --provider routeros --address https://10.0.0.1 --insecure

# OpenWrt AP / Gateway
michibiki device add livingroom-ap --provider openwrt --address http://192.168.1.2

# VyOS Router
michibiki device add vyos-gw --provider vyos --address https://172.16.0.1 --insecure

# pfSense Firewall
michibiki device add branch-fw --provider pfsense --address https://192.168.10.1 --insecure

# FRRouting Switch
michibiki device add tor-switch --provider frr --address 10.10.0.1 --port 22
```

## Managing Devices

```bash
# List all profiles
michibiki device list

# Test reachability and authentication
michibiki device test core-fw

# Remove profile
michibiki device remove livingroom-ap
```

## Credential Resolution Hierarchy

When connecting to a device, Michibiki resolves credentials in order:

1. **Environment Variables:**
   - `MICHIBIKI_<DEVICE>_USERNAME`
   - `MICHIBIKI_<DEVICE>_PASSWORD`
   - `MICHIBIKI_<DEVICE>_API_KEY`
   - `MICHIBIKI_<DEVICE>_API_SECRET`
   - `MICHIBIKI_<DEVICE>_TOKEN`
   - `MICHIBIKI_<DEVICE>_SSH_KEY`
   Example:
   ```bash
   export MICHIBIKI_CORE_FW_API_KEY="my-api-key"
   export MICHIBIKI_CORE_FW_API_SECRET="my-api-secret"
   ```

2. **Encrypted File Vault (`vault.enc`):**
   Encrypted with AES-256-GCM. Reference keys inside device profiles:
   ```yaml
   devices:
     core-fw:
       provider: opnsense
       address: https://192.168.1.1
       credential_ref: "vault:core-fw-creds"
   ```

3. **Inline Environment Variable Reference:**
   ```yaml
   credential_ref: "env:MY_ROUTER_TOKEN"
   ```
