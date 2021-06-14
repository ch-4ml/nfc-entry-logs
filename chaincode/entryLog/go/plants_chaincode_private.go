/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke plants, pass private data as base64 encoded bytes in transient map ====
//
// export PLANT=$(echo -n "{\"name\":\"plant1\",\"color\":\"blue\",\"size\":35,\"owner\":\"tom\",\"price\":99}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n plantsp -c '{"Args":["initPlant"]}' --transient "{\"plant\":\"$PLANT\"}"
//
// export PLANT=$(echo -n "{\"name\":\"plant2\",\"color\":\"red\",\"size\":50,\"owner\":\"tom\",\"price\":102}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n plantsp -c '{"Args":["initPlant"]}' --transient "{\"plant\":\"$PLANT\"}"
//
// export PLANT=$(echo -n "{\"name\":\"plant3\",\"color\":\"blue\",\"size\":70,\"owner\":\"tom\",\"price\":103}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n plantsp -c '{"Args":["initPlant"]}' --transient "{\"plant\":\"$PLANT\"}"
//
// export PLANT_OWNER=$(echo -n "{\"name\":\"plant2\",\"owner\":\"jerry\"}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n plantsp -c '{"Args":["transferPlant"]}' --transient "{\"plant_owner\":\"$PLANT_OWNER\"}"
//
// export PLANT_DELETE=$(echo -n "{\"name\":\"plant1\"}" | base64 | tr -d \\n)
// peer chaincode invoke -C dmcchannel -n plantsp -c '{"Args":["delete"]}' --transient "{\"plant_delete\":\"$PLANT_DELETE\"}"

// ==== Query plants, since queries are not recorded on chain we don't need to hide private data in transient map ====
// peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["readPlant","plant1"]}'
// peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["readPlantPrivateDetails","plant1"]}'
// peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["getPlantsByRange","plant1","plant4"]}'
//
// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["queryPlantsByOwner","tom"]}'
//   peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["queryPlants","{\"selector\":{\"owner\":\"tom\"}}"]}'

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
// This plants02_private example chaincode demonstrates a packaged index which you
// can find in META-INF/statedb/couchdb/collection/collectionPlants/indexes/indexOwner.json.
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
// In the examples below you can find index definitions that support plants02_private
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
//   peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["queryPlants","{\"selector\":{\"docType\":\"plant\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C dmcchannel -n plantsp -c '{"Args":["queryPlants","{\"selector\":{\"docType\":{\"$eq\":\"plant\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

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

type plant struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

