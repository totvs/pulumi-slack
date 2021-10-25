package provider

import (
	"fmt"
	"time"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	logger "github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/slack-go/slack"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var DEPRECATION_CHANNEL_NAME_FORMAT = "%s-deleted-%s"

type SlackConversationResource struct {
	config SlackConfig
}

type SlackConversationInput struct {
	Name       string
	Topic      string
	Purpose    string
	IsPrivate  bool
	IsArchived bool
	Members    []string
}

func (i *SlackConversationInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["topic"] = resource.NewPropertyValue(i.Topic)
	pm["purpose"] = resource.NewPropertyValue(i.Purpose)
	pm["isPrivate"] = resource.NewPropertyValue(i.IsPrivate)
	pm["isArchived"] = resource.NewPropertyValue(i.IsArchived)
	pm["members"] = resource.NewPropertyValue(i.Members)

	return pm
}

func (r *SlackConversationResource) ToSlackConversationInput(inputMap resource.PropertyMap) SlackConversationInput {
	input := SlackConversationInput{}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}
	if inputMap["topic"].HasValue() && inputMap["topic"].IsString() {
		input.Topic = inputMap["topic"].StringValue()
	}
	if inputMap["purpose"].HasValue() && inputMap["purpose"].IsString() {
		input.Purpose = inputMap["purpose"].StringValue()
	}
	if inputMap["isPrivate"].HasValue() && inputMap["isPrivate"].IsBool() {
		input.IsPrivate = inputMap["isPrivate"].BoolValue()
	}
	if inputMap["isArchived"].HasValue() && inputMap["isArchived"].IsBool() {
		input.IsArchived = inputMap["isArchived"].BoolValue()
	}
	if inputMap["members"].HasValue() && inputMap["members"].IsArray() {
		for _, m := range inputMap["members"].ArrayValue() {
			if m.HasValue() && m.IsString() {
				input.Members = append(input.Members, m.StringValue())
			}
		}
	}

	return input
}

func (c *SlackConversationResource) Name() string {
	return "slack:index:Conversation"
}

func (c *SlackConversationResource) Configure(config SlackConfig) {
	c.config = config
}

func (c *SlackConversationResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	diffs := olds["__inputs"].ObjectValue().Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes:             pulumirpc.DiffResponse_DIFF_NONE,
			Replaces:            []string{},
			Stables:             []string{},
			DeleteBeforeReplace: false,
		}, nil
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if diffs.Changed("name") ||
		diffs.Changed("purpose") ||
		diffs.Changed("topic") ||
		diffs.Changed("isPrivate") ||
		diffs.Changed("isArchived") ||
		diffs.Changed("members") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            []string{},
		Stables:             []string{},
		DeleteBeforeReplace: false,
	}, nil
}

func (scr *SlackConversationResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsConversation := scr.ToSlackConversationInput(inputs)
	channelId, err := scr.createConversation(inputsConversation)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation '%s': %s", inputsConversation.Name, err.Error())
	}

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputs)

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         *channelId,
		Properties: outputProperties,
	}, nil
}

func (c *SlackConversationResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	_, err := c.deleteConversation(req.Id)

	return &pbempty.Empty{}, err
}

