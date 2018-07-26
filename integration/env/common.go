package env

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarClusterID is the process environment variable representing the
	// CLUSTER_NAME env var.
	//
	// TODO rename to CLUSTER_ID. Note this also had to be changed in the
	// framework package of e2e-harness.
	EnvVarClusterID = "CLUSTER_NAME"
	// EnvVarK8sApiUrl is the process environment variable representing the
	// k8s api url for testing cluster.
	EnvVarK8sApiUrl = "K8S_API_URL"
	// EnvVarK8sCert is the process environment variable representing the
	// k8s kubeconfig cert value for testing cluster.
	EnvVarK8sCert = "K8S_CERT_ENCODED"
	// EnvVarK8sCert is the process environment variable representing the
	// k8s kubeconfig ca cert value for testing cluster.
	EnvVarK8sCertCA = "K8S_CERT_CA_ENCODED"
	// EnvVarK8sCert is the process environment variable representing the
	// k8s kubeconfig private key value for testing cluster.
	EnvVarK8sCertPrivate = "K8S_CERT_PRIVATE_ENCODED"
	// EnvVarCommonDomain is the process environment variable representing the
	// COMMON_DOMAIN env var.
	EnvVarCommonDomain = "COMMON_DOMAIN"
	// EnvVarGithubBotToken is the process environment variable representing
	// the GITHUB_BOT_TOKEN env var.
	EnvVarGithubBotToken = "GITHUB_BOT_TOKEN"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"
	// EnvVarTestedVersion is the process environment variable representing the
	// TESTED_VERSION env var.
	EnvVarTestedVersion = "TESTED_VERSION"
	// EnvVarTestDir is the process environment variable representing the
	// TEST_DIR env var.
	EnvVarTestDir = "TEST_DIR"
	// EnvVaultToken is the process environment variable representing the
	// VAULT_TOKEN env var.
	EnvVaultToken = "VAULT_TOKEN"
	// EnvVarVersionBundleVersion is the process environment variable representing
	// the VERSION_BUNDLE_VERSION env var.
	EnvVarVersionBundleVersion = "VERSION_BUNDLE_VERSION"
)

var (
	circleCI             string
	circleSHA            string
	clusterID            string
	commonDomain         string
	k8sApiUrl            string
	k8sCert              string
	k8sCertPrivate       string
	k8sCertCA            string
	testDir              string
	testedVersion        string
	keepResources        string
	vaultToken           string
	versionBundleVersion string
)

func init() {
	var err error

	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	testedVersion = os.Getenv(EnvVarTestedVersion)
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarTestedVersion))
	}

	testDir = os.Getenv(EnvVarTestDir)

	k8sApiUrl = os.Getenv(EnvVarK8sApiUrl)
	if k8sApiUrl == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarK8sApiUrl))
	}
	k8sCert = os.Getenv(EnvVarK8sCertCA)
	if k8sCert == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarK8sCertCA))
	}
	k8sCertCA = os.Getenv(EnvVarK8sCert)
	if k8sCertCA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarK8sCert))
	}
	k8sCertPrivate = os.Getenv(EnvVarK8sCertPrivate)
	if k8sCertPrivate == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarK8sCertPrivate))
	}

	// NOTE that implications of changing the order of initialization here means
	// breaking the initialization behaviour.
	clusterID := os.Getenv(EnvVarClusterID)
	if clusterID == "" {
		os.Setenv(EnvVarClusterID, ClusterID())
	}

	commonDomain = os.Getenv(EnvVarCommonDomain)
	if commonDomain == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCommonDomain))
	}

	vaultToken = os.Getenv(EnvVaultToken)
	if vaultToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVaultToken))
	}

	token := os.Getenv(EnvVarGithubBotToken)
	params := &framework.VBVParams{
		Component: "kvm-operator",
		Provider:  "kvm",
		Token:     token,
		VType:     TestedVersion(),
	}
	versionBundleVersion, err = framework.GetVersionBundleVersion(params)
	if err != nil {
		panic(err.Error())
	}
	// TODO there should be a not found error returned by the framework in such
	// cases.
	if VersionBundleVersion() == "" {
		if strings.ToLower(TestedVersion()) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}
		panic("version bundle version  must not be empty")
	}
	os.Setenv(EnvVarVersionBundleVersion, VersionBundleVersion())
}

func CircleCI() string {
	return circleCI
}

func CircleSHA() string {
	return circleSHA
}

// ClusterID returns a cluster ID unique to a run integration test. It might
// look like ci-wip-3cc75-5e958.
//
//     ci is a static identifier stating a CI run of the aws-operator.
//     wip is a version reference which can also be cur for the current version.
//     3cc75 is the Git SHA.
//     5e958 is a hash of the integration test dir, if any.
//
func ClusterID() string {
	var parts []string

	parts = append(parts, "ci")
	parts = append(parts, TestedVersion()[0:3])
	parts = append(parts, CircleSHA()[0:5])
	if TestHash() != "" {
		parts = append(parts, TestHash())
	}

	return strings.Join(parts, "-")
}

func CommonDomain() string {
	return commonDomain
}

func K8sApiUrl() string {
	return k8sApiUrl
}

func K8sCert() string {
	return k8sCert
}

func K8sCertCa() string {
	return k8sCertCA
}

func K8sCertPrivate() string {
	return k8sCertPrivate
}

func KeepResources() string {
	return keepResources
}

func TestedVersion() string {
	return testedVersion
}

func TestDir() string {
	return testDir
}

func TestHash() string {
	if TestDir() == "" {
		return ""
	}

	h := sha1.New()
	h.Write([]byte(TestDir()))
	s := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	return s
}

func VaultToken() string {
	return vaultToken
}

func VersionBundleVersion() string {
	return versionBundleVersion
}
