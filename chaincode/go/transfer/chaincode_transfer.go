
package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
	"encoding/json"
)

var logger = shim.NewLogger("TransferChaincode")

// TransferChaincode example simple Chaincode implementation
type TransferChaincode struct {
}

type Domicile struct {
	Name string
	Community string
}

func (t *TransferChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *TransferChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	function, args := stub.GetFunctionAndParameters()
	if function == "depart" {
		return t.depart(stub, name, org)
	} else if function == "arrive" {
		return t.arrive(stub, name, org)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"Invalid invoke function name."}
}

func (t *TransferChaincode) depart(stub shim.ChaincodeStubInterface, name string, org string) pb.Response {

	orgBytes, err := stub.GetState(name)
	if err != nil {
		return shim.Error(err.Error())
	}

	currentOrg := string(orgBytes)

	if currentOrg != org {
		return pb.Response{Status:409,Message:"Not your current community"}
	}

	deceased := t.checkDeceased(stub, name)
	if deceased {
		return pb.Response{Status:409,Message:"Wha?! You're a dead man!"}
	}

	err = stub.PutState(name, []byte("in transit"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *TransferChaincode) arrive(stub shim.ChaincodeStubInterface, name string, org string) pb.Response {

	orgBytes, err := stub.GetState(name)
	if err != nil {
		return shim.Error(err.Error())
	}

	deceased := t.checkDeceased(stub, name)
	if deceased {
		return pb.Response{Status:409,Message:"Wha?! You're a dead man!"}
	}

	if orgBytes != nil && "in transit" != string(orgBytes) {
		return pb.Response{Status:409,Message:"Has not departed"}
	}

	err = stub.PutState(name, []byte(org))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *TransferChaincode) checkDeceased(stub shim.ChaincodeStubInterface, name string) bool {
	response := stub.InvokeChaincode("governance", [][]byte{[]byte("query"),[]byte(name)}, "common")

	if response.Status != shim.OK {
		return false
	} else {
		return true
	}
}

func (t *TransferChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 1 {
		name := args[0]

		valBytes, err := stub.GetState(name)
		if err != nil {
			return shim.Error(err.Error())
		}

		if valBytes == nil {
			return pb.Response{Status:404, Message:"Entity not found"}
		}

		d := Domicile{Name: name, Community: string(valBytes)}

		jsonBytes, err := json.Marshal(d)
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(jsonBytes)
	} else if len(args) == 0 {
		it, err := stub.GetStateByRange("", "")
		if err != nil {
			return shim.Error(err.Error())
		}

		defer it.Close()

		domiciles := []Domicile{}

		for it.HasNext() {
			next, err := it.Next()
			if err != nil {
				return shim.Error(err.Error())
			}

			d := Domicile{Name: next.Key, Community: string(next.Value)}

			domiciles = append(domiciles, d)
		}

		jsonBytes, err := json.Marshal(domiciles)
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(jsonBytes)

	}

	return pb.Response{Status:403, Message:"No name provided"}
}

var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(TransferChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
