/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke entryLogs, pass private data as base64 encoded bytes in transient map ====
//
// export entryLog=$(echo -n "{\"name\":\"entryLog1\",\"color\":\"blue\",\"size\":35,\"owner\":\"tom\",\"price\":99}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n entryLogsp -c '{"Args":["setEntryLog"]}' --transient "{\"entryLog\":\"$entryLog\"}"
//
// export entryLog=$(echo -n "{\"name\":\"entryLog2\",\"color\":\"red\",\"size\":50,\"owner\":\"tom\",\"price\":102}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n entryLogsp -c '{"Args":["setEntryLog"]}' --transient "{\"entryLog\":\"$entryLog\"}"
//
// export entryLog=$(echo -n "{\"name\":\"entryLog3\",\"color\":\"blue\",\"size\":70,\"owner\":\"tom\",\"price\":103}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n entryLogsp -c '{"Args":["setEntryLog"]}' --transient "{\"entryLog\":\"$entryLog\"}"
//
// export entryLog_OWNER=$(echo -n "{\"name\":\"entryLog2\",\"owner\":\"jerry\"}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n entryLogsp -c '{"Args":["updateAddress"]}' --transient "{\"entryLog_owner\":\"$entryLog_OWNER\"}"
//
// export entryLog_DELETE=$(echo -n "{\"name\":\"entryLog1\"}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n entryLogsp -c '{"Args":["delete"]}' --transient "{\"entryLog_delete\":\"$entryLog_DELETE\"}"

// ==== Query entryLogs, since queries are not recorded on chain we don't need to hide private data in transient map ====
// peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["getEntryLog","entryLog1"]}'
// peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["getEntryLogPrivateDetails","entryLog1"]}'
// peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["queryEntryLogsByPersonalID","entryLog1","entryLog4"]}'
//
// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["queryEntryLogsByFacilityID","tom"]}'
//   peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["queryEntryLogs","{\"selector\":{\"owner\":\"tom\"}}"]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Or for indexes on private data
// collections, in a META-INF/statedb/couchdb/collections/<collection_name>/indexes directory.
// Each index must be defined in its own text file with extension *.json with the index
// definition formatted in JSON following the CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This entryLogs02_private example chaincode demonstrates a packaged index which you
// can find in META-INF/statedb/couchdb/collection/collectionEntryLog/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory or
// META-INF/statedb/couchdb/collections/<collection_name>/indexes directory, for packaging
// and deployment to managed environments.
//
// In the examples below you can find index definitions that support entryLogs02_private
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":["data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}

// Index for docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["queryEntryLogs","{\"selector\":{\"docType\":\"entryLog\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C dmcchannel -n entryLogsp -c '{"Args":["queryEntryLogs","{\"selector\":{\"docType\":{\"$eq\":\"entryLog\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type entryLog struct {
	ObjectType string `json:"docType"`	 	// docType is used to distinguish the various types of objects in state database
	EntryLogID string `json:"entryLogID`	// entryLog1, entryLog2, entryLog3, ...
	FacilityID string `json:"facilityID` 	// the fieldtags are needed to keep case from bouncing around
	Year       string `json:"year"`    
	Sex        string `json:"sex"`
	EntryTime  string `json:"entryTime"`
}

