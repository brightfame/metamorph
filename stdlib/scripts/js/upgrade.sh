#!/bin/bash

set -e

function log {
  local -r level="$1"
  shift
  local -ra message=("$@")
  local -r timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  local -r scriptname="$(basename "$0")"

  echo >&2 -e "${timestamp} [${level}] [$scriptname] ${message[*]}"
}

function log_info {
  local -ra message=("$@")
  log "INFO" "${message[@]}"
}

function log_warn {
  local -ra message=("$@")
  log "WARN" "${message[@]}"
}

function log_error {
  local -ra message=("$@")
  log "ERROR" "${message[@]}"
}

# Check that the given binary is available on the PATH. If it's not, exit with an error.
function assert_is_installed {
  local -r name="$1"

  if [[ ! "$(command -v "${name}")" ]]; then
    log_error "The binary '$name' is required by this script but is not installed or in the system's PATH."
    exit 1
  fi
}

checkout_git_branch() {
  local branch_name="$1"

  # check if we're currently on the branch
  local current_branch=$(git rev-parse --abbrev-ref HEAD)

  # if we're not on the branch, check it out
  if [[ "$current_branch" != "$branch_name" ]]; then
    git checkout -b "$branch_name"
  fi
}

upgrade_dockerfile() {
  local filename="$1"

  # Check if the file exists
  if [[ ! -f "$filename" ]]; then
    log_warn "Dockerfile '$filename' does not exist. Skipping update."
    return 0
  fi

  # Apply a series of sed commands
  sed -i '' \
    -e 's/alpine-node18/alpine-node22/' \
    -e 's/node:18/node:22/' \
    "$filename"

  log_info "Dockerfile: $filename updated successfully."
}

yarn_install() {
  log_info "Running yarn install..."
  yarn install
}

# Detect and upgrade a Yarn package
update_and_verify_yarn_package() {
  local package_name="$1"
  local target_version="$2"

  # Check if package.json exists
  if [[ ! -f "package.json" ]]; then
    log_error "Error: package.json not found in the current directory."
    return 1
  fi

  # Check if the package is listed as a dependency or devDependency
  if yarn why $package_name >/dev/null 2>&1; then
    log_info "Package '$package_name' found. Updating..."
    yarn add "$package_name"
  else
    log_info "Package '$package_name' is not installed."
  fi
}

run_yarn_tests() {
  log_info "Running yarn test..."
  yarn test
}

# Build the Docker container
build_docker_container() {
  local dockerfile="$1"
}

function run {
  assert_is_installed "sed"
  assert_is_installed "jq"
  assert_is_installed "yq"

  # checkout a new git branch
  checkout_git_branch node-22

  # Update the Dockerfiles to use Node 22
  upgrade_dockerfile Dockerfile
  upgrade_dockerfile Dockerfile.e2e

  # Update the Yarn dependencies
  yarn_install
  update_and_verify_yarn_package "@types/node" "^22.10.5"
  update_and_verify_yarn_package "@types/amqplib" "^0.10.5"

  # attempt to build the container locally
  #build_docker_container Dockerfile

  # run the tests
  run_yarn_tests
}

run "$@"
