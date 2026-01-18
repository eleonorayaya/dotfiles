package kubernetes

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		Aliases: []shizukuenv.Alias{
			{Name: "ktl", Command: "kubectl"},
		},
		Functions: []shizukuenv.ShellFunction{
			{Name: "ktls", Body: ktlsFunction},
			{Name: "ktld", Body: ktldFunction},
			{Name: "ktlog", Body: ktlogFunction},
			{Name: "ktlsh", Body: ktlshFunction},
		},
	}, nil
}

const ktlsFunction = `    namespace="${1:-default}"
    if [[ "$2" == "-w" ]]; then
        watch kubectl get pods -n "$namespace"
    else
        kubectl get pods -n "$namespace"
    fi`

const ktldFunction = `    namespace="${1:-default}"
    pods=$(kubectl get pods -n "$namespace" --no-headers -o custom-columns=":metadata.name")
    selected=$(echo "$pods" | fzf)
    if [[ -n "$selected" ]]; then
        kubectl describe pod "$selected" -n "$namespace"
    fi`

const ktlogFunction = `    namespace="default"
    follow=""
    label_filter=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -f) follow="-f" ;;
            -l) label_filter="$2"; shift ;;
            *) namespace="$1" ;;
        esac
        shift
    done

    if [[ -n "$label_filter" ]]; then
        pods=$(kubectl get pods -n "$namespace" -l "$label_filter" --no-headers -o custom-columns=":metadata.name")
    else
        pods=$(kubectl get pods -n "$namespace" --no-headers -o custom-columns=":metadata.name")
    fi

    selected=$(echo "$pods" | fzf)
    if [[ -n "$selected" ]]; then
        kubectl logs "$selected" -n "$namespace" $follow
    fi`

const ktlshFunction = `    namespace="${1:-default}"
    pods=$(kubectl get pods -n "$namespace" --no-headers -o custom-columns=":metadata.name")
    selected=$(echo "$pods" | fzf)
    if [[ -n "$selected" ]]; then
        kubectl exec -it "$selected" -n "$namespace" -- /bin/sh
    fi`
