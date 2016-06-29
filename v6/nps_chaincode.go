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

var surveyIndexStr = "_surveyindex"				//name for the key/value that will store a list of all known surveys

type Survey struct{
	Key string `json:"key"`					//the fieldtags are needed to keep case from bouncing around
	Survey string `json:"survey"`
	Customer string `json:"customer"`
        Score string `json:"score"`
        Feedback string `json:"feedback"`        
        SubmittedDate string `json:"submitteddate"`
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
	err = stub.PutState(surveyIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	var trades AllTrades
	jsonAsBytes, _ = json.Marshal(trades)								//clear the open trade struct
	err = stub.PutState(openTradesStr, jsonAsBytes)
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
	} else if function == "delete" {										//deletes an entity from its state
		res, err := t.Delete(stub, args)
		cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "read" {											//writes a value to the chaincode state
		return t.read(stub, args)
	} else if function == "init_survey" {									//create a new survey
		return t.init_survey(stub, args)
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
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting key of the var to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	key := args[0]
	err := stub.DelState(key)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the survey index
	surveysAsBytes, err := stub.GetState(surveyIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get survey index")
	}
	var surveyIndex []string
	json.Unmarshal(surveysAsBytes, &surveyIndex)								//un stringify it aka JSON.parse()
	
	//remove survey from index
	for i,val := range surveyIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + key)
		if val == key{															//find the correct survey
			fmt.Println("found survey")
			surveyIndex = append(surveyIndex[:i], surveyIndex[i+1:]...)			//remove it
			for x:= range surveyIndex{											//debug prints...
				fmt.Println(string(x) + " - " + surveyIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(surveyIndex)									//save new index
	err = stub.PutState(surveyIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var key, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	key = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Marble - create a new survey, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_survey(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	//input sanitation
	fmt.Println("- start init survey")
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
	key := args[0]
	survey := strings.ToLower(args[1])
	customer := strings.ToLower(args[2])
	score, err := strconv.Atoi(args[3])
        feedback := strings.ToLower(args[4])
        submitteddate := strings.ToLower(args[5])

	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	//check if survey already exists
	surveyAsBytes, err := stub.GetState(key)
	if err != nil {
		return nil, errors.New("Failed to get survey key")
	}
	res := Survey{}
	json.Unmarshal(surveyAsBytes, &res)
	if res.Key == key{
		fmt.Println("This survey arleady exists: " + key)
		fmt.Println(res);
		return nil, errors.New("This survey arleady exists")				//all stop a survey by this key exists
	}
	
	//build the survey json string manually
	str := `{"key": "` + key + `", "survey": "` + survey + `", "customer": "` + customer + `", "score": ` + strconv.Itoa(score) + `, "feedback": "` + feedback + `", "submitteddate": "` + submitteddate + `"}`
	err = stub.PutState(key, []byte(str))									//store survey with id as key
	if err != nil {
		return nil, err
	}
		
	//get the survey index
	surveysAsBytes, err := stub.GetState(surveyIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get survey index")
	}
	var surveyIndex []string
	json.Unmarshal(surveysAsBytes, &surveyIndex)							//un stringify it aka JSON.parse()
	
	//append
	surveyIndex = append(surveyIndex, key)									//add survey key to index list
	fmt.Println("! survey index: ", surveyIndex)
	jsonAsBytes, _ := json.Marshal(surveyIndex)
	err = stub.PutState(surveyIndexStr, jsonAsBytes)						//store key of survey

	fmt.Println("- end init survey")
	return nil, nil
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

