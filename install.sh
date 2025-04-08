#!/bin/bash
# Install `imt`

function main {
  local p="install: error:"
  local ARCH=""
  local OS=""
  local VERSION=""

  # check that required one of the commands are installed before doing anything.
  if command -v curl &>/dev/null; then
    cmd="curl -sSL"
  elif command -v wget &>/dev/null; then
    cmd="wget -qO -"
  else
    echo "${p} neither curl nor wget was found, please install of them and try again!"
    exit 1
  fi

  # Get the latest version
  VERSION=$(${cmd} "https://api.github.com/repos/faabiosr/imt/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | cut -c 2-)

  if [ -z "${VERSION}" ]; then
    echo "${p} failed to get the latest version, please check your network connection and try again!"
    exit 1
  fi

  echo "installing imt v${VERSION} ..."

  # Check if the OS is supported
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  if [ "${OS}" = "darwin" ]; then
    OS="macos"
  fi

  ARCH=$(uname -m)
  if [ "${ARCH}" = "arm64" ]; then
    ARCH="aarch64"
  fi

  local FILENAME="imt_${VERSION}_${OS}_${ARCH}"
  local INSTALL_PATH="${INSTALL_PATH:-${HOME}/.local/bin}"

  echo "https://github.com/faabiosr/imt/releases/download/v${VERSION}/${TAR_FILE}"
  echo "installation path: ${INSTALL_PATH}"

  mkdir -p "${INSTALL_PATH}" >/dev/null 2>&1 || {
    >&2 echo "${p} failed to create ${INSTALL_PATH} directory, please check your sudo permissions and try again!"
    exit 1
  }

  ${cmd} "https://github.com/faabiosr/imt/releases/download/v${VERSION}/${FILENAME}.tar.gz" |
    tar -xzvf - -C "${INSTALL_PATH}" "${FILENAME}/imt" --strip-components=1

  echo "imt installed successfully!"
}

main "$@"
