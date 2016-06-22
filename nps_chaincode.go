/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var marbleIndexStr = "_marbleindex"				//name for the key/value that will store a list of all known marbles

type Marble struct{
	Id string `json:"Id"`					//the fieldtags are needed to keep case from bouncing around
	Survey_Id string `json:"Survey_Id"`
        Customer_Id string `json:"Customer_Id"`
	Score int `json:"Score"`
	Feedback string `json:"Feedback"`
        Submitted_Date string `json:"Submitted_Date"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(marbleIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}		
	
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_marble" {									//create a new marble
		return t.init_marble(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var Id, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting Id of the var to query")
	}

	Id = args[0]
	valAsbytes, err := stub.GetState(Id)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + Id + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var Id, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. Id of the variable and value to set")
	}

	Id = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(Id, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_marble(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0          1            2              3        4           5
	// "object id", "survey id", "customer id", "score", "feedback", "submitted date"
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	//input sanitation
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
        if len(args[4]) <= 0 {
		return nil, errors.New("5th argument must be a non-empty string")
	}
        if len(args[5]) <= 0 {
		return nil, errors.New("6th argument must be a non-empty string")
	}
	Id := args[0]
	Survey_Id := strings.ToLower(args[1])
        Customer_Id := strings.ToLower(args[2])
        Score, err := strconv.Atoi(args[3])
	Feedback := strings.ToLower(args[4])
        Created_Date := strings.ToLower(args[5])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	//check if marble already exists
	marbleAsBytes, err := stub.GetState(Id)
	if err != nil {
		return nil, errors.New("Failed to get marble Id")
	}
	res := Marble{}
	json.Unmarshal(marbleAsBytes, &res)
	if res.Id == Id{
		fmt.Println("This marble arleady exists: " + Id)
		fmt.Println(res);
		return nil, errors.New("This marble arleady exists")				//all stop a marble by this Id exists
	}
	
	//build the marble json string manually
	str := `{"Id": "` + Id + `", "Survey_Id": "` + Survey_Id + `", "Customer_Id": "` + Customer_Id + `", "Score": ` + strconv.Itoa(Score) + `, "Feedback": "` + Feedback + `", "Submitted_Date": "` + Submitted_Date + `"}`
	err = stub.PutState(Id, []byte(str))									//store marble with id as key
	if err != nil {
		return nil, err
	}
		
	//get the marble index
	marblesAsBytes, err := stub.GetState(marbleIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var marbleIndex []string
	json.Unmarshal(marblesAsBytes, &marbleIndex)							//un stringify it aka JSON.parse()
	
	//append
	marbleIndex = append(marbleIndex, Id)									//add marble Id to index list
	fmt.Println("! marble index: ", marbleIndex)
	jsonAsBytes, _ := json.Marshal(marbleIndex)
	err = stub.PutState(marbleIndexStr, jsonAsBytes)						//store Id of marble

	fmt.Println("- end init marble")
	return nil, nil
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

