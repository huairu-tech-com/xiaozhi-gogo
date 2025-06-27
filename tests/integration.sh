#!/usr/bin/env bash

ENDPOTINT="http://localhost:3456"

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
  curl -s -X POST "$ENDPOTINT/$1" -d "$2"
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

test_health
