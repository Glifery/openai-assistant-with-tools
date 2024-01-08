package openai_service

import (
	"context"
	"fmt"
	"github.com/glifery/openai-assistant-with-tools/types"
	"github.com/sashabaranov/go-openai"
	"time"
)

type AssistantConfig struct {
	Name         string
	Instructions string
	Token        string
}

type assistantService struct {
	config    AssistantConfig
	functions []types.FunctionCalling
	client    *openai.Client
	assistant *openai.Assistant
	thread    *openai.Thread
}

func NewAssistantService(config AssistantConfig, functions []types.FunctionCalling) *assistantService {
	client := openai.NewClient(config.Token)

	return &assistantService{
		config:    config,
		functions: functions,
		client:    client,
	}
}

func (a *assistantService) ConnectToExistingThread(ctx context.Context, assistantId string, threadId string) (err error) {
	assistant, err := a.client.RetrieveAssistant(ctx, assistantId)
	if err != nil {
		return
	}

	thread, err := a.client.RetrieveThread(ctx, threadId)
	if err != nil {
		return
	}

	a.assistant = &assistant
	a.thread = &thread

	return nil
}

//func (a *assistantService) CreateNewThread(ctx context.Context, assistantId string, threadId string) (err error) {
//	assistant, err := a.client.RetrieveAssistant(ctx, assistantId)
//	if err != nil {
//		return
//	}
//
//	existedThread, err := a.client.RetrieveThread(ctx, threadId)
//	if err != nil {
//		return
//	}
//
//	_, err = a.client.DeleteThread(ctx, existedThread.ID)
//
//	newThread, err := a.client.CreateThread(
//		context.Background(),
//		openai.ThreadRequest{},
//	)
//
//	a.assistant = &assistant
//	a.thread = &newThread
//
//	return nil
//}

func (a *assistantService) CreateFromScratch() (err error) {
	tools := []openai.AssistantTool{}
	for _, function := range a.functions {
		tools = append(tools, openai.AssistantTool{
			Type: openai.AssistantToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        function.GetFunctionName(),
				Description: function.GetFunctionDescription(),
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": function.GetFunctionProperties(),
					"required":   function.GetFunctionRequiredProperties(),
				},
			},
		})
	}

	assistant, err := a.client.CreateAssistant(
		context.Background(),
		openai.AssistantRequest{
			Name:         &a.config.Name,
			Instructions: &a.config.Instructions,
			Tools:        tools,
			Model:        openai.GPT3Dot5Turbo1106,
		},
	)

	if err != nil {
		return
	}

	thread, err := a.client.CreateThread(
		context.Background(),
		openai.ThreadRequest{},
	)

	a.assistant = &assistant
	a.thread = &thread

	fmt.Printf("Assistant ID: %s\n", assistant.ID)
	fmt.Printf("Thread ID: %s\n", thread.ID)

	return nil
}

func (a *assistantService) SendMessageAndRun(ctx context.Context, message string) (run openai.Run, err error) {
	_, err = a.client.CreateMessage(ctx, a.thread.ID, openai.MessageRequest{
		Role:    "user",
		Content: message,
		FileIds: nil,
	})
	if err != nil {
		return
	}

	run, err = a.client.CreateRun(ctx, a.thread.ID, openai.RunRequest{
		AssistantID: a.assistant.ID,
	})

	return
}

func (a *assistantService) GetRunResponse(ctx context.Context, runInput *openai.Run) (messages openai.MessagesList, err error) {
	for {
		run, err := a.client.RetrieveRun(ctx, a.thread.ID, runInput.ID)
		if err != nil {
			return openai.MessagesList{}, err
		}

		fmt.Printf("Run status: %s\n", run.Status)

		if run.Status == "requires_action" {
			toolOutputs := []openai.ToolOutput{}
			for _, toolCall := range run.RequiredAction.SubmitToolOutputs.ToolCalls {
				toolOutput, err := a.makeToolCall(toolCall)
				if err != nil {
					return openai.MessagesList{}, err
				}
				toolOutputs = append(toolOutputs, *toolOutput)
			}

			run, err = a.client.SubmitToolOutputs(ctx, a.thread.ID, run.ID, openai.SubmitToolOutputsRequest{
				ToolOutputs: toolOutputs,
			})
			if err != nil {
				return openai.MessagesList{}, err
			}

			//toolCallId := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].ID
			//name := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].Function.Name
			//argumentsString := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].Function.Arguments
			//
			//for _, function := range a.functions {
			//	if function.GetFunctionName() != name {
			//		continue
			//	}
			//
			//	output, err := function.Execute(argumentsString)
			//	if err != nil {
			//		return openai.MessagesList{}, err
			//	}
			//	run, err = a.client.SubmitToolOutputs(ctx, a.thread.ID, run.ID, openai.SubmitToolOutputsRequest{
			//		ToolOutputs: []openai.ToolOutput{
			//			{
			//				ToolCallID: toolCallId,
			//				Output:     output,
			//			},
			//		},
			//	})
			//	if err != nil {
			//		return openai.MessagesList{}, err
			//	}
			//}
		}

		if run.Status == "completed" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	order := "asc"
	messages, err = a.client.ListMessage(ctx, a.thread.ID, nil, &order, nil, nil)

	return
}

func (a *assistantService) makeToolCall(toolCall openai.ToolCall) (toolOutput *openai.ToolOutput, err error) {
	toolCallId := toolCall.ID
	name := toolCall.Function.Name
	argumentsString := toolCall.Function.Arguments

	fmt.Printf("  tool required: %s (%s)\n", name, argumentsString)

	function, err := a.findFunction(name)
	if err != nil {
		return nil, err
	}

	output, err := (*function).Execute(argumentsString)
	if err != nil {
		return nil, err
	}

	return &openai.ToolOutput{
		ToolCallID: toolCallId,
		Output:     output,
	}, nil
}

func (a *assistantService) findFunction(name string) (functionCalling *types.FunctionCalling, err error) {
	for _, function := range a.functions {
		if function.GetFunctionName() == name {
			return &function, nil
		}
	}

	return nil, fmt.Errorf("function with name %s not found", name)
}
