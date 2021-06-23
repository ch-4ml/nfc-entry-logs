const express = require('express');
const router = express.Router();

const { Wallets, Gateway } = require('fabric-network');
const path = require('path');
const fs = require('fs');

/* GET users listing. */
router.get('/entryLog/:facilityID', async function(req, res, next) {
  try {
    const facilityID = req.params.facilityID;
    const ccpPath = path.resolve(__dirname, '..', '..', 'first-network', 'connection-org2.json');
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

    const data = JSON.parse(await contract.evaluateTransaction('queryEntryLogsByFacilityID', facilityID));
    res.status(200).send(data);
    console.log(data);

  } catch (err) {
    console.error(err);
  }
});

module.exports = router;
