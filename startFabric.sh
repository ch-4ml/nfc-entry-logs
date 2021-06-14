#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error
set -e

sudo chown -R bstudent:bstudent ./first-network/*
rm -rf ./app1/wallet/*
rm -rf ./app2/wallet/*
rm -rf ./app3/wallet/*

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1
starttime=$(date +%s)
CC_SRC_LANGUAGE=${1:-"go"}
CC_SRC_LANGUAGE=`echo "$CC_SRC_LANGUAGE" | tr [:upper:] [:lower:]`
if [ "$CC_SRC_LANGUAGE" = "go" -o "$CC_SRC_LANGUAGE" = "golang"  ]; then
	CC_RUNTIME_LANGUAGE=golang
	CC_SRC_PATH=github.com/chaincode/entryLog/go
elif [ "$CC_SRC_LANGUAGE" = "java" ]; then
	CC_RUNTIME_LANGUAGE=java
	CC_SRC_PATH=/opt/gopath/src/github.com/chaincode/fabcar/java
elif [ "$CC_SRC_LANGUAGE" = "javascript" ]; then
	CC_RUNTIME_LANGUAGE=node # chaincode runtime language is node.js
	CC_SRC_PATH=/opt/gopath/src/github.com/chaincode/fabcar/javascript
elif [ "$CC_SRC_LANGUAGE" = "typescript" ]; then
	CC_RUNTIME_LANGUAGE=node # chaincode runtime language is node.js
	CC_SRC_PATH=/opt/gopath/src/github.com/chaincode/fabcar/typescript
	echo Compiling TypeScript code into JavaScript ...
	pushd ../chaincode/fabcar/typescript
	npm install
	npm run build
	popd
	echo Finished compiling TypeScript code into JavaScript
else
	echo The chaincode language ${CC_SRC_LANGUAGE} is not supported by this script
	echo Supported chaincode languages are: go, javascript, and typescript
	exit 1
fi

# clean the keystore
rm -rf ./hfc-key-store

# launch network; create channel and join peer to channel
cd ./first-network
echo y | ./byfn.sh down
echo y | ./byfn.sh up -a -n -s couchdb -o etcdraft

CONFIG_ROOT=/opt/gopath/src/github.com/hyperledger/fabric/peer
ORG1_MSPCONFIGPATH=${CONFIG_ROOT}/crypto/peerOrganizations/org1.dmc.ajou.ac.kr/users/Admin@org1.dmc.ajou.ac.kr/msp
ORG1_TLS_ROOTCERT_FILE=${CONFIG_ROOT}/crypto/peerOrganizations/org1.dmc.ajou.ac.kr/peers/peer0.org1.dmc.ajou.ac.kr/tls/ca.crt
ORG2_MSPCONFIGPATH=${CONFIG_ROOT}/crypto/peerOrganizations/org2.dmc.ajou.ac.kr/users/Admin@org2.dmc.ajou.ac.kr/msp
ORG2_TLS_ROOTCERT_FILE=${CONFIG_ROOT}/crypto/peerOrganizations/org2.dmc.ajou.ac.kr/peers/peer0.org2.dmc.ajou.ac.kr/tls/ca.crt
ORG3_MSPCONFIGPATH=${CONFIG_ROOT}/crypto/peerOrganizations/org3.dmc.ajou.ac.kr/users/Admin@org3.dmc.ajou.ac.kr/msp
ORG3_TLS_ROOTCERT_FILE=${CONFIG_ROOT}/crypto/peerOrganizations/org3.dmc.ajou.ac.kr/peers/peer0.org3.dmc.ajou.ac.kr/tls/ca.crt
ORDERER_TLS_ROOTCERT_FILE=${CONFIG_ROOT}/crypto/ordererOrganizations/dmc.ajou.ac.kr/orderers/orderer.dmc.ajou.ac.kr/msp/tlscacerts/tlsca.dmc.ajou.ac.kr-cert.pem
set -x

echo "Installing smart contract on peer0.org1.dmc.ajou.ac.kr"
docker exec \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_ADDRESS=peer0.org1.dmc.ajou.ac.kr:7051 \
  -e CORE_PEER_MSPCONFIGPATH=${ORG1_MSPCONFIGPATH} \
  -e CORE_PEER_TLS_ROOTCERT_FILE=${ORG1_TLS_ROOTCERT_FILE} \
  cli \
  peer chaincode install \
    -n fabcar \
    -v 1.0 \
    -p "$CC_SRC_PATH" \
    -l "$CC_RUNTIME_LANGUAGE"

echo "Installing smart contract on peer0.org2.dmc.ajou.ac.kr"
docker exec \
  -e CORE_PEER_LOCALMSPID=Org2MSP \
  -e CORE_PEER_ADDRESS=peer0.org2.dmc.ajou.ac.kr:10051 \
  -e CORE_PEER_MSPCONFIGPATH=${ORG2_MSPCONFIGPATH} \
  -e CORE_PEER_TLS_ROOTCERT_FILE=${ORG2_TLS_ROOTCERT_FILE} \
  cli \
  peer chaincode install \
    -n fabcar \
    -v 1.0 \
    -p "$CC_SRC_PATH" \
    -l "$CC_RUNTIME_LANGUAGE"

echo "Installing smart contract on peer0.org3.dmc.ajou.ac.kr"
docker exec \
  -e CORE_PEER_LOCALMSPID=Org3MSP \
  -e CORE_PEER_ADDRESS=peer0.org3.dmc.ajou.ac.kr:11051 \
  -e CORE_PEER_MSPCONFIGPATH=${ORG3_MSPCONFIGPATH} \
  -e CORE_PEER_TLS_ROOTCERT_FILE=${ORG3_TLS_ROOTCERT_FILE} \
  cli \
  peer chaincode install \
    -n fabcar \
    -v 1.0 \
    -p "$CC_SRC_PATH" \
    -l "$CC_RUNTIME_LANGUAGE"    

echo "Instantiating smart contract on dmcchannel"
docker exec \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=${ORG1_MSPCONFIGPATH} \
  cli \
  peer chaincode instantiate \
    -o orderer.dmc.ajou.ac.kr:7050 \
    -C dmcchannel \
    -n fabcar \
    -l "$CC_RUNTIME_LANGUAGE" \
    -v 1.0 \
    -c '{"Args":[]}' \
    -P "OR('Org1MSP.member','Org2MSP.member','Org3MSP.member')" \
    --tls \
    --collections-config /opt/gopath/src/github.com/chaincode/entryLog/collections_config.json \
    --cafile ${ORDERER_TLS_ROOTCERT_FILE} \
    --peerAddresses peer0.org1.dmc.ajou.ac.kr:7051 \
    --tlsRootCertFiles ${ORG1_TLS_ROOTCERT_FILE} \

echo "Waiting for instantiation request to be committed ..."
sleep 10

export entryLogValue=$(echo -n "{\"entryLogID\":\"EntryLog1\",\"facilityID\":\"Facility1\",\"year\":\"1995\",\"sex\":\"1\",\"entryTime\":\"2021-06-14 17:29:30\",\"personalID\":\"Person1\",\"name\":\"Chpark\",\"phone\":\"010-6223-2277\",\"address\":\"경기도 수원시\"}" | base64 | tr -d \\n)

echo "Submitting initLedger transaction to smart contract on dmcchannel"
# echo "The transaction is sent to the two peers with the chaincode installed (peer0.org1.dmc.ajou.ac.kr and peer0.org2.dmc.ajou.ac.kr) so that chaincode is built before receiving the following requests"
docker exec \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=${ORG1_MSPCONFIGPATH} \
  cli \
  peer chaincode invoke \
    -o orderer.dmc.ajou.ac.kr:7050 \
    -C dmcchannel \
    -n fabcar \
    -c '{"Args":["setEntryLog"]}' \
    --transient "{\"entryLog\":\"$entryLogValue\"}" \
    --waitForEvent \
    --tls \
    --cafile ${ORDERER_TLS_ROOTCERT_FILE} \
    --peerAddresses peer0.org1.dmc.ajou.ac.kr:7051 \
    --tlsRootCertFiles ${ORG1_TLS_ROOTCERT_FILE} \
set +x

cat <<EOF

Total setup execution time : $(($(date +%s) - starttime)) secs ...

EOF
