/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Wallets, Gateway } = require('fabric-network');
const path = require('path');
const fs = require('fs');

const ccpPath = path.resolve(__dirname, '..', 'first-network', 'connection-org1.json');
const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

async function main() {
    try {

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userExists = await wallet.get('user1');
        if (!userExists) {
            console.log('An identity for the user "user1" does not exist in the wallet');
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: 'user1', discovery: { enabled: true, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('dmcchannel');

        // Get the contract from the network.
        const contract = network.getContract('entryLog');
        
        const tzOffset = new Date().getTimezoneOffset() * 60000;
        const tzDate = new Date(Date.now() - tzOffset);

        const entryLogIndexFile = path.resolve(__dirname, 'entryLogIndex.json');
        const entryLogIndex = JSON.parse(fs.readFileSync(entryLogIndexFile, 'utf-8'));

        const transientData = {
            entryLogID: `EntryLog${entryLogIndex.next}`,
            facilityID: `Facility1`,  // NFC로 등록해놓고 입력받기
            year: '1995',
            sex: '1',
            entryTime: tzDate.toISOString().replace(/T/, ' ').replace(/\..+/, ''),
            personalID: `Person1`,
            name: '박찬형',
            phone: '010-6223-2277',
            address: '경기도 수원시'
        }

        const entryLog = Buffer.from(JSON.stringify(transientData)).toString('base64');

        // Submit the specified transaction.
        await contract.createTransaction('setEntryLog')
            .setTransient({ entryLog: entryLog })
            .submit();
        console.log('Transaction has been submitted');

        entryLogIndex.next += 1;
        fs.writeFileSync(entryLogIndexFile, JSON.stringify(entryLogIndex, null, 2));

        // Disconnect from the gateway.
        await gateway.disconnect();

        process.exit(0);
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
