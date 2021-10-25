package provider

import (
	"fmt"
	"os"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

var SlackResources = []SlackResource{
	&SlackConversationResource{},
}
var SlackFunctions = []SlackFunction{
	&SlackUserFunction{},
}

type SlackConfig struct {
	Config map[string]string
}

type SlackFunction interface {
	Configure(config SlackConfig)
	Invoke(s *slackProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error)
	Name() string
}

type SlackResource interface {
	Configure(config SlackConfig)
	Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error)
	Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error)
	Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error)
	Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error)
	Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error)
	Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error)
	Name() string
}

type ResourceBase interface {
	GetUrn() string
}

func (sc *SlackConfig) getConfig(configName, envName string) string {
	if val, ok := sc.Config[configName]; ok {
		return val
	}

	return os.Getenv(envName)
}

func (sc *SlackConfig) getSlackToken() (*string, error) {
	token := sc.getConfig("token", "SLACK_TOKEN")

	if len(token) == 0 {
		return nil, fmt.Errorf("no slack token found")
	}

	return &token, nil
}

func getSlackResource(name string) SlackResource {
	for _, r := range SlackResources {
		if r.Name() == name {
			return r
		}
	}

	return &SlackUnknownResource{}
}

func getSlackFunction(name string) SlackFunction {
	for _, r := range SlackFunctions {
		if r.Name() == name {
			return r
		}
	}

	return &SlackUnknownFunction{}
}

func getResourceNameFromRequest(req ResourceBase) string {
	urn := resource.URN(req.GetUrn())
	return urn.Type().String()
}

func findStringInsideArray(arr []string, text string) bool {
	for _, s := range arr {
		if s == text {
			return true
		}
	}

	return false
}
