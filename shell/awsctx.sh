#!/usr/bin/env bash
#
# awsctx tab completions (optional)
#
# Source this in your ~/.bashrc or ~/.zshrc for tab completion support.
# This is NOT required for awsctx to work — it only adds tab completions.

# --- Bash completions ---
if [[ -n "$BASH_VERSION" ]]; then
  _awsctx_completions() {
    local cur prev subcmd
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
      COMPREPLY=($(compgen -W "profile p region r -h --help -v --version" -- "$cur"))
      return
    fi

    subcmd="${COMP_WORDS[1]}"
    case "$subcmd" in
      profile|p)
        if [[ ${COMP_CWORD} -eq 2 ]]; then
          local profiles
          profiles="$(command awsctx --fzf-list profile 2>/dev/null)"
          COMPREPLY=($(compgen -W "$profiles -c --current - -h --help" -- "$cur"))
        fi
        ;;
      region|r)
        if [[ ${COMP_CWORD} -eq 2 ]]; then
          local regions
          regions="$(command awsctx --fzf-list region 2>/dev/null)"
          COMPREPLY=($(compgen -W "$regions -c --current - -h --help" -- "$cur"))
        fi
        ;;
    esac
  }
  complete -F _awsctx_completions awsctx
fi

# --- Zsh completions ---
if [[ -n "$ZSH_VERSION" ]]; then
  _awsctx() {
    local -a subcmds
    subcmds=(
      'profile:list or switch AWS profiles'
      'p:list or switch AWS profiles'
      'region:list or switch AWS regions'
      'r:list or switch AWS regions'
    )

    if (( CURRENT == 2 )); then
      _describe 'subcommand' subcmds
      return
    fi

    case "${words[2]}" in
      profile|p)
        if (( CURRENT == 3 )); then
          local -a profiles flags
          profiles=("${(@f)$(command awsctx --fzf-list profile 2>/dev/null)}")
          flags=('-c:show current profile' '--current:show current profile' '-:switch to previous')
          _describe 'profile' profiles
          _describe 'flag' flags
        fi
        ;;
      region|r)
        if (( CURRENT == 3 )); then
          local -a regions flags
          regions=("${(@f)$(command awsctx --fzf-list region 2>/dev/null)}")
          flags=('-c:show current region' '--current:show current region' '-:switch to previous')
          _describe 'region' regions
          _describe 'flag' flags
        fi
        ;;
    esac
  }
  compdef _awsctx awsctx
fi
