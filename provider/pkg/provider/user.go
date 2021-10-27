package provider

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/slack-go/slack"
)

type User struct {
	Email string `pulumi:"email"`
	Id    string `pulumi:"id"`
	Name  string `pulumi:"name"`
}

type SlackUserFunction struct {
	config SlackConfig
}

func (u *SlackUserFunction) Name() string {
	return "slack:index:LookupUser"
}

func (u *SlackUserFunction) Configure(config SlackConfig) {
	u.config = config
}

func (c *SlackUserFunction) lookupUser(email string) (*User, error) {
	token, err := c.config.getSlackToken()
	if err != nil {
		return nil, err
	}

	api := slack.New(*token)
	user, err := api.GetUserByEmail(email)

	if err != nil {
		return nil, err
	}

	return &User{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Profile.Email,
	}, nil
}

func (u *SlackUserFunction) Invoke(s *slackProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	label := fmt.Sprintf("%s.Invoke(%s)", s.name, req.Tok)
	logging.V(9).Infof("%s executing", label)

	inputs, err := plugin.UnmarshalProperties(req.GetArgs(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.args", label), KeepUnknowns: true, SkipNulls: true, KeepSecrets: true,
	})
	if err != nil {
		return nil, err
	}

	outputs := make(map[string]interface{})
	if req.Tok != "slack:index:LookupUser" {
		return nil, fmt.Errorf("invalid function call %s", req.GetTok())
	}

	emailInput := inputs["email"].StringValue()
	user, err := u.lookupUser(emailInput)
	if err != nil {
		return nil, err
	}
	outputs["user"] = user

	result, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputs),
		plugin.MarshalOptions{Label: fmt.Sprintf("%s.response", label), KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	return &rpc.InvokeResponse{Return: result}, nil
}
