# Hyperledger Fabric REST API

A REST API service that provides HTTP endpoints to interact with Hyperledger Fabric chaincode through the Fabric Gateway.

## Overview

This project provides a REST API layer on top of Hyperledger Fabric's Gateway SDK, allowing applications to interact with Student Records chaincode without direct integration with the Fabric SDK. The API simplifies blockchain interactions by abstracting away the complexity of Fabric's connection protocols.

## Features

- HTTP endpoints for all Student Records chaincode operations
- Integration with Fabric Gateway for simplified blockchain access
- JSON-based request and response formats
- Connection management with Hyperledger Fabric networks

## Prerequisites

- Go 1.16 or higher
- Access to a running Hyperledger Fabric network (test-network)
- Student Records chaincode deployed on the Fabric network
- Valid network credentials (certificates and private keys)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/VishnuKC26/Hyperledger-Fabric-Rest-API.git
   cd Hyperledger-Fabric-Rest-API
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Configuration

Configure the API to connect to your Fabric network by setting the following environment variables or updating the configuration in the code:

- `FABRIC_MSP_ID` - The MSP ID for your organization
- `FABRIC_CHANNEL_NAME` - The channel where your chaincode is deployed
- `FABRIC_CHAINCODE_NAME` - The name of your deployed chaincode
- `FABRIC_CERT_PATH` - Path to the user certificate
- `FABRIC_KEY_PATH` - Path to the user private key
- `FABRIC_GATEWAY_PEER` - Address of the peer node

## Usage

1. Start the REST API server:
   ```bash
   go run rest-api.go
   ```

2. The API will be available at `http://localhost:8080` (or your configured port)

## API Endpoints

### Student Records API

- `POST /students`: Create a new student record
- `GET /students/:id`: Retrieve a student record by ID
- `PUT /students/:id`: Update an existing student record
- `DELETE /students/:id`: Delete a student record
- `GET /students`: Query all student records

## Integration with Fabric

The API connects to Fabric using the Gateway SDK with the following components:

- `studentrecords_client.go`: Client implementation for interacting with the chaincode
- `rest-api.go`: HTTP server that exposes the API endpoints

## Development

To modify or extend the API:

1. Update the route handlers in `rest-api.go`
2. Add new chaincode functions in `studentrecords_client.go`
3. Test your changes by running the API and making requests

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Hyperledger Fabric Community
- Fabric Samples Project