type entryLogPrivateDetails struct {
	ObjectType string `json:"docType"` 		// docType is used to distinguish the various types of objects in state database
	EntryLogID string `json:"entryLogID`	// entryLog1, entryLog2, entryLog3, ...
	PersonalID string `json:"personalID"`   // the fieldtags are needed to keep case from bouncing around
	Name       string `json:"name"`   	
	Phone      string `json:"phone"`
	Address	   string `json:"address"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	switch function {
	case "setEntryLog":
		//create a new entryLog
		return t.setEntryLog(stub, args)
	case "getEntryLog":
		//read a entryLog
		return t.getEntryLog(stub, args)
	case "getEntryLogPrivateDetails":
		//read a entryLog private details
		return t.getEntryLogPrivateDetails(stub, args)
	case "updateAddress":
		//change owner of a specific entryLog
		return t.updateAddress(stub, args)
	case "delete":
		//delete a entryLog
		return t.delete(stub, args)
	case "queryEntryLogsByFacilityID":
		//find entryLogs for owner X using rich query
		return t.queryEntryLogsByFacilityID(stub, args)
	case "queryEntryLogsByPersonalID":
		//get entryLogs based on range query
		return t.queryEntryLogsByPersonalID(stub, args)
	case "queryEntryLogs":
		//find entryLogs based on an ad hoc rich query
		return t.queryEntryLogs(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// ============================================================
// setEntryLog - create a new entryLog, store into chaincode state
// ============================================================
func (t *SimpleChaincode) setEntryLog(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	type entryLogTransientInput struct {
		EntryLogID string `json:"entryLogID`	// entryLog1, entryLog2, entryLog3, ...
		FacilityID string `json:"facilityID` 	// the fieldtags are needed to keep case from bouncing around
		Year       string `json:"year"`    
		Sex        string `json:"sex"`
		EntryTime  string `json:"entryTime"`
	// ***************************************
		PersonalID string `json:"personalID"`   // the fieldtags are needed to keep case from bouncing around
		Name       string `json:"name"`   	
		Phone      string `json:"phone"`
		Address	   string `json:"address"`
	}

	// ==== Input sanitation ====
	fmt.Println("- start init entry log")

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private entry log data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["entryLog"]; !ok {
		return shim.Error("entry log must be a key in the transient map")
	}

	if len(transMap["entryLog"]) == 0 {
		return shim.Error("entry log value in the transient map must be a non-empty JSON string")
	}

	var entryLogInput entryLogTransientInput
	err = json.Unmarshal(transMap["entryLog"], &entryLogInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["entryLog"]))
	}

	if len(entryLogInput.EntryLogID) == 0 {
		return shim.Error("entryLogID field must be a non-empty string")
	}
	if len(entryLogInput.FacilityID) == 0 {
		return shim.Error("facilityID field must be a non-empty string")
	}
	if len(entryLogInput.Year) == 0 {
		return shim.Error("year field must be a non-empty string")
	}
	if len(entryLogInput.Sex) == 0 {
		return shim.Error("sex field must be a non-empty string")
	}
	if len(entryLogInput.EntryTime) == 0 {
		return shim.Error("entryIndex field must be a non-empty string")
	}
	if len(entryLogInput.PersonalID) == 0 {
		return shim.Error("personalID field must be a non-empty string")
	}
	if len(entryLogInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(entryLogInput.Phone) == 0 {
		return shim.Error("phone field must be a non-empty string")
	}
	if len(entryLogInput.Address) == 0 {
		return shim.Error("address field must be a non-empty string")
	}

	// ==== Check if entryLog already exists ====
	entryLogAsBytes, err := stub.GetPrivateData("collectionEntryLog", entryLogInput.EntryLogID)
	if err != nil {
		return shim.Error("Failed to get entry log: " + err.Error())
	} else if entryLogAsBytes != nil {
		fmt.Println("This entry log already exists: " + entryLogInput.EntryLogID)
		return shim.Error("This entry log already exists: " + entryLogInput.EntryLogID)
	}

	// ==== Create entryLog object, marshal to JSON, and save to state ====
	entryLog := &entryLog{
		ObjectType: "entryLog",
		EntryLogID: entryLogInput.EntryLogID,
		FacilityID: entryLogInput.FacilityID,
		Year:		entryLogInput.Year,      
		Sex:		entryLogInput.Sex,      
		EntryTime:	entryLogInput.EntryTime,
	}
	entryLogJSONasBytes, err := json.Marshal(entryLog)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save entryLog to state ===
	err = stub.PutPrivateData("collectionEntryLog", entryLogInput.EntryLogID, entryLogJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Create entryLog private details object with price, marshal to JSON, and save to state ====
	entryLogPrivateDetails := &entryLogPrivateDetails{
		ObjectType: "entryLogPrivateDetails",
		EntryLogID: entryLogInput.EntryLogID,
		PersonalID: entryLogInput.PersonalID,
		Name:		entryLogInput.Name,
		Phone:		entryLogInput.Phone,
		Address:	entryLogInput.Address,
	}
	entryLogPrivateDetailsBytes, err := json.Marshal(entryLogPrivateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutPrivateData("collectionEntryLogPrivateDetails", entryLogInput.EntryLogID, entryLogPrivateDetailsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the entryLog to enable color-based range queries, e.g. return all blue entryLogs ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~facility~entryLog.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "facility~entryLog"
	facilityEntryLogIndexKey, err := stub.CreateCompositeKey(indexName, []string{entryLog.FacilityID, entryLog.EntryLogID})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the entryLog.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutPrivateData("collectionEntryLog", facilityEntryLogIndexKey, value)

	indexName = "personal~entryLog"
	personalEntryLogIndexKey, err := stub.CreateCompositeKey(indexName, []string{entryLogPrivateDetails.PersonalID, entryLogPrivateDetails.EntryLogID})
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutPrivateData("collectionEntryLogPrivateDetails", personalEntryLogIndexKey, value)
	
	// ==== entryLog saved and indexed. Return success ====
	fmt.Println("- end init entryLog")
	return shim.Success(nil)
}

// ===============================================
// getEntryLog - read a entryLog from chaincode state
// ===============================================
func (t *SimpleChaincode) getEntryLog(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var entryLogID, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting entryLogID of the entryLog to query")
	}

	entryLogID = args[0]
	valAsbytes, err := stub.GetPrivateData("collectionEntryLog", entryLogID) //get the entryLog from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + entryLogID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"entryLog does not exist: " + entryLogID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ===============================================
// getEntryLoggetEntryLogPrivateDetails - read a entryLog private details from chaincode state
// ===============================================
func (t *SimpleChaincode) getEntryLogPrivateDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var entryLogID, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting entryLogID of the entryLog to query")
	}

	entryLogID = args[0]
	valAsbytes, err := stub.GetPrivateData("collectionEntryLogPrivateDetails", entryLogID) //get the entryLog private details from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get private details for " + entryLogID + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"entryLog private details does not exist: " + entryLogID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a entryLog key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start delete entryLog")

	type entryLogDeleteTransientInput struct {
		EntryLogID string `json:"entryLogID"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private entryLogID must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["entryLog_delete"]; !ok {
		return shim.Error("entryLog_delete must be a key in the transient map")
	}

	if len(transMap["entryLog_delete"]) == 0 {
		return shim.Error("entryLog_delete value in the transient map must be a non-empty JSON string")
	}

	var entryLogDeleteInput entryLogDeleteTransientInput
	err = json.Unmarshal(transMap["entryLog_delete"], &entryLogDeleteInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["entryLog_delete"]))
	}

	if len(entryLogDeleteInput.EntryLogID) == 0 {
		return shim.Error("entryLogID field must be a non-empty string")
	}

	// to maintain the facility~entryLog index, we need to read the entryLog first and get its color
	valAsbytes, err := stub.GetPrivateData("collectionEntryLog", entryLogDeleteInput.EntryLogID) //get the entryLog from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + entryLogDeleteInput.EntryLogID)
	} else if valAsbytes == nil {
		return shim.Error("entryLog does not exist: " + entryLogDeleteInput.EntryLogID)
	}

	var entryLogToDelete entryLog
	err = json.Unmarshal([]byte(valAsbytes), &entryLogToDelete)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(valAsbytes))
	}

	// delete the entryLog from state
	err = stub.DelPrivateData("collectionEntryLog", entryLogDeleteInput.EntryLogID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Also delete the entryLog from the facility~entryLog index
	indexName := "facility~entryLog"
	facilityEntryLogIndexKey, err := stub.CreateCompositeKey(indexName, []string{entryLogToDelete.FacilityID, entryLogToDelete.EntryLogID})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.DelPrivateData("collectionEntryLog", facilityEntryLogIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

    // *******************************************************************************************************************************************

	// to maintain the facility~entryLog index, we need to read the entryLog first and get its color
	valAsbytes, err = stub.GetPrivateData("collectionEntryLogPrivateDetails", entryLogDeleteInput.EntryLogID) //get the entryLog from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + entryLogDeleteInput.EntryLogID)
	} else if valAsbytes == nil {
		return shim.Error("entryLog does not exist: " + entryLogDeleteInput.EntryLogID)
	}

	var entryLogPrivateDetailsToDelete entryLogPrivateDetails
	err = json.Unmarshal([]byte(valAsbytes), &entryLogPrivateDetailsToDelete)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(valAsbytes))
	}

	indexName = "personal~entryLog"
	personalEntryLogIndexKey, err := stub.CreateCompositeKey(indexName, []string{entryLogPrivateDetailsToDelete.PersonalID, entryLogPrivateDetailsToDelete.EntryLogID})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.DelPrivateData("collectionEntryLogPrivateDetails", personalEntryLogIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Finally, delete private details of entryLog
	err = stub.DelPrivateData("collectionEntryLogPrivateDetails", entryLogDeleteInput.EntryLogID)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ===========================================================
// transfer a entryLog by setting a new owner name on the entryLog
// ===========================================================
func (t *SimpleChaincode) updateAddress(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("- start transfer entryLog")

	type entryLogTransferTransientInput struct {
		EntryLogID  string `json:"entryLogID"`
		Address 	string `json:"address"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private entryLog data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["entryLog_address"]; !ok {
		return shim.Error("entryLog_address must be a key in the transient map")
	}

	if len(transMap["entryLog_address"]) == 0 {
		return shim.Error("entryLog_address value in the transient map must be a non-empty JSON string")
	}

	var entryLogTransferInput entryLogTransferTransientInput
	err = json.Unmarshal(transMap["entryLog_address"], &entryLogTransferInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["entryLog_address"]))
	}

	if len(entryLogTransferInput.EntryLogID) == 0 {
		return shim.Error("entryLogID field must be a non-empty string")
	}
	if len(entryLogTransferInput.Address) == 0 {
		return shim.Error("address field must be a non-empty string")
	}

	entryLogAsBytes, err := stub.GetPrivateData("collectionEntryLogPrivateDetails", entryLogTransferInput.EntryLogID)
	if err != nil {
		return shim.Error("Failed to get entryLog:" + err.Error())
	} else if entryLogAsBytes == nil {
		return shim.Error("entryLog does not exist: " + entryLogTransferInput.EntryLogID)
	}

	entryLogToTransfer := entryLogPrivateDetails{}
	err = json.Unmarshal(entryLogAsBytes, &entryLogToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	entryLogToTransfer.Address = entryLogTransferInput.Address //change the owner

	entryLogJSONasBytes, _ := json.Marshal(entryLogToTransfer)
	err = stub.PutPrivateData("collectionEntryLogPrivateDetails", entryLogToTransfer.EntryLogID, entryLogJSONasBytes) //rewrite the entryLog
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end updateAddress (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// queryEntryLogsByPersonalID performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) queryEntryLogsByPersonalID(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	facilityID := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"entryLog\",\"facilityID\":\"%s\"}}", facilityID)

	queryResults, err := getQueryPrivateDetailsResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryEntryLogsByFacilityID queries for entryLogs based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryEntryLogsByFacilityID(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	facilityID := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"entryLog\",\"facilityID\":\"%s\"}}", facilityID)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryEntryLogs uses a query string to perform a query for entryLogs.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryEntryLogsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryEntryLogs(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetPrivateDataQueryResult("collectionEntryLog", queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func getQueryPrivateDetailsResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryPrivateDetailsResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetPrivateDataQueryResult("collectionEntryLogPrivateDetails", queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryPrivateDetailsResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
