/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "dns:///localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
	listenAddr   = ":3000" // REST API server port
)

// Global variables to store Fabric client connections
var (
	contract *client.Contract
	network  *client.Network
	gw       *client.Gateway
)

// Student represents a student record
type Student struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Year       string `json:"year"`
	CGPA       string `json:"cgpa"`
}

func main() {
	// Initialize Fabric connection
	initFabricClient()
	defer gw.Close()

	// Initialize and start the REST API server
	router := setupRouter()
	log.Printf("Starting REST API server on %s", listenAddr)
	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initFabricClient initializes the connection to the Fabric network
func initFabricClient() {
	// The gRPC client connection is shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()

	id := newIdentity()
	sign := newSign()

	// Establish a Gateway connection using identity, sign function, and gRPC connection
	var err error
	gw, err = client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		// Set timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	// Override default chaincode and channel names through environment variables if present
	chaincodeName := "studentrecords"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	// Get the network and contract instances
	network = gw.GetNetwork(channelName)
	contract = network.GetContract(chaincodeName)

	log.Println("Fabric client initialized successfully")
}

// setupRouter configures the Gin router with endpoints
func setupRouter() *gin.Engine {
	router := gin.Default()

	// Middleware for handling errors
	router.Use(gin.Recovery())

	// Define API routes
	router.GET("/api/students", getAllStudents)
	router.GET("/api/students/:id", getStudentByID)
	router.POST("/api/students", createStudent)
	router.PUT("/api/students/:id", updateStudent)
	router.DELETE("/api/students/:id", deleteStudent)
	router.POST("/api/init", initLedger)

	return router
}

// initLedger initializes the ledger with sample data
func initLedger(c *gin.Context) {
	log.Println("Initializing ledger...")

	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize ledger: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ledger initialized successfully"})
}

// getAllStudents retrieves all student records
func getAllStudents(c *gin.Context) {
	log.Println("Retrieving all students...")

	result, err := contract.EvaluateTransaction("GetAllStudents")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get students: %v", err)})
		return
	}

	var students []map[string]interface{}
	if err := json.Unmarshal(result, &students); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse student data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, students)
}

// getStudentByID retrieves a specific student by ID
func getStudentByID(c *gin.Context) {
	id := c.Param("id")
	log.Printf("Retrieving student with ID: %s", id)

	result, err := contract.EvaluateTransaction("ReadStudent", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Student not found: %v", err)})
		return
	}

	var student map[string]interface{}
	if err := json.Unmarshal(result, &student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse student data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, student)
}

// createStudent adds a new student record
func createStudent(c *gin.Context) {
	var student Student

	// Parse request body
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	log.Printf("Creating student with ID: %s", student.ID)

	// Submit transaction to create student
	_, err := contract.SubmitTransaction(
		"CreateStudent", 
		student.ID, 
		student.Name, 
		student.Department, 
		student.Year, 
		student.CGPA,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create student: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, student)
}

// updateStudent updates an existing student record
func updateStudent(c *gin.Context) {
	id := c.Param("id")
	var student Student

	// Parse request body
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	log.Printf("Updating student with ID: %s", id)

	// Use the ID from the URL path rather than from the JSON body
	_, err := contract.SubmitTransaction(
		"UpdateStudent", 
		id, 
		student.Name, 
		student.Department, 
		student.Year, 
		student.CGPA,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update student: %v", err)})
		return
	}

	// Set the ID to be consistent with the URL parameter
	student.ID = id
	c.JSON(http.StatusOK, student)
}

// deleteStudent removes a student record
func deleteStudent(c *gin.Context) {
	id := c.Param("id")
	log.Printf("Deleting student with ID: %s", id)

	_, err := contract.SubmitTransaction("DeleteStudent", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete student: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Student %s deleted successfully", id)})
}

// newGrpcConnection creates a secure gRPC connection to the Fabric gateway (peer)
func newGrpcConnection() *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS certificate file: %w", err))
	}

	// Parse the TLS certificate from PEM
	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	// Create a certificate pool and add our peer's TLS certificate
	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)

	// Create transport credentials that enforce TLS and check the server's name
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	// Create the gRPC client connection using the peer endpoint and transport credentials
	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity using an X.509 certificate
func newIdentity() *identity.X509Identity {
	certificatePEM, err := readFirstFile(certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	// Parse the certificate
	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	// Create a new X509 identity using the MSP ID and the parsed certificate
	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a signing function using the user's private key
func newSign() identity.Sign {
	privateKeyPEM, err := readFirstFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	// Parse the PEM-encoded private key
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	// Create a signing function from the private key
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

// readFirstFile reads the first file found within the given directory
func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	fileNames, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path.Join(dirPath, fileNames[0]))
}

// formatJSON pretty prints JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return string(data)
	}
	return prettyJSON.String()
}
