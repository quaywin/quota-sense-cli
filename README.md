# QuotaSense CLI 🛡️

QuotaSense CLI is a powerful terminal utility to monitor and manage your AI model usage quotas. Inspired by the QuotaSense menu bar app, it brings real-time quota visibility directly to your terminal.

## Features

- **Multi-Account Support**: Monitor multiple accounts and providers in one view.
- **Color-coded Status**: Instantly identify low quotas with visual cues (Green/Yellow/Red).
- **Lightweight & Fast**: Written in Go for maximum performance.
- **Easy Installation**: Simple one-line install script.

## Installation

### Quick Install (Recommended)

Run the following command in your terminal:

```bash
curl -sSL https://raw.githubusercontent.com/quaywin/quota-sense-cli/main/install.sh | bash
```

### From Source

If you have Go installed, you can build it locally:

```bash
git clone https://github.com/quaywin/quota-sense-cli.git
cd quota-sense-cli
make install
```

## Usage

### 1. Initial Configuration

The first time you run `qs`, it will prompt you for your server configuration:

```bash
qs
```

You will need:
- **Remote Server URL**: The address of your QuotaSense server.
- **Management Token**: Your secret key for authentication.

### 2. View Quotas

Simply run the main command to see all your active model limits:

```bash
qs
```

### 3. Other Commands

- `qs version`: Show current version.
- `qs plan`: Review planned usage (simulation).
- `qs --help`: List all available commands and flags.

## Configuration

Configuration is stored locally in `~/.quota-sense.json`. To reset your configuration, simply delete this file:

```bash
rm ~/.quota-sense.json
```

## Development

To build the project for development:

```bash
make build
./qs
```

To generate release binaries:

```bash
make release
```

---
Built with 💙 for the AI Developer Community.
