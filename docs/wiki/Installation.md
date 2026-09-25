# Installation

Michibiki is distributed as a single self-contained binary with zero external runtime dependencies.

## Option 1: Via Go Install (Recommended)

Requires Go 1.24+:

```bash
# Install tagged release
go install github.com/sekai-labs/michibiki/cmd/michibiki@v0.1.0

# Or install latest commit from main
go install github.com/sekai-labs/michibiki/cmd/michibiki@latest
```

Make sure `$GOPATH/bin` or `~/go/bin` is in your system `$PATH`:
```bash
export PATH="$HOME/go/bin:$PATH"
```

## Option 2: Building from Source

```bash
git clone https://codeberg.org/sekai-labs/michibiki.git
cd michibiki
go build -o michibiki ./cmd/michibiki
sudo mv michibiki /usr/local/bin/
```

## Verification

Verify that the installation was successful:

```bash
michibiki --help
michibiki plugin list
```
