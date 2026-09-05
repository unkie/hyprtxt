#!/usr/bin/env bash
# Bash completions for hyprtxt

_hyprtxt() {
	local cur prev opts word
	local options_done=false
	local expects_value=false
	local i

	cur="${COMP_WORDS[COMP_CWORD]}"
	prev=""
	if ((COMP_CWORD > 0)); then
		prev="${COMP_WORDS[COMP_CWORD-1]}"
	fi
	opts="-p --prefix -P --postfix -F --font -f --figlet -m --missing -c --charset -e --examples -v --version -h --help"
	COMPREPLY=()

	case "$prev" in
	-p|--prefix|-P|--postfix)
		return 0
		;;
	-F|--font)
		mapfile -t COMPREPLY < <(compgen -W "hyprtxt hyprblk" -- "$cur")
		return 0
		;;
	esac
	if [[ "$cur" == --font=* ]]; then
		local value="${cur#*=}"
		mapfile -t COMPREPLY < <(compgen -P "--font=" -W "hyprtxt hyprblk" -- "$value")
		return 0
	fi

	for ((i = 1; i < COMP_CWORD; i++)); do
		word="${COMP_WORDS[i]}"
		if [[ "$expects_value" == true ]]; then
			expects_value=false
			continue
		fi

		case "$word" in
		--)
			options_done=true
			break
			;;
		-p|--prefix|-P|--postfix|-F|--font)
			expects_value=true
			;;
		-*) ;;
		*)
			options_done=true
			break
			;;
		esac
	done

	if [[ "$options_done" == true ]]; then
		return 0
	fi

	if [[ "$cur" == -* ]]; then
		mapfile -t COMPREPLY < <(compgen -W "$opts" -- "$cur")
	fi

	return 0
}

complete -F _hyprtxt hyprtxt
