
package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("GovernanceChaincode")

// GovernanceChaincode example simple Chaincode implementation
type GovernanceChaincode struct {
}

func (t *GovernanceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *GovernanceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	function, args := stub.GetFunctionAndParameters()
	if function == "decease" {
		return t.decease(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"Invalid invoke function name."}
}

func (t *GovernanceChaincode) decease(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.")
	}

	name := args[0]

	err := stub.PutState(name, []byte("deceased"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *GovernanceChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.")
	}

	name := args[0]

	valBytes, err := stub.GetState(name)
	if err != nil {
		return shim.Error(err.Error())
	}

	if string(valBytes) == "deceased" {
		return shim.Success(nil)
	} else {
		return pb.Response{Status:404}
	}
}

func main() {
	err := shim.Start(new(GovernanceChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
