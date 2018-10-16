package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/giantswarm/azure-operator/integration/network"
)

const (
	EnvVarAzureCIDR             = "AZURE_CIDR"
	EnvVarAzureCalicoSubnetCIDR = "AZURE_CALICO_SUBNET_CIDR"
	EnvVarAzureMasterSubnetCIDR = "AZURE_MASTER_SUBNET_CIDR"
	EnvVarAzureVPNSubnetCIDR    = "AZURE_VPN_SUBNET_CIDR"
	EnvVarAzureWorkerSubnetCIDR = "AZURE_WORKER_SUBNET_CIDR"

	EnvVarAzureClientID       = "AZURE_CLIENTID"
	EnvVarAzureClientSecret   = "AZURE_CLIENTSECRET"
	EnvVarAzureLocation       = "AZURE_LOCATION"
	EnvVarAzureSubscriptionID = "AZURE_SUBSCRIPTIONID"
	EnvVarAzureTenantID       = "AZURE_TENANTID"

	EnvVarAzureGuestClientID       = "AZURE_GUEST_CLIENTID"
	EnvVarAzureGuestClientSecret   = "AZURE_GUEST_CLIENTSECRET"
	EnvVarAzureGuestSubscriptionID = "AZURE_GUEST_SUBSCRIPTIONID"
	EnvVarAzureGuestTenantID       = "AZURE_GUEST_TENANTID"

	EnvVarCommonDomainResourceGroup = "COMMON_DOMAIN_RESOURCE_GROUP"
	EnvVarCircleBuildNumber         = "CIRCLE_BUILD_NUM"
)

var (
	azureClientID       string
	azureClientSecret   string
	azureLocation       string
	azureSubscriptionID string
	azureTenantID       string

	azureGuestClientID       string
	azureGuestClientSecret   string
	azureGuestSubscriptionID string
	azureGuestTenantID       string

	azureCIDR             string
	azureCalicoSubnetCIDR string
	azureMasterSubnetCIDR string
	azureVPNSubnetCIDR    string
	azureWorkerSubnetCIDR string

	commonDomainResourceGroup string
)

func init() {
	azureClientID = os.Getenv(EnvVarAzureClientID)
	if azureClientID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureClientID))
	}

	azureClientSecret = os.Getenv(EnvVarAzureClientSecret)
	if azureClientSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureClientSecret))
	}

	azureLocation = os.Getenv(EnvVarAzureLocation)
	if azureLocation == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureLocation))
	}

	azureSubscriptionID = os.Getenv(EnvVarAzureSubscriptionID)
	if azureSubscriptionID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureSubscriptionID))
	}

	azureTenantID = os.Getenv(EnvVarAzureTenantID)
	if azureTenantID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureTenantID))
	}

	azureGuestClientID = os.Getenv(EnvVarAzureGuestClientID)
	if azureGuestClientID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureGuestClientID))
	}

	azureGuestClientSecret = os.Getenv(EnvVarAzureGuestClientSecret)
	if azureGuestClientSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureGuestClientSecret))
	}

	azureGuestSubscriptionID = os.Getenv(EnvVarAzureGuestSubscriptionID)
	if azureGuestSubscriptionID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureGuestSubscriptionID))
	}

	azureGuestTenantID = os.Getenv(EnvVarAzureGuestTenantID)
	if azureGuestTenantID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureGuestTenantID))
	}

	commonDomainResourceGroup = os.Getenv(EnvVarCommonDomainResourceGroup)
	if commonDomainResourceGroup == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCommonDomainResourceGroup))
	}

	// azureCDIR must be provided along with other CIDRs,
	// otherwise we compute CIDRs base on EnvVarCircleBuildNumber value.
	azureCDIR := os.Getenv(EnvVarAzureCIDR)
	if azureCDIR == "" {
		buildNumber, err := strconv.ParseUint(os.Getenv(EnvVarCircleBuildNumber), 10, 32)
		if err != nil {
			panic(err)
		}

		subnets, err := network.ComputeSubnets(uint(buildNumber))
		if err != nil {
			panic(err)
		}

		azureCIDR = subnets.Parent.String()
		azureCalicoSubnetCIDR = subnets.Calico.String()
		azureMasterSubnetCIDR = subnets.Master.String()
		azureVPNSubnetCIDR = subnets.VPN.String()
		azureWorkerSubnetCIDR = subnets.Worker.String()
	}
}

func AzureCalicoSubnetCIDR() string {
	return azureCalicoSubnetCIDR
}

func AzureClientID() string {
	return azureClientID
}

func AzureClientSecret() string {
	return azureClientSecret
}

func AzureCIDR() string {
	return azureCIDR
}

func AzureGuestClientID() string {
	return azureGuestClientID
}

func AzureGuestClientSecret() string {
	return azureGuestClientSecret
}

func AzureGuestSubscriptionID() string {
	return azureGuestSubscriptionID
}

func AzureGuestTenantID() string {
	return azureGuestTenantID
}

func AzureLocation() string {
	return azureLocation
}

func AzureMasterSubnetCIDR() string {
	return azureMasterSubnetCIDR
}

func AzureSubscriptionID() string {
	return azureSubscriptionID
}

func AzureTenantID() string {
	return azureTenantID
}

func AzureVPNSubnetCIDR() string {
	return azureVPNSubnetCIDR
}

func AzureWorkerSubnetCIDR() string {
	return azureWorkerSubnetCIDR
}

func CommonDomainResourceGroup() string {
	return commonDomainResourceGroup
}
