package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Student structure
type Student struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	CGPA   string `json:"cgpa"`
}

// SmartContract provides functions for managing students
type SmartContract struct {
	contractapi.Contract
}

// InitLedger adds initial students
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	students := []Student{
		{ID: "S1", Name: "Alice", Branch: "CSE", CGPA: "9.1"},
		{ID: "S2", Name: "Bob", Branch: "ECE", CGPA: "8.5"},
	}

	for _, student := range students {
		studentJSON, err := json.Marshal(student)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(student.ID, studentJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// CreateStudent adds a new student
func (s *SmartContract) CreateStudent(ctx contractapi.TransactionContextInterface, id string, name string, branch string, cgpa string) error {
	exists, err := s.StudentExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the student %s already exists", id)
	}

	student := Student{
		ID:     id,
		Name:   name,
		Branch: branch,
		CGPA:   cgpa,
	}

	studentJSON, err := json.Marshal(student)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, studentJSON)
}

// ReadStudent returns a student
func (s *SmartContract) ReadStudent(ctx contractapi.TransactionContextInterface, id string) (*Student, error) {
	studentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if studentJSON == nil {
		return nil, fmt.Errorf("the student %s does not exist", id)
	}

	var student Student
	err = json.Unmarshal(studentJSON, &student)
	if err != nil {
		return nil, err
	}

	return &student, nil
}

// GetAllStudents returns all students
func (s *SmartContract) GetAllStudents(ctx contractapi.TransactionContextInterface) ([]*Student, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var students []*Student
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var student Student
		err = json.Unmarshal(queryResponse.Value, &student)
		if err != nil {
			return nil, err
		}
		students = append(students, &student)
	}

	return students, nil
}

// StudentExists returns true if student exists
func (s *SmartContract) StudentExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	studentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}

	return studentJSON != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode: %v", err))
	}

	if err := chaincode.Start(); err != nil {
		panic(fmt.Sprintf("Error starting chaincode: %v", err))
	}
}