type plantPrivateDetails struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Price      int    `json:"price"`
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
	case "initPlant":
		//create a new plant
		return t.initPlant(stub, args)
	case "readPlant":
		//read a plant
		return t.readPlant(stub, args)
	case "readPlantPrivateDetails":
		//read a plant private details
		return t.readPlantPrivateDetails(stub, args)
	case "transferPlant":
		//change owner of a specific plant
		return t.transferPlant(stub, args)
	case "delete":
		//delete a plant
		return t.delete(stub, args)
	case "queryPlantsByOwner":
		//find plants for owner X using rich query
		return t.queryPlantsByOwner(stub, args)
	case "queryPlants":
		//find plants based on an ad hoc rich query
		return t.queryPlants(stub, args)
	case "getPlantsByRange":
		//get plants based on range query
		return t.getPlantsByRange(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// ============================================================
// initPlant - create a new plant, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initPlant(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	type plantTransientInput struct {
		Name  string `json:"name"` //the fieldtags are needed to keep case from bouncing around
		Color string `json:"color"`
		Size  int    `json:"size"`
		Owner string `json:"owner"`
		Price int    `json:"price"`
	}

	// ==== Input sanitation ====
	fmt.Println("- start init plant")

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private plant data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["plant"]; !ok {
		return shim.Error("plant must be a key in the transient map")
	}

	if len(transMap["plant"]) == 0 {
		return shim.Error("plant value in the transient map must be a non-empty JSON string")
	}

	var plantInput plantTransientInput
	err = json.Unmarshal(transMap["plant"], &plantInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["plant"]))
	}

	if len(plantInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(plantInput.Color) == 0 {
		return shim.Error("color field must be a non-empty string")
	}
	if plantInput.Size <= 0 {
		return shim.Error("size field must be a positive integer")
	}
	if len(plantInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}
	if plantInput.Price <= 0 {
		return shim.Error("price field must be a positive integer")
	}

	// ==== Check if plant already exists ====
	plantAsBytes, err := stub.GetPrivateData("collectionPlants", plantInput.Name)
	if err != nil {
		return shim.Error("Failed to get plant: " + err.Error())
	} else if plantAsBytes != nil {
		fmt.Println("This plant already exists: " + plantInput.Name)
		return shim.Error("This plant already exists: " + plantInput.Name)
	}

	// ==== Create plant object, marshal to JSON, and save to state ====
	plant := &plant{
		ObjectType: "plant",
		Name:       plantInput.Name,
		Color:      plantInput.Color,
		Size:       plantInput.Size,
		Owner:      plantInput.Owner,
	}
	plantJSONasBytes, err := json.Marshal(plant)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save plant to state ===
	err = stub.PutPrivateData("collectionPlants", plantInput.Name, plantJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Create plant private details object with price, marshal to JSON, and save to state ====
	plantPrivateDetails := &plantPrivateDetails{
		ObjectType: "plantPrivateDetails",
		Name:       plantInput.Name,
		Price:      plantInput.Price,
	}
	plantPrivateDetailsBytes, err := json.Marshal(plantPrivateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutPrivateData("collectionPlantPrivateDetails", plantInput.Name, plantPrivateDetailsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the plant to enable color-based range queries, e.g. return all blue plants ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{plant.Color, plant.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the plant.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutPrivateData("collectionPlants", colorNameIndexKey, value)

	// ==== Plant saved and indexed. Return success ====
	fmt.Println("- end init plant")
	return shim.Success(nil)
}

// ===============================================
// readPlant - read a plant from chaincode state
// ===============================================
func (t *SimpleChaincode) readPlant(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the plant to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetPrivateData("collectionPlants", name) //get the plant from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Plant does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ===============================================
// readPlantreadPlantPrivateDetails - read a plant private details from chaincode state
// ===============================================
func (t *SimpleChaincode) readPlantPrivateDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the plant to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetPrivateData("collectionPlantPrivateDetails", name) //get the plant private details from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get private details for " + name + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Plant private details does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a plant key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start delete plant")

	type plantDeleteTransientInput struct {
		Name string `json:"name"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private plant name must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["plant_delete"]; !ok {
		return shim.Error("plant_delete must be a key in the transient map")
	}

	if len(transMap["plant_delete"]) == 0 {
		return shim.Error("plant_delete value in the transient map must be a non-empty JSON string")
	}

	var plantDeleteInput plantDeleteTransientInput
	err = json.Unmarshal(transMap["plant_delete"], &plantDeleteInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["plant_delete"]))
	}

	if len(plantDeleteInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}

	// to maintain the color~name index, we need to read the plant first and get its color
	valAsbytes, err := stub.GetPrivateData("collectionPlants", plantDeleteInput.Name) //get the plant from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + plantDeleteInput.Name)
	} else if valAsbytes == nil {
		return shim.Error("Plant does not exist: " + plantDeleteInput.Name)
	}

	var plantToDelete plant
	err = json.Unmarshal([]byte(valAsbytes), &plantToDelete)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(valAsbytes))
	}

	// delete the plant from state
	err = stub.DelPrivateData("collectionPlants", plantDeleteInput.Name)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Also delete the plant from the color~name index
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{plantToDelete.Color, plantToDelete.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.DelPrivateData("collectionPlants", colorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Finally, delete private details of plant
	err = stub.DelPrivateData("collectionPlantPrivateDetails", plantDeleteInput.Name)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ===========================================================
// transfer a plant by setting a new owner name on the plant
// ===========================================================
func (t *SimpleChaincode) transferPlant(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("- start transfer plant")

	type plantTransferTransientInput struct {
		Name  string `json:"name"`
		Owner string `json:"owner"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private plant data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["plant_owner"]; !ok {
		return shim.Error("plant_owner must be a key in the transient map")
	}

	if len(transMap["plant_owner"]) == 0 {
		return shim.Error("plant_owner value in the transient map must be a non-empty JSON string")
	}

	var plantTransferInput plantTransferTransientInput
	err = json.Unmarshal(transMap["plant_owner"], &plantTransferInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["plant_owner"]))
	}

	if len(plantTransferInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(plantTransferInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}

	plantAsBytes, err := stub.GetPrivateData("collectionPlants", plantTransferInput.Name)
	if err != nil {
		return shim.Error("Failed to get plant:" + err.Error())
	} else if plantAsBytes == nil {
		return shim.Error("Plant does not exist: " + plantTransferInput.Name)
	}

	plantToTransfer := plant{}
	err = json.Unmarshal(plantAsBytes, &plantToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	plantToTransfer.Owner = plantTransferInput.Owner //change the owner

	plantJSONasBytes, _ := json.Marshal(plantToTransfer)
	err = stub.PutPrivateData("collectionPlants", plantToTransfer.Name, plantJSONasBytes) //rewrite the plant
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferPlant (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// getPlantsByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getPlantsByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetPrivateDataByRange("collectionPlants", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- getPlantsByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
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
// queryPlantsByOwner queries for Plants based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryPlantsByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"plant\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryPlants uses a query string to perform a query for plants.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryPlantsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryPlants(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

	resultsIterator, err := stub.GetPrivateDataQueryResult("collectionPlants", queryString)
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
