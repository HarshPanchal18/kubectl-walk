# Bash completion for kubectl-walk.
#
# Source this file after the standard kubectl completion script. It completes
# both `kubectl-walk` and `kubectl walk` without replacing kubectl's existing
# completion for other commands.

# Resource completion
_kubectl_walk_resource_types() {
    command kubectl api-resources --no-headers -o name 2>/dev/null
}

# Namespace completion
_kubectl_walk_namespaces() {
    command kubectl get namespaces --no-headers -o custom-columns=":metadata.name" 2>/dev/null
}

# Resource name completion
_kubectl_walk_resource_names() {
    local resource="$1" namespace="$2" item namespaced

    # Find the NAMESPACED column for the requested resource
    namespaced="$(
        command kubectl api-resources --no-headers 2>/dev/null |
        awk -v resource="$resource" '$1 == resource {print $(NF-1);exit}'
    )"

    # Namespaced resource
    while IFS= read -r item; do
        printf '%s\n' "${item#*/}"
    done < <(command kubectl get "$resource" --namespace "$namespace" --no-headers -o name 2>/dev/null)

    # Cluster-scoped resource
    while IFS= read -r item; do
        printf '%s\n' "${item#*/}"
    done < <(command kubectl get "$resource" --no-headers -o name 2>/dev/null)

}

# File completion
_kubectl_walk_complete_files() {
    local cur="$1"
    COMPREPLY=( $(compgen -f -- "$cur") )
}

# Main completion function
_kubectl_walk() {
    local cur prev command_offset arg namespace="" file_source="" selector=""
    local -a walk_args positionals
    local i

    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD - 1]}"

    # The plugin can be invoked either as kubectl-walk or kubectl walk.
    command_offset=1
    if [[ ${COMP_WORDS[0]} == kubectl ]]; then
        command_offset=2
    fi
    i=$command_offset

    # Complete values for flags whose values come immediately afterwards.
    case "$prev" in
        -n|--namespace)
            COMPREPLY=( $(compgen -W "$(_kubectl_walk_namespaces)" -- "$cur") )
            return
            ;;
        -f|--file|-o|--output|--kubeconfig)
            _kubectl_walk_complete_files "$cur"
            return
            ;;
        -d|--depth)
            # Numeric value.
            return
            ;;

        -g|--grep|--find|--entry)
            # Free-form value.
            return
            ;;

        -l|--selector)
            # Free-form label selector.
            return
            ;;
    esac

    # Flags can appear before, between, or after positional arguments. Gather
    # only the latter while retaining the namespace and selector values needed
    # for live API completion.
    while (( i < COMP_CWORD )); do
        arg="${COMP_WORDS[i]}"

        case "$arg" in

        # Namespace
            -n|--namespace)
                ((i++))
                namespace="${COMP_WORDS[i]}"
                ;;
            --namespace=*) namespace="${arg#*=}" ;;
            -n?*) namespace="${arg#-n}" ;;

        # File
            -f|--file)
                ((i++))
                file_source="${COMP_WORDS[i]}"
                ;;
            --file=*) file_source="${arg#*=}" ;;
            -f?*) file_source="${arg#-f}" ;;

        # Selector
            -l|--selector)
                ((i++))
                selector="${COMP_WORDS[i]}"
                ;;
            --selector=*) selector="${arg#*=}" ;;
            -l?*) selector="${arg#-l}" ;;

        # Flags which consume the next argument
            -e|--entry|-o|--output|--kubeconfig|-g|--grep|--find|-d|--depth)
                ((i++))
                ;;

        # Flags using --flag=value
            --entry=*|--output=*|--kubeconfig=*|--grep=*|--find=*|--depth=*) ;;

        # Short flags using attached values
            -e?*|-o?*|-g?*|-d?*) ;;

        # Other flags
            -*) ;;

        # Positional argument
            *) positionals+=("$arg") ;;
        esac
        ((i++))
    done

    # A YAML source does not take resource arguments.
    if [[ -n $file_source ]]; then
        return
    fi

    # Flag completion
    if [[ $cur == -* ]]; then
        COMPREPLY=( $(compgen -W '--all --depth --entry --file --find --grep --help --keys --kubeconfig --namespace --output --pure --selector --tree --values --version -A -d -e -f -g -h -k -l -n -o -p -t -v' -- "$cur") )
        return
    fi

    # No positional arguments yet - kubectl walk dep<TAB>
    if (( ${#positionals[@]} == 0 )); then
        COMPREPLY=( $(compgen -W "$(_kubectl_walk_resource_types)" -- "$cur") )
    elif (( ${#positionals[@]} == 1 )) && [[ -z $selector ]]; then

        # Use explicitly supplied namespace
        namespace="${namespace:-$(command kubectl config view --minify -o 'jsonpath={..namespace}' 2>/dev/null)}"

        # Fall back to default namespace
        namespace="${namespace:-default}"

        # Get resource names
        COMPREPLY=( $(compgen -W "$(_kubectl_walk_resource_names "${positionals[0]}" "$namespace")" -- "$cur") )
    fi
}

# Standalone plugin completion
complete -o default -F _kubectl_walk kubectl-walk

# kubectl integration
# Preserve standard kubectl completion and only take over when its first
# subcommand is the walk plugin. Source `kubectl completion bash` first.
if declare -F __start_kubectl >/dev/null; then
    _kubectl_with_walk_completion() {
        if [[ ${COMP_WORDS[1]} == walk ]]; then
            _kubectl_walk
        else
            __start_kubectl
        fi
    }

    complete -o default -F _kubectl_with_walk_completion kubectl
fi
