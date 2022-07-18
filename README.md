# Network API

A multi cloud network API that let you automate the management of multiple VPCs in diferent cloud providers.

It works as a serverless application running primarily on AWS, using AWS SAM for orchestrating the deployment and DynamoDB for storing network metadata. Providers are pluggable by using a webhook, currently only `aws-provider` is available.

### Subnets Layout

The API generates subnets for any given size of network, from `/16` down to `/24`, using the following layout, example given using a `10.0.0.0/20` network.

|                  | subnet ranges | type    |
|------------------|---------------|---------|
| **10.0.0.0/22**  | 10.0.0.0/23   | private |
|                  | 10.0.2.0/24   | public  |
|                  | 10.0.3.0/28   | tgw     |
| **10.0.4.0/22**  | 10.0.4.0/23   | private |
|                  | 10.0.6.0/24   | public  |
|                  | 10.0.7.0/28   | tgw     |
| **10.0.8.0/22**  | 10.0.8.0/23   | private |
|                  | 10.0.10.0/24  | public  |
|                  | 10.0.11.0/28  | tgw     |
| **10.0.12.0/22** |               | spare   |

## Providers

Providers webhook receive the following payload when called:

```go
const (
	CreateNetwork EventType = "create_network"
	CheckNework   EventType = "check_network"
	DeleteNetwork EventType = "delete_network"
	QueryNetwork  EventType = "query_network"
)

type ProviderWebhook struct {
	Event       EventType `json:"event"`
	NetworkID   string    `json:"networkID" validate:"required"`
	Account     string    `json:"account" validate:"required"`
	Region      string    `json:"region" validate:"required"`
	Environment string    `json:"environment" validate:"required"`
	CIDR        string    `json:"cidr" validate:"required_if=Event create_network,omitempty,cidr"`
	Subnets     []*Subnet `json:"subnets,omitempty" validate:"required_if=Event create_network,omitempty"`
}
```

### AWS

AWS Provider uses a cloudformation template for creating new VPCs, which is stored in a bucket. The lambda has a default role that allows it to assume roles in multiple accounts, for this to work you have to deploy a stackset on your master account using `aws_provider_trust_role.yaml`.

For security it uses a KMS key to validate the used token, to create a token use `aws-provider-token` command.

The cloudformation template has the following parameters:

| Parameter          | Description                                                               |
|--------------------|---------------------------------------------------------------------------|
| VPCName            | The name of the VPC being created.                                        |
| Environment        | The VPC environment. Values: prod or qa                                   |
| VPCCidr            | The CIDR of the VPC being created. Example: 10.0.0.0/16                   |
| PublicSubnet0Cidr  | The CIDR of the public subnet being created. Example: 10.0.0.0/16         |
| PublicSubnet1Cidr  |                                                                           |
| PublicSubnet2Cidr  |                                                                           |
| PrivateSubnet0Cidr | The CIDR of the private subnet being created. Example: 10.0.0.0/16        |
| PrivateSubnet1Cidr |                                                                           |
| PrivateSubnet2Cidr |                                                                           |
| TGWSubnet0Cidr     | The CIDR of the TGW Attachment subnet being created. Example: 10.0.0.0/16 |
| TGWSubnet1Cidr     |                                                                           |
| TGWSubnet2Cidr     |                                                                           |

## Running local

```
GOARCH=amd64 GOOS=linux go build -o deployment/network-api ./cmd/network-api
sam local start-api --template-file deployment/sam_network_api.yaml

# or:
make run
```

## Deploy API

Fill parameters.json:
```json
[
    {
        "ParameterKey": "VpcId",
        "ParameterValue": ""
    },
    {
        "ParameterKey": "SubnetIds",
        "ParameterValue": ""
    },
    {
        "ParameterKey": "OIDCAudience",
        "ParameterValue": ""
    },
    {
        "ParameterKey": "OIDCIssuer",
        "ParameterValue": ""
    },
    {
        "ParameterKey": "OIDCScopes",
        "ParameterValue": ""
    },
    {
        "ParameterKey": "OIDCJwksURL",
        "ParameterValue": ""
    }
]
```

Then:
```
make package
make deploy
```

## Deploy AWS Provider
```
make package_provider
make deploy_provider
```

## Configuring

```bash
export ENDPOINT="..."

# Pools
curl -H "Content-Type: application/json" -d '{"region":"us-east-1","name":"main-aws","subnetIP":"10.0.0.0","subnetMaxIP":"10.240.255.255"}' -v $ENDPOINT/api/v1/pools
curl -H "Content-Type: application/json" -d '{"region":"sa-east-1","name":"southamerica-aws","subnetIP":"10.240.0.0","subnetMask":16}' -v $ENDPOINT/api/v1/pools

# Providers
curl -H "Content-Type: application/json" -d '{"name":"aws","webhookURL":"https://something.localhost","apiToken":"1234token"}' -v $ENDPOINT/api/v1/providers

# Networks
curl -H "Content-Type: application/json" -d '{"region":"us-east-1","subnetSize":16,"account":"123","provider":"aws","environment":"prod","attachTGW":true,"privateSubnet":true,"publicSubnet":true}' -v $ENDPOINT/api/v1/networks
```

## Network-CLI
There's a CLI tool to easily call network-api actions and help you automate some jobs.

Install:
```
go install ./cmd/network-cli

# or:
make install_cli
```

Show available commands:
```
network-cli --help
```
