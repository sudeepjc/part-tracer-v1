package parttracer

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	s "strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PartTrade struct {
	contractapi.Contract
}

func (pt *PartTrade) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("initLedger has been invoked")

	ci, _ := ctx.GetClientIdentity().GetID()
	fmt.Println("ClientIdentity : ", ci)

	msp, _:= ctx.GetClientIdentity().GetMSPID()
	fmt.Println("MSPID : ", msp)

	tx:= ctx.GetStub().GetTxID()
	fmt.Println("TXID : ", tx)

	chanl:= ctx.GetStub().GetChannelID()
	fmt.Println("ChannelID : ", chanl)

	tim, _ := ctx.GetStub().GetTxTimestamp()

	txTime, _ := ptypes.Timestamp(tim)

	PartID := s.Join([]string{"pName",txTime.Format("2006-01-02_5:04:05")},"_")
	
	fmt.Println("Tx timestamp : ", PartID)


	return nil
}

func (pt *PartTrade) AddPart(ctx contractapi.TransactionContextInterface, partId string, pName string, desc string, qprice uint32, maker string) error {

	if len(partId) == 0 {
		return fmt.Errorf("Invalid part ID")
	}

	if len(pName) == 0 {
		return fmt.Errorf("Invalid part Name info")
	}

	if len(desc) == 0 {
		return fmt.Errorf("Invalid description ")
	}

	if len(maker) == 0 {
		return fmt.Errorf("Invalid manufacturer info")
	}

	if qprice <= 0 {
		return fmt.Errorf("Invalid quote price info")
	}

	owner, _:= ctx.GetClientIdentity().GetMSPID()

	// partId := s.Join([]string{pName,currentTime.Format("2006-01-02_5:04:05")},"_")
	part := Part{ PartID: partId, PartName: pName, Description: desc, QuotePrice: qprice, Manufacturer:maker, Owner:owner }
	part.SetNew()


	// use tx time for deterministic behavior of the execution
	tim, _ := ctx.GetStub().GetTxTimestamp()
	txTime, _ := ptypes.Timestamp(tim)
	part.EventTime = txTime.Format("2006-01-02_5:04:05")

	partAsBytes, err := part.Serialize()

	if err != nil {
		return fmt.Errorf("Failed to add part while serializing data %s", err.Error())
	}

	fmt.Println("added part ", partId)

	return ctx.GetStub().PutState(partId, partAsBytes)
}

func (pt *PartTrade) QueryPart(ctx contractapi.TransactionContextInterface, partId string) (*Part, error) {

	if len(partId) == 0 {
		return nil, fmt.Errorf("Invalid part ID")
	}

	partAsBytes, err := ctx.GetStub().GetState(partId)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if partAsBytes == nil {
		return nil, fmt.Errorf("%s : does not exist", partId)
	}

	part := new(Part)
	_ = Deserialize(partAsBytes, part)

	return part, nil
}

func (pt *PartTrade) SellPart(ctx contractapi.TransactionContextInterface, partId string, buyer string, dprice uint32 ) (*Part, error) {

	if len(partId) == 0 {
		return nil, fmt.Errorf("Invalid part ID")
	}

	if dprice <= 0 {
		return nil, fmt.Errorf("Invalid dprice")
	}
	
	partAsBytes, err := ctx.GetStub().GetState(partId)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if partAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", partId)
	}

	part := new(Part)
	_ = Deserialize(partAsBytes, part)

	seller, _:= ctx.GetClientIdentity().GetMSPID()

	if part.Owner != seller {
		return nil, fmt.Errorf("Part %s is not owned by %s", partId, seller)
	}

	if part.IsNew() {
		part.SetUsed()
	}

	tim, _ := ctx.GetStub().GetTxTimestamp()
	txTime, _ := ptypes.Timestamp(tim)
	part.EventTime = txTime.Format("2006-01-02_5:04:05")

	part.SetOwner(buyer)
	part.DealPrice = dprice

	updatedPartAsBytes, err := part.Serialize()

	if err != nil {
		return nil, fmt.Errorf("Failed to update part while serializing data %s", err.Error())
	}

	fmt.Println("updated part ", partId)

	err = ctx.GetStub().PutState(partId, updatedPartAsBytes)

	if err != nil {
		return nil, fmt.Errorf("Error while trying to add sell data to state: %s", err.Error())
	}

	return part, nil
}


