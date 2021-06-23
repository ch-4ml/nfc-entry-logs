const express = require('express');
const router = express.Router();

const { Wallets, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');

router.get('/', async (req, res) => {

  

  res.status(200).send();
});

router.get('/entry', async (req, res) => {
  try {
    const ccpPath = path.resolve(__dirname, '..', '..', 'first-network', 'connection-org1.json');
    const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

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
  
    const entryLogIndexFile = path.resolve(__dirname, '..', 'modules', 'entryLogIndex.json');
    const entryLogIndex = JSON.parse(fs.readFileSync(entryLogIndexFile, 'utf-8'));
  
    const peopleData = require('../modules/people');
    const personIndex = getRandomInt(0, 5);
    const facilityIndex = getRandomInt(0, 5);
    const person = peopleData[personIndex];
  
    const transientData = {
      entryLogID: `EntryLog${entryLogIndex.next}`,
      facilityID: `Facility${facilityIndex}`,  
      entryTime: tzDate.toISOString().replace(/T/, ' ').replace(/\..+/, ''),
      personalID: `Person${personIndex}`,
      ...person
    }

    console.log(transientData);
  
    const entryLog = Buffer.from(JSON.stringify(transientData)).toString('base64');
  
    // Submit the specified transaction.
    await contract.createTransaction('setEntryLog')
        .setTransient({ entryLog: entryLog })
        .submit();
    console.log('Transaction has been submitted');
  
    entryLogIndex.next += 1;
    fs.writeFileSync(entryLogIndexFile, JSON.stringify(entryLogIndex, null, 2));
  
    await gateway.disconnect();

    res.status(200).send('출입 등록 완료');
  } catch(err) {
    console.error(err);
  }
});

router.get('/entryLog/:personalID', async (req, res) => {
  try {
    const personalID = req.params.personalID;
    const ccpPath = path.resolve(__dirname, '..', '..', 'first-network', 'connection-org1.json');
    const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

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

    const data = JSON.parse(await contract.evaluateTransaction('queryEntryLogsByPersonalID', personalID));
    const privateData = JSON.parse(await contract.evaluateTransaction('getPrivateEntryLogByPerson', personalID));
    let result = [];
    for(let i = 0; i < data.length; i++) {
      result.push(Object.assign(data[i].Record, privateData[i].Record));
    }
    res.status(200).send(result);
    console.log(result);

  } catch (err) {
    console.error(err);
  }
});

function getRandomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);
  return Math.floor(Math.random() * (max - min)) + min;
}

module.exports = router;