func (k *SlackConversationResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (k *SlackConversationResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	inputsOld, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	conversationOld := k.ToSlackConversationInput(inputsOld["__inputs"].ObjectValue())
	conversationNew := k.ToSlackConversationInput(inputsNew)

	err = k.updateConversation(
		req.Id,
		conversationOld,
		conversationNew,
	)
	if err != nil {
		return nil, err
	}

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputsNew)

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (k *SlackConversationResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Update is not yet implemented for "+k.Name())
}

func (c *SlackConversationResource) updateConversation(
	id string,
	old SlackConversationInput,
	new SlackConversationInput) (err error) {

	if old.IsPrivate != new.IsPrivate {
		return fmt.Errorf("cannot change conversation type (private/public), only admin can do it")
	}

	token, err := c.config.getSlackToken()
	if err != nil {
		return err
	}

	api := slack.New(*token)
	if old.Name != new.Name {
		_, err = api.RenameConversation(id, new.Name)
	}
	if err != nil {
		return err
	}

	if old.Topic != new.Topic {
		_, err = api.SetTopicOfConversation(id, new.Topic)
	}
	if err != nil {
		return err
	}

	if old.Purpose != new.Purpose {
		_, err = api.SetPurposeOfConversation(id, new.Purpose)
	}
	if err != nil {
		return err
	}

	if old.IsArchived != new.IsArchived {
		if new.IsArchived {
			err = api.UnArchiveConversation(id)
		} else {
			err = api.ArchiveConversation(id)
		}
	}
	if err != nil {
		return err
	}

	// add new users
	for _, member := range new.Members {
		if !findStringInsideArray(old.Members, member) {
			_, err = api.InviteUsersToConversation(id, member)
			if err != nil {
				return fmt.Errorf("cannot add member [%s] to conversation [%s]: %v", member, id, err)
			}
		}
	}

	// remove users
	for _, member := range old.Members {
		if !findStringInsideArray(new.Members, member) {
			err = api.KickUserFromConversation(id, member)
			if err != nil {
				return fmt.Errorf("cannot remove member [%s] to conversation [%s]: %v", member, id, err)
			}
		}
	}

	return nil
}

func (c *SlackConversationResource) createConversation(input SlackConversationInput) (*string, error) {
	token, err := c.config.getSlackToken()
	if err != nil {
		return nil, err
	}

	api := slack.New(*token)
	channel, err := api.CreateConversation(input.Name, input.IsPrivate)
	if err != nil {
		return nil, err
	}
	logger.V(9).Infof("Created channel id [%s] with name [%s]\n", channel.ID, channel.Name)

	for _, memberId := range input.Members {
		channel, err = api.InviteUsersToConversation(channel.ID, memberId)
		if err != nil {
			return nil, err
		}
		logger.V(9).Infof("Invited user [%s] to channel channel id [%s]%s\n", memberId, channel.ID, channel.Name)
	}

	if len(input.Members) == 0 && input.IsPrivate {
		logger.V(3).Infof("No members invited for channel id [%s] with name [%s]\n", channel.ID, input.Name)
	}

	if input.Purpose != "" {
		channel, err = api.SetPurposeOfConversation(channel.ID, input.Purpose)
		if err != nil {
			return nil, err
		}
		logger.V(9).Infof("Set channel purpose id [%s] with [%s]\n", channel.ID, input.Purpose)
	}

	if input.Topic != "" {
		channel, err = api.SetTopicOfConversation(channel.ID, input.Topic)
		if err != nil {
			return nil, err
		}
		logger.V(9).Infof("Set channel topic id [%s] with [%s]\n", channel.ID, input.Topic)
	}

	if channel.IsArchived {
		err = api.ArchiveConversation(channel.ID)
		if err != nil {
			return nil, err
		}
		logger.V(9).Infof("Set channel archive id [%s] with [%t]\n", channel.ID, input.IsArchived)
	}

	return &channel.ID, nil
}

func (c *SlackConversationResource) deleteConversation(channelId string) (*slack.Channel, error) {
	token, err := c.config.getSlackToken()
	if err != nil {
		return nil, err
	}

	api := slack.New(*token)
	channel, err := api.GetConversationInfo(channelId, true)
	if err != nil {
		return nil, fmt.Errorf("error getting conversation information: %v", err)
	}

	deprecatedName := fmt.Sprintf(DEPRECATION_CHANNEL_NAME_FORMAT, channel.Name, time.Now().Format("2006-01-02-15-04-05"))
	channel, err = api.RenameConversation(channelId, deprecatedName)
	if err != nil {
		return nil, fmt.Errorf("error renamming conversation from [%s] to [%s]: %v", channel.Name, deprecatedName, err)
	}

	err = api.ArchiveConversation(channelId)
	if err != nil {
		return nil, fmt.Errorf("error archiving conversation [%s]: %v", deprecatedName, err)
	}

	return channel, err
}
