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

        "github.com/hyperledger/fabric/core/chaincode/shim"
)

// LoadChaincode example simple Chaincode implementation
type LoadChaincode struct {
}

var loadNumberIndexStr = "_loadindex"                           //name for the key/value that will store a list of all known loads

type Load struct{
        LoadNumber string `json:"name"`                                 //the fieldtags are needed to keep case from bouncing around
        CarrierName string `json:"carrier"`
        ShipDate string `json:"shipdate"`
        DeliveryDate string `json:"deliverydate"`
        Status string `json:"status"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
        err := shim.Start(new(LoadChaincode))
        if err != nil {
                fmt.Printf("Error starting Simple chaincode: %s", err)
        }
}

// ============================================================================================================================
func (t *LoadChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
        var err error

        var empty []string
        jsonAsBytes, _ := json.Marshal(empty)                                                           //marshal an emtpy array of strings to clear the index
        err = stub.PutState(loadNumberIndexStr, jsonAsBytes)
        if err != nil {
                return nil, err
        }

        return nil, nil
}


// ============================================================================================================================
// invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *LoadChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
        fmt.Println("invoke is running " + function)

        // Handle different functions
        if function == "init" {
                return t.Init(stub, function, args)
        } else if function == "delete" {                                                                                //deletes an entity from its state
                res, err := t.Delete(stub, args)
                return res, err
        } else if function == "write" {                                                                                 //writes a value to the chaincode state
                return t.Write(stub, args)
        } else if function == "init_load" {                                                                     //create a new marble
                return t.init_load(stub, args)
        }
        fmt.Println("invoke did not find func: " + function)                                    //error

        return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *LoadChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
        fmt.Println("query is running " + function)

        // Handle different functions
        if function == "read" {                                                                                                 //read a variable
                return t.read(stub, args)
        }else if function == "show_all" {
	        	return t.read_all(stub)
        }
        fmt.Println("query did not find func: " + function)                                             //error

        return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *LoadChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
        var name, jsonResp string
        var err error

        name = args[0]
        valAsbytes, err := stub.GetState(name)                                                                  //get the var from chaincode state
        if err != nil {
                jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
                return nil, errors.New(jsonResp)
        }


        return valAsbytes, nil                                                                                                  //send it onward
}

func (t *LoadChaincode) read_all(stub shim.ChaincodeStubInterface) ([]byte, error) {
        var err error

		//get the loads index
		loadsAsBytes, err := stub.GetState(loadNumberIndexStr)
		if err != nil {
			return nil, errors.New("Failed to get load index")
		}
		var loadIndex []string
		json.Unmarshal(loadsAsBytes, &loadIndex)								//un stringify it aka JSON.parse()
		
		//remove load from index
		jsonResponse := "" 
		jsonResponse +=  "{ \"LoadNumber\":\""
		for i,val := range loadIndex {
			fmt.Println(strconv.Itoa(i) + " - looking at " + val )
			if i == 0 {
				jsonResponse += val
			}else {	
				jsonResponse += "," + val  
			}
		}
		jsonResponse +=  "\"}"
		fmt.Printf("Query Response:%s\n", jsonResponse)

	bs := []byte(jsonResponse) 
        return bs, nil                                                                                                  //send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
func (t *LoadChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {        
        if len(args) != 1 {
                return nil, errors.New("Incorrect number of arguments. Expecting 1")
        }

        loadNumber := args[0]
       err := stub.DelState(loadNumber) 
	 if err != nil {
                return nil, errors.New("Failed to delete state")
        }

        //get the marble index
        loadsAsBytes, err := stub.GetState(loadNumberIndexStr)
        if err != nil {
                return nil, errors.New("Failed to get load index")
        }
        var loadIndex []string
        json.Unmarshal(loadsAsBytes, &loadIndex)                                                                //un stringify it aka JSON.parse()

        //remove load from index
        for i,val := range loadIndex{
                fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + loadNumber)
                        fmt.Println("found load")
                        loadIndex = append(loadIndex[:i], loadIndex[i+1:]...)                   //remove it
                        for x:= range loadIndex{                                                                                        //debug prints...
                                fmt.Println(string(x) + " - " + loadIndex[x])
                        }
                        break
                }
       
	jsonAsBytes, _ := json.Marshal(loadIndex)
	 err = stub.PutState(loadNumberIndexStr, jsonAsBytes)
        return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *LoadChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

        // 0            loadnumber
        // 1            status
        // 2            carrier
        // 3            date

        var loadNumber, status, carrier, date string // Entities
        var err error
        fmt.Println("running write()")

        if len(args) != 4 {
                return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
        }
	loadNumber = args[0]
        status = args[1]
        carrier = args[2]
	date = args[0]


        //get the load
        loadsAsBytes, err := stub.GetState(loadNumber)
        if err != nil {
                return nil, errors.New("Failed to get load index")
        }

        var receivedLoadStruct Load
        json.Unmarshal(loadsAsBytes, &receivedLoadStruct)

        receivedLoadStruct.Status = status
        receivedLoadStruct.CarrierName = carrier

        if receivedLoadStruct.Status == "Delivered"{
                receivedLoadStruct.DeliveryDate = date
        }else{
                receivedLoadStruct.ShipDate = date
        }

        receivedLoadStructBytes, err := json.Marshal(receivedLoadStruct)
	err = stub.PutState(loadNumber, receivedLoadStructBytes)

        if err != nil {
                return nil, err
        }
        return nil, nil
}

// ============================================================================================================================
// Init load - create a new load, store into chaincode state
// ============================================================================================================================
func (t *LoadChaincode) init_load(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
        var err error
        fmt.Println("Invoke is running init_load" )

        //   0       1       2     3
        // "asdf", "blue", "35", "bob"
        // 0            loadnumber
        // 1            status
        // 2            carrier
        // 3            shipdate
        // 4            deliverydate

        if len(args) != 5 {
                return nil, errors.New("Incorrect number of arguments. Expecting 4")
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

        loadNumber := args[0]
        carrier := args[1]
        shipDate := args[2]
        deliveryDate := args[3]
        status := args[4]


        //check if marble already exists
        loadAsBytes, err := stub.GetState(loadNumber)
        if err != nil {
                return nil, errors.New("Failed to get load number")
        }
        res := Load{}
        json.Unmarshal(loadAsBytes, &res)
        if res.LoadNumber == loadNumber{
                fmt.Println("This Load arleady exists: " + loadNumber)
                fmt.Println(res);
                return nil, errors.New("This Load arleady exists")                              //all stop a marble by this name exists
        }

        loadStruct := Load{}
        loadStruct.LoadNumber = loadNumber
        loadStruct.CarrierName = carrier
        loadStruct.ShipDate = shipDate
        loadStruct.DeliveryDate = deliveryDate
        loadStruct.Status = status

        loadJsonAsBytes, _ := json.Marshal(loadStruct)

        err = stub.PutState(loadNumber, loadJsonAsBytes)                                                                        //store marble with id as key
        if err != nil {
                return nil, err
        }

        //get the marble index
        loadsAsBytes, err := stub.GetState(loadNumberIndexStr)
        if err != nil {
                return nil, errors.New("Failed to get load index")
        }

        var loadIndex []string
        json.Unmarshal(loadsAsBytes, &loadIndex)                                                        //un stringify it aka JSON.parse()

        //append
        loadIndex = append(loadIndex, loadNumber)                                                                       //add marble name to index list
        fmt.Println("! load index: ", loadIndex)
        jsonAsBytes, _ := json.Marshal(loadIndex)
        err = stub.PutState(loadNumberIndexStr, jsonAsBytes)                                            //store name of marble

        fmt.Println("- end init load")
        return nil, nil
}


