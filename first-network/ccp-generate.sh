#!/bin/bash

function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function json_ccp {
    local PP=$(one_line_pem $1)
    local CP=$(one_line_pem $2)
    if [ $3 -eq 4 ]; then  # 1개만 씀
        sed -e "s/\${ORG}/$3/" \
            -e "s/\${P0PORT}/$4/" \
            -e "s/\${P1PORT}/$5/" \
            -e "s/\${P2PORT}/$6/" \
            -e "s/\${CAPORT}/$7/" \
            -e "s#\${PEERPEM}#$PP#" \
            -e "s#\${CAPEM}#$CP#" \
            ccp-template-3peers.json 
    else
        sed -e "s/\${ORG}/$3/" \
            -e "s/\${P0PORT}/$4/" \
            -e "s/\${CAPORT}/$5/" \
            -e "s#\${PEERPEM}#$PP#" \
            -e "s#\${CAPEM}#$CP#" \
            ccp-template-1peer.json 
    fi
}

function yaml_ccp {
    local PP=$(one_line_pem $1)
    local CP=$(one_line_pem $2)
    if [ $3 -eq 1 ]; then
        sed -e "s/\${ORG}/$3/" \
            -e "s/\${P0PORT}/$4/" \
            -e "s/\${P1PORT}/$5/" \
            -e "s/\${P2PORT}/$6/" \
            -e "s/\${CAPORT}/$7/" \
            -e "s#\${PEERPEM}#$PP#" \
            -e "s#\${CAPEM}#$CP#" \
            ccp-template-3peers.yaml | sed -e $'s/\\\\n/\\\n        /g'
    else
        sed -e "s/\${ORG}/$3/" \
            -e "s/\${P0PORT}/$4/" \
            -e "s/\${CAPORT}/$5/" \
            -e "s#\${PEERPEM}#$PP#" \
            -e "s#\${CAPEM}#$CP#" \
            ccp-template-1peer.yaml | sed -e $'s/\\\\n/\\\n        /g'
    fi
}

ORG=1
P0PORT=7051
P1PORT=8051
P2PORT=9051
CAPORT=7054
PEERPEM=crypto-config/peerOrganizations/org1.dmc.ajou.ac.kr/tlsca/tlsca.org1.dmc.ajou.ac.kr-cert.pem
CAPEM=crypto-config/peerOrganizations/org1.dmc.ajou.ac.kr/ca/ca.org1.dmc.ajou.ac.kr-cert.pem

# echo "$(json_ccp $PEERPEM $CAPEM $ORG $P0PORT $P1PORT $P2PORT $CAPORT)" > connection-org1.json
# echo "$(yaml_ccp $PEERPEM $CAPEM $ORG $P0PORT $P1PORT $P2PORT $CAPORT)" > connection-org1.yaml
echo "$(json_ccp $PEERPEM $CAPEM $ORG $P0PORT $CAPORT)" > connection-org1.json
echo "$(yaml_ccp $PEERPEM $CAPEM $ORG $P0PORT $CAPORT)" > connection-org1.yaml

ORG=2
P0PORT=10051
CAPORT=8054
PEERPEM=crypto-config/peerOrganizations/org2.dmc.ajou.ac.kr/tlsca/tlsca.org2.dmc.ajou.ac.kr-cert.pem
CAPEM=crypto-config/peerOrganizations/org2.dmc.ajou.ac.kr/ca/ca.org2.dmc.ajou.ac.kr-cert.pem

echo "$(json_ccp $PEERPEM $CAPEM $ORG $P0PORT $CAPORT)" > connection-org2.json
echo "$(yaml_ccp $PEERPEM $CAPEM $ORG $P0PORT $CAPORT)" > connection-org2.yaml

ORG=3
P0PORT=11051
CAPORT=9054
PEERPEM=crypto-config/peerOrganizations/org3.dmc.ajou.ac.kr/tlsca/tlsca.org3.dmc.ajou.ac.kr-cert.pem
CAPEM=crypto-config/peerOrganizations/org3.dmc.ajou.ac.kr/ca/ca.org3.dmc.ajou.ac.kr-cert.pem

echo "$(json_ccp $PEERPEM $CAPEM $ORG $P0PORT $CAPORT)" > connection-org3.json
echo "$(yaml_ccp $PEERPEM $CAPEM $ORG $P0PORT $CAPORT)" > connection-org3.yaml
