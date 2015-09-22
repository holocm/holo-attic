#compdef holo
typeset -A opt_args

(( $+functions[_holo_command] )) || _holo_command()
{
    local -a _commands
    _commands=(
        'apply:Apply available configuration to some or all targets'
        'scan:Scan for configuration targets'
    )
    _describe -t commands 'holo command' _commands
    return 0
}

(( $+functions[_holo_target] )) || _holo_target()
{
    _alternative "targets:configuration targets:($(holo scan --short))"
    return 0
}

(( $+functions[_holo_zsh_comp] )) || _holo_zsh_comp()
{
    if (( CURRENT == 2 )); then
        _arguments : \
            '--help[Print short usage information.]' \
            '--version[Print a short version string.]' \
            '1::holo command:_holo_command'
    else
        case "$words[2]" in
            apply)
                _arguments : \
                    {-f,--force}'[overwrite manual changes on targets]' \
                    '*:target:_holo_target'
                ;;
            scan)
                local -a _options
                _options=(
                    {-s,--short}':print only target names'
                )
                _describe -t scan_opts 'holo scan option' _options
                ;;
        esac
    fi
    return 0
}

_holo_zsh_comp "$@"
