/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

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
	PersonalID string `json:"personalID"`   // the fieldtags are needed to keep case from bouncing around
	Year       string `json:"year"`    
	Gender     string `json:"gender"`
	EntryTime  string `json:"entryTime"`
}

type entryLogPrivateDetails struct {
	ObjectType string `json:"docType"` 		// docType is used to distinguish the various types of objects in state database
	EntryLogID string `json:"entryLogID`	// entryLog1, entryLog2, entryLog3, ...
	FacilityID string `json:"facilityID` 	// the fieldtags are needed to keep case from bouncing around
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
	case "getPrivateEntryLogByFacility":
		return t.getPrivateEntryLogByFacility(stub, args)
	case "getPrivateEntryLogByPerson":
		return t.getPrivateEntryLogByPerson(stub, args)
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
		Gender     string `json:"gender"`
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
	if len(entryLogInput.Gender) == 0 {
		return shim.Error("gender field must be a non-empty string")
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
		PersonalID: entryLogInput.PersonalID,
		Year:		entryLogInput.Year,      
		Gender:		entryLogInput.Gender,      
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
		FacilityID: entryLogInput.FacilityID,
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

	indexName := "facility~entryLog"
	facilityEntryLogIndexKey, err := stub.CreateCompositeKey(indexName, []string{entryLogPrivateDetails.FacilityID, entryLogPrivateDetails.EntryLogID})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the plant.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutPrivateData("collectionEntryLogPrivateDetails", facilityEntryLogIndexKey, value)

	indexName = "personal~entryLog"
	personalEntryLogIndexKey, err := stub.CreateCompositeKey(indexName, []string{entryLogPrivateDetails.PersonalID, entryLogPrivateDetails.EntryLogID})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the plant.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
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
	valAsBytes, err := stub.GetPrivateData("collectionEntryLog", entryLogID) //get the entryLog from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + entryLogID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp = "{\"Error\":\"entryLog does not exist: " + entryLogID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsBytes)
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
	valAsBytes, err := stub.GetPrivateData("collectionEntryLogPrivateDetails", entryLogID) //get the entryLog private details from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get private details for " + entryLogID + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp = "{\"Error\":\"entryLog private details does not exist: " + entryLogID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsBytes)
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

	// delete the entryLog from state
	err = stub.DelPrivateData("collectionEntryLog", entryLogDeleteInput.EntryLogID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

    // *******************************************************************************************************************************************

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

	personalID := args[0]

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"entryLog\",\"personalID\":\"%s\"}}", personalID)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
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

	facilityID := args[0]

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"entryLog\",\"FacilityID\":\"%s\"}}", facilityID)
	fmt.Println(queryString);

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	fmt.Println(queryResults);
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
		res, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(res.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(res.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getPrivateEntryLogByFacility(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	facilityID := args[0]
	indexKey := "facility~entryLog"

	results, err := getEntryLogPrivateDetailsByCompositeKey(stub, facilityID, indexKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(results)
}

func (t *SimpleChaincode) getPrivateEntryLogByPerson(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	personalID := args[0]
	indexKey := "personal~entryLog"

	results, err := getEntryLogPrivateDetailsByCompositeKey(stub, personalID, indexKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(results)
}

func getEntryLogPrivateDetailsByCompositeKey(stub shim.ChaincodeStubInterface, key string, indexKey string) ([]byte, error) {
	resultsIterator, err := stub.GetPrivateDataByPartialCompositeKey("collectionEntryLogPrivateDetails", indexKey, []string{key})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		res, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(res.Key)
		if err != nil {
			return nil, err
		}

		returnedID := compositeKeyParts[1]

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(returnedID)
		buffer.WriteString("\"")

		valAsBytes, err := stub.GetPrivateData("collectionEntryLogPrivateDetails", returnedID) //get the entryLog private details from chaincode state
		if err != nil {
			return nil, err
		}

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(valAsBytes))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- Result:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

