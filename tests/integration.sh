#!/usr/bin/env bash

ENDPOTINT="http://localhost:3456"
MAC="00:00:00:00:00:01"
CLIENT_ID="uuid-1234-5678-9012-3456"
UA="xiaozhi-esp32"
LANG="zh-CN"
CURRENT_DIR=$(dirname "$(readlink -f "$0")")

# make sure utils command exists
if ! command -v curl &>/dev/null; then
  echo "curl command not found. Please install curl to run this script."
  exit 1
fi

# make sure jq command exists
if ! command -v jq &>/dev/null; then
  echo "jq command not found. Please install jq to run this script."
  exit 1
fi

function GET() {
  curl -s -X GET "$ENDPOTINT/$1"
}

function POST() {
  curl -s -X POST "$ENDPOTINT/$1" -d "${2}" "${@:3}"
}

# display message in GREEN color function OK() { echo -e "\033[0;32m$1\033[0m" }
function OK() {
  echo -e "\033[0;32m$1\033[0m"
}

# display message in RED color
function FAIL() {
  echo -e "\033[0;31m$1\033[0m"
}

# display message in block color
function BLOCK() {
  echo -e "\033[0;34m$1\033[0m"
}

#### testing function here
function test_health() {
  BLOCK "Testing ping..."

  body=$(GET "health")
  # check if response body message field is "ping"
  if [[ $(echo "$body" | jq -r '.message') == "ok" ]]; then
    OK "Ping test passed."
  else
    FAIL "Ping test failed. Response: $body"
    exit 1
  fi
}

function test_ota() {
  BLOCK "Testing OTA..."
  deviceBody=$(cat "${CURRENT_DIR}"/example.json)
  body=$(curl -s -XPOST "${ENDPOTINT}/xiaozhi/ota/" -d "${deviceBody}" -H "Content-Type: application/json" -H "Client-Id: ${CLIENT_ID}" -H "User-Agent: ${UA}" -H "Accept-Language: ${LANG}" \ -H "Device-Id: ${MAC}")
  echo ${body}
}

test_health
test_ota
