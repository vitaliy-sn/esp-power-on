#!/bin/bash

release_name="esp-power-on"
namespace="esp-power-on"

if [[ $# -eq 0 ]]; then
  options=("Install: create namespace and install release" \
           "Upgrade: update release" \
           "Uninstall: remove release and namespace" \
           "Exit")
  commands=("install" "upgrade" "uninstall" "exit")
  selected=0

  while true; do
    clear
    echo "Select a command (use ↑/↓ and Enter):"
    for i in "${!options[@]}"; do
      if [[ $i -eq $selected ]]; then
        echo -e "  \033[1;32m> ${options[$i]}\033[0m"
      else
        echo "    ${options[$i]}"
      fi
    done

    read -rsn1 key
    if [[ $key == $'\x1b' ]]; then
      read -rsn2 -t 0.1 key # read 2 more chars
      if [[ $key == "[A" ]]; then
        ((selected--))
        ((selected<0)) && selected=$((${#options[@]}-1))
      elif [[ $key == "[B" ]]; then
        ((selected++))
        ((selected>=${#options[@]})) && selected=0
      fi
    elif [[ $key == "" ]]; then
      clear
      echo "You have selected: ${options[$selected]}"
      read -rp "Are you sure you want to proceed? [y/N]: " confirm
      if [[ "$confirm" =~ ^[Yy]$ ]]; then
        case "${commands[$selected]}" in
          install)
            set -- install
            break
            ;;
          upgrade)
            set -- upgrade
            break
            ;;
          uninstall)
            set -- uninstall
            break
            ;;
          exit)
            exit 0
            ;;
        esac
      fi
    fi
  done
fi

if [[ "$1" == "help" ]]; then
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  install     Install: create namespace and install release"
  echo "  upgrade     Upgrade: update release"
  echo "  uninstall   Uninstall: remove release and namespace"
  echo "  help        Show this help message"
  exit 0
elif [[ "$1" == "install" ]]; then
  # Install: create namespace and install release
  if ! kubectl get namespace "$namespace" >/dev/null 2>&1; then
    kubectl create namespace "$namespace"
  fi
  if helm status "$release_name" --namespace "$namespace" >/dev/null 2>&1; then
    echo "Release '$release_name' already exists in namespace '$namespace'."
    exit 1
  fi
  helm install "$release_name" . --namespace "$namespace"
elif [[ "$1" == "upgrade" ]]; then
  # Upgrade: update release
  helm upgrade "$release_name" . --namespace "$namespace"
elif [[ "$1" == "uninstall" ]]; then
  # Uninstall: remove release and namespace
  helm uninstall "$release_name" --namespace "$namespace"
  kubectl delete namespace "$namespace"
else
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  install     Install: create namespace and install release"
  echo "  upgrade     Upgrade: update release"
  echo "  uninstall   Uninstall: remove release and namespace"
  echo "  help        Show this help message"
  exit 1
fi
