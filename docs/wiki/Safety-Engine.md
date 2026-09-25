# Safety Engine & Commit-Confirm

Modifying remote network devices can lead to catastrophic administrative lockouts if an IP, interface, or routing change is applied inadvertently. Michibiki includes a multi-layered safety engine.

---

## 1. Pre-Flight Safety Inspection

Before committing any configuration file, Michibiki inspects candidate changes:

```bash
michibiki -d core-fw config validate candidate.conf
```

The safety engine flags:
- **Lockout risks**: Verifies that the current management IP and interface remain reachable in the new configuration.
- **Default route deletion**: Detects deletion of default gateways (`0.0.0.0/0`).
- **Global drop policies**: Identifies blanket firewall block rules that could isolate the administrator.

---

## 2. Commit-Confirm Rollback Timer

When applying changes, specify a confirmation timeout in seconds with `--confirm`:

```bash
michibiki -d core-fw config commit candidate.conf --confirm 60
```

1. Michibiki applies the candidate configuration to the router.
2. An automatic rollback timer begins counting down (e.g. 60 seconds).
3. If the administrator loses connectivity, the router or background session automatically restores the previous configuration when the timer expires.
4. If connectivity is verified, the administrator makes the change permanent:

```bash
michibiki -d core-fw config rollback confirm <id>
```

---

## 3. Viewing Configuration Diff

Compare candidate configurations against the running state:

```bash
michibiki -d core-fw config diff candidate.conf
```
