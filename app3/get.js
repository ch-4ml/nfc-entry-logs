/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Wallets, Gateway } = require('fabric-network');
const path = require('path');
const fs = require('fs');

const ccpPath = path.resolve(__dirname, '..', 'first-network', 'connection-org3.json');
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

        // Evaluate the specified transaction.
        const resultGetEntryLog = await contract.evaluateTransaction('getEntryLog', 'EntryLog4');
        const resultGetEntryLogPD = await contract.evaluateTransaction('getEntryLogPrivateDetails', 'EntryLog4')
        const result = { ...JSON.parse(resultGetEntryLog), ...JSON.parse(resultGetEntryLogPD) };
        console.log(`Transaction has been evaluated, result is: ${JSON.stringify(result)}`);

        process.exit(0);
    } catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
        process.exit(1);
    }
}

main();
