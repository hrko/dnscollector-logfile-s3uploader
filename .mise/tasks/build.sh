#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# Target Operating Systems
TARGET_OS=("linux" "freebsd" "darwin" "windows")
# Target Architectures
TARGET_ARCH=("amd64" "arm64")

# --- Function for logging ---
log() {
  echo "--> $1"
}

# --- Function to show usage ---
usage() {
  echo "Usage: $0 [-o <output_dir>] <source_dir> [program_name]"
  echo
  echo "Arguments:"
  echo "  <source_dir>          Required: The directory containing the Go source code (e.g., ./cmd/my_app)."
  echo "  [program_name]        Optional: The name of the output binary. Defaults to the basename of <source_dir>."
  echo
  echo "Options:"
  echo "  -o, --output <dir>    Optional: The directory to place the compiled binaries. Defaults to the current directory ('./build')."
  exit 1
}

# --- Argument Parsing ---
OUTPUT_DIR="./build" # Default output directory

while [[ "$#" -gt 0 ]]; do
  case $1 in
  -o | --output)
    if [[ -n "$2" ]]; then
      OUTPUT_DIR="$2"
      shift 2
    else
      echo "Error: --output requires a non-empty option argument."
      exit 1
    fi
    ;;
  -h | --help)
    usage
    ;;
  *)
    # Assume the rest are positional arguments
    break
    ;;
  esac
done

# Positional arguments
SOURCE_DIR=$1
PROG_NAME=$2

# Check if source directory is provided
if [[ -z "$SOURCE_DIR" ]]; then
  echo "Error: Source directory is not specified."
  usage
fi

# Check if source directory exists
if [[ ! -d "$SOURCE_DIR" ]]; then
  echo "Error: Source directory '$SOURCE_DIR' not found."
  exit 1
fi

# Make source directory is an absolute path
SOURCE_DIR=$(realpath "$SOURCE_DIR")

# If program name is not provided, use the source directory's basename
if [[ -z "$PROG_NAME" ]]; then
  PROG_NAME=$(basename "$SOURCE_DIR")
fi

# --- Main Script ---
log "Starting cross-compilation for '$PROG_NAME'..."
log "Source: $SOURCE_DIR"
log "Output will be saved to: $OUTPUT_DIR"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Loop through each OS and Architecture
for os in "${TARGET_OS[@]}"; do
  for arch in "${TARGET_ARCH[@]}"; do
    log "Compiling for $os/$arch..."

    # Set environment variables for the go build command
    export GOOS=$os
    export GOARCH=$arch

    # Set the output file name, add .exe for Windows
    OUTPUT_NAME="${PROG_NAME}_${os}_${arch}"
    if [[ "$os" == "windows" ]]; then
      OUTPUT_NAME+=".exe"
    fi

    # Define the full output path
    OUTPUT_PATH="$OUTPUT_DIR/$OUTPUT_NAME"

    # Build the program
    # CGO_ENABLED=0 is important for cross-compilation to avoid C dependencies.
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$OUTPUT_PATH" "$SOURCE_DIR"
  done
done

log "Cross-compilation finished successfully."
