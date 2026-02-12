# awsctx

Fast AWS profile and region switcher, inspired by [kubectx](https://github.com/ahmetb/kubectx).

## Features

- Switch AWS profiles and regions with a single command
- Works immediately — no shell wrapper required
- Modifies `[default]` in `~/.aws/config` and `~/.aws/credentials` (original backed up)
- Interactive selection with [fzf](https://github.com/junegunn/fzf) (if installed)
- Switch back to previous profile/region with `-`
- Tab completions for bash, zsh, and fish (optional)
- Current profile/region highlighted in listing

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap yetmike/tap
brew install awsctx
```

### Automatic Install (Linux / macOS)

The easiest way to install without homebrew is via the installation script:

```bash
curl -sL https://raw.githubusercontent.com/yetmike/awsctx/main/install.sh | bash
```

### From Source

```bash
# Using go install
go install github.com/yetmike/awsctx/cmd/awsctx@latest

# Or manual build
git clone https://github.com/yetmike/awsctx
cd awsctx
make install
```

## Usage

```sh
awsctx                          # show current profile and region

# Profile switching
awsctx profile                  # list profiles (interactive fzf if available)
awsctx p dev                    # switch to "dev" profile
awsctx p -c                     # show current profile
awsctx p -                      # switch to previous profile
awsctx p default                # restore original default profile

# Region switching
awsctx region                   # list regions (interactive fzf if available)
awsctx r us-east-1              # switch to us-east-1
awsctx r -c                     # show current region
awsctx r -                      # switch to previous region
```

`p` is short for `profile`, `r` is short for `region`.

## How it works

When you switch to a profile (e.g., `awsctx p dev`):
1. The original `[default]` section is backed up (in both config and credentials files) to `[_awsctx_original_default]`.
2. The target profile's settings are copied into `[default]`.
3. AWS CLI reads `[default]` automatically — no env vars needed.

Run `awsctx p default` to restore the original default profile from the backup.

No shell wrapper or `source` command needed. Just install the binary and use it.

## Tab completions (optional)

For tab completion support, add the following to your shell configuration file (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`):

**Bash/Zsh:**
```bash
source /path/to/awsctx/shell/awsctx.sh
```

**Note:** If installed via Homebrew, completions are handled automatically (ensure your brew shell completion is set up).

## Requirements

- [AWS CLI](https://aws.amazon.com/cli/) installed and configured (`~/.aws/config`)
- [fzf](https://github.com/junegunn/fzf) (optional, for interactive selection)
