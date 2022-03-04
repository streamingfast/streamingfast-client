#!/bin/bash
# Copyright 2021 dfuse Platform Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

# Protobuf definitions (required to be a sibling of this repository, no check perform yet, needs to be done manually)
PROTO_ETHEREUM=${1:-"$ROOT/../proto-ethereum/"}
PROTO_SOLANA=${1:-"$ROOT/../proto-solana/"}
PROTO_NEAR=${1:-"$ROOT/../proto-near/"}

function main() {
  checks

  current_dir="`pwd`"
  trap "cd \"$current_dir\"" EXIT

  pushd "$ROOT/pb" &> /dev/null
  generate $PROTO_ETHEREUM "sf/ethereum/codec/v1/codec.proto"
  generate $PROTO_ETHEREUM "sf/ethereum/transform/v1/transforms.proto"

  generate $PROTO_SOLANA "sf/solana/codec/v1/codec.proto"
  generate $PROTO_SOLANA "sf/solana/transforms/v1/transforms.proto"

  generate $PROTO_NEAR "sf/near/codec/v1/codec.proto"
  test -e ${PROTO_NEAR}/sf/near/transform/v1/transform.proto && 
    generate $PROTO_NEAR "sf/near/transform/v1/transform.proto"

  echo "generate.sh - `date` - `whoami`" > $ROOT/pb/last_generate.txt
  echo "streamingfast/proto-ethereum revision: `GIT_DIR=$PROTO_ETHEREUM.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt
  echo "streamingfast/proto-solana revision: `GIT_DIR=$PROTO_SOLANA.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt
  echo "streamingfast/proto-near revision: `GIT_DIR=$PROTO_NEAR.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt
}

# usage:
# - generate <protoPath>
# - generate <protoBasePath/> [<file.proto> ...]
function generate() {
    base=""
    if [[ "$#" -gt 1 ]]; then
      base="$1"; shift
    fi

    for file in "$@"; do
      protoc \
      -I$PROTO_ETHEREUM \
      -I$PROTO_SOLANA \
      -I$PROTO_NEAR \
        --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
         $base$file
    done
}

function checks() {
  # The old `protoc-gen-go` did not accept any flags. Just using `protoc-gen-go --version` in this
  # version waits forever. So we pipe some wrong input to make it exit fast. This in the new version
  # which supports `--version` correctly print the version anyway and discard the standard input
  # so it's good with both version.
  result=`printf "" | protoc-gen-go --version 2>&1 | grep -Eo v[0-9\.]+`
  if [[ "$result" == "" ]]; then
    echo "Your version of 'protoc-gen-go' (at `which protoc-gen-go`) is not recent enough."
    echo ""
    echo "To fix your problem, perform those commands:"
    echo ""
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0"
    echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0"
    echo ""
    echo ""
    echo "If everything is working as expetcted, the command:"
    echo ""
    echo "  protoc-gen-go --version"
    echo ""
    echo "Should print 'protoc-gen-go v1.25.0' (if it just hangs, you don't have the correct version)"
    exit 1
  fi
}

main "$@"
