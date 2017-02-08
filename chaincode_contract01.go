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

// ContractChaincode example simple Chaincode implementation
type ContractChaincode struct {
}

var contractNumberIndexStr = "_contractindex"                           //name for the key/value that will store a list of all known loads

type Load struct{
        ContractNumber string `json:"contractNumber"`                                 //the fieldtags are needed to keep case from bouncing around
        CarrierName string `json:"carrier"`
        Origin string `json:"origin"`
        Destination string `json:"destination"`
        Service string `json:"service"`
        EquipmentType string `json:"equipmentType"`
        BaseRate string `json:"baseRate"`
        AccessorialRate string `json:"accessorialRate"`
        TenderExpiry string `json:"tenderExpiry"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
        err := shim.Start(new(ContractChaincode))
        if err != nil {
                fmt.Printf("Error starting Simple chaincode: %s", err)
        }
}

// ============================================================================================================================
func (t *ContractChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
        var err error

        var empty []string
        jsonAsBytes, _ := json.Marshal(empty)    //marshal an emtpy array of strings to clear the index
        err = stub.PutState(contractNumberIndexStr, jsonAsBytes)
        if err != nil {
                return nil, err
        }

        return nil, nil
}


// ============================================================================================================================
// invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *ContractChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
        fmt.Println("invoke is running " + function)

        // Handle different functions
        if function == "init" {
                return t.Init(stub, function, args)
        } else if function == "delete" {            //deletes an entity from its state
                res, err := t.Delete(stub, args)
                return res, err
        } else if function == "write" {        //writes a value to the chaincode state
                return t.Write(stub, args)
        } else if function == "init_load" {      //create a new marble
                return t.init_load(stub, args)
        }
        fmt.Println("invoke did not find func: " + function)       //error

        return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *ContractChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
        fmt.Println("query is running " + function)

        // Handle different functions
        if function == "read" {          //read one load information
                return t.read(stub, args)
        }else if function == "read_all" {  // read all loads
	        	return t.read_all(stub)
        }
        fmt.Println("query did not find func: " + function)                                             //error

        return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read one load from chaincode state
// ============================================================================================================================
func (t *ContractChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
        var loadNumber, jsonResp string
        var err error

        loadNumber = args[0]
        valAsbytes, err := stub.GetState(loadNumber)    //get the var from chaincode state
        if err != nil {
                jsonResp = "{\"Error\":\"Failed to get state for " + loadNumber + "\"}"
                return nil, errors.New(jsonResp)
        }


        return valAsbytes, nil    //send it onward
}
// ============================================================================================================================
// Read - read all loads from chaincode state
// ============================================================================================================================
func (t *ContractChaincode) read_all(stub shim.ChaincodeStubInterface) ([]byte, error) {
        var err error

		//get the loads index
		loadsAsBytes, err := stub.GetState(contractNumberIndexStr)
		if err != nil {
			return nil, errors.New("Failed to get load index")
		}
		var loadIndex []string
		json.Unmarshal(loadsAsBytes, &loadIndex)	//un stringify it aka JSON.parse()
		
		//remove load from index
		jsonResponse := "" 
		jsonResponse +=  "{ \"contracts\":["
		for i,val := range loadIndex {
			fmt.Println(strconv.Itoa(i) + " - looking at " + val )
			
			valAsbytes, err := stub.GetState(val)    //get the loadnumber from chaincode state
	        if err != nil {
	                jsonResp := "{\"Error\":\"Failed to get state for " + val + "\"}"
	                return nil, errors.New(jsonResp)
	        }
			
			if i == 0 {
				jsonResponse += string(valAsbytes)
			}else {	
				if len(valAsbytes) >0 {
					jsonResponse += "," + string(valAsbytes) 
				} 
			}
		}
		jsonResponse +=  "]}"
		fmt.Printf("Query Response:%s\n", jsonResponse)

		bs := []byte(jsonResponse) 
        return bs, nil                                                                                                  //send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
func (t *ContractChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {        
        if len(args) != 1 {
                return nil, errors.New("Incorrect number of arguments. Expecting 1")
        }

        loadNumber := args[0]
       err := stub.DelState(loadNumber) 
	 if err != nil {
                return nil, errors.New("Failed to delete state")
        }

        //get the marble index
        loadsAsBytes, err := stub.GetState(contractNumberIndexStr)
        if err != nil {
                return nil, errors.New("Failed to get load index")
        }
        var loadIndex []string
        json.Unmarshal(loadsAsBytes, &loadIndex)     //un stringify it aka JSON.parse()

        //remove load from index
        for i,val := range loadIndex{
                fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + loadNumber)
                        fmt.Println("found load")
                        loadIndex = append(loadIndex[:i], loadIndex[i+1:]...)                   //remove it
                        for x:= range loadIndex{           //debug prints...
                                fmt.Println(string(x) + " - " + loadIndex[x])
                        }
                        break
                }
       
	jsonAsBytes, _ := json.Marshal(loadIndex)
	 err = stub.PutState(contractNumberIndexStr, jsonAsBytes)
        return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *ContractChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

		//1. contractnumber
		//2. carrier
		//3. origin
		//4. destination
		//5. service
		//6. equipmentType
		//7. baseRate
		//8. acessorialRate
		//9. tenderExpiry 

        var contractNumber, carrier, origin, destination, service, equipmentType, baseRate, accessorialRate, tenderExpiry  string // Entities
        var err error
        fmt.Println("running write()")

        if len(args) < 9 {
                return nil, errors.New("Incorrect number of arguments. Expecting atleast 9 of the variable and value to set")
        }
        
		contractNumber = args[0]
		carrier = args[1]
		origin = args[2]
        destination = args[3]
		service = args[4]
		equipmentType  = args[5]
		baseRate  = args[6]
		accessorialRate  = args[7]
		tenderExpiry  = args[8]

        //get the load
        loadsAsBytes, err := stub.GetState(contractNumber)
        if err != nil {
                return nil, errors.New("Contract number :" + contractNumber + " does not exists")
        }

        var receivedLoadStruct Load
        json.Unmarshal(loadsAsBytes, &receivedLoadStruct)

        receivedLoadStruct.CarrierName = carrier
        receivedLoadStruct.Origin = origin
        receivedLoadStruct.Destination = destination
        receivedLoadStruct.Service = service
        receivedLoadStruct.EquipmentType = equipmentType
        receivedLoadStruct.BaseRate = baseRate
        receivedLoadStruct.AccessorialRate = accessorialRate
        receivedLoadStruct.TenderExpiry = tenderExpiry

        receivedLoadStructBytes, err := json.Marshal(receivedLoadStruct)
		err = stub.PutState(contractNumber, receivedLoadStructBytes)

        if err != nil {
                return nil, err
        }
        fmt.Println("write() completed")
        
        return nil, nil
}

// ============================================================================================================================
// Init load - create a new load, store into chaincode state
// ============================================================================================================================
func (t *ContractChaincode) init_load(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
        var err error
        fmt.Println("Invoke is running init_load" )
        
        //1. contractnumber
		//2. carrier
		//3. origin
		//4. destination
		//5. service
		//6. equipmentType
		//7. baseRate
		//8. acessorialRate
		//9. tenderExpiry 

        var contractNumber, carrier, origin, destination, service, equipmentType, baseRate, accessorialRate, tenderExpiry  string // Entities

        if len(args) < 9 {
                return nil, errors.New("Incorrect number of arguments. Expecting 10")
        }

        //input sanitation
        if len(args[0]) <= 0 {
                return nil, errors.New("1st argument must be a non-empty string")
        }
        if len(args[1]) <= 0 {
                return nil, errors.New("2nd argument must be a non-empty string")
        }
        if len(args[2]) <= 0 {
                return nil, errors.New("3rd argument must be a non-empty string")
        }

        contractNumber = args[0]
		carrier = args[1]
		origin = args[2]
        destination = args[3]
		service = args[4]
		equipmentType  = args[5]
		baseRate  = args[6]
		accessorialRate  = args[7]
		tenderExpiry  = args[8]


        //check if contractNumber already exists
        loadAsBytes, err := stub.GetState(contractNumber)
        if err != nil {
                return nil, errors.New("Failed to get load number")
        }
        res := Load{}
        json.Unmarshal(loadAsBytes, &res)
        if res.ContractNumber == contractNumber{
                fmt.Println("This contractNumber arleady exists: " + contractNumber)
                fmt.Println(res);
                return nil, errors.New("This Load arleady exists")      
        }

        loadStruct := Load{}
        loadStruct.ContractNumber = contractNumber
	loadStruct.CarrierName = carrier
        loadStruct.Origin = origin
        loadStruct.Destination = destination
        loadStruct.Service = service
        loadStruct.EquipmentType = equipmentType
        loadStruct.BaseRate = baseRate
        loadStruct.AccessorialRate = accessorialRate
        loadStruct.TenderExpiry = tenderExpiry
        

        loadJsonAsBytes, _ := json.Marshal(loadStruct)

        err = stub.PutState(contractNumber, loadJsonAsBytes)           //store loads with contractNumber as key
        if err != nil {
                return nil, err
        }

        //get the marble index
        loadsAsBytes, err := stub.GetState(contractNumberIndexStr)
        if err != nil {
                return nil, errors.New("Failed to get load index")
        }

        var loadIndex []string
        json.Unmarshal(loadsAsBytes, &loadIndex)        //un stringify it aka JSON.parse()

        //append
        loadIndex = append(loadIndex, contractNumber)        //add loadnumber to index list
        fmt.Println("! load index: ", loadIndex)
        jsonAsBytes, _ := json.Marshal(loadIndex)
        err = stub.PutState(contractNumberIndexStr, jsonAsBytes)     //store load

        fmt.Println("- Completed init_load()")
        return nil, nil
}
