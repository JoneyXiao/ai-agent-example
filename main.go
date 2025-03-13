package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	tools "agent-example/tools"

	prompt "agent-example/prompt"

	godotenv "github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// ChatMessages represents a collection of chat messages for OpenAI API
type ChatMessages []*ChatMessage

type ChatMessage struct {
	Msg openai.ChatCompletionMessage
}

// MessageStore holds the conversation history for the current session
var MessageStore ChatMessages

// Constants for message roles to improve code readability and avoid typos
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

func (cm *ChatMessages) Clear() {
	*cm = make([]*ChatMessage, 0)
	cm.AddSystem(prompt.SystemPrompt)
}

// init loads environment variables from .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	MessageStore = make(ChatMessages, 0)
	MessageStore.Clear()
}

func NewOpenAiClient() *openai.Client {
	apiKey := os.Getenv("DASH_SCOPE_API_KEY")
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = os.Getenv("DASH_SCOPE_URL")

	return openai.NewClientWithConfig(config)
}

// AppendMessage adds a new message to the chat history with specified role and content
func (cm *ChatMessages) AppendMessage(msg string, role string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:    role,
			Content: msg,
		},
	})
}

// GetMessages returns a copy of all messages in the chat history
// This creates a new slice to avoid modifying the original messages
func (cm *ChatMessages) GetMessages() []openai.ChatCompletionMessage {
	ret := make([]openai.ChatCompletionMessage, len(*cm))
	for i, msg := range *cm {
		ret[i] = msg.Msg
	}
	return ret
}

// AddToolCall adds a tool call to the chat history
func (cm *ChatMessages) AddToolCall(rsp openai.ChatCompletionMessage, role string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:         role,
			Content:      rsp.Content,
			FunctionCall: rsp.FunctionCall,
			ToolCalls:    rsp.ToolCalls,
		},
	})
}

// AddSystem adds a system message to the chat history
func (cm *ChatMessages) AddSystem(msg string) {
	cm.AppendMessage(msg, RoleSystem)
}

// AddAssistant adds an assistant message to the chat history
func (cm *ChatMessages) AddAssistant(rsp openai.ChatCompletionMessage) {
	cm.AddToolCall(rsp, RoleAssistant)
}

// AddUser adds a user message to the chat history
func (cm *ChatMessages) AddUser(msg string) {
	cm.AppendMessage(msg, RoleUser)
}

// AddTool adds a tool to the chat history
func (cm *ChatMessages) AddTool(msg string, name string, toolCallId string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:       RoleTool,
			Content:    msg,
			Name:       name,
			ToolCallID: toolCallId,
		},
	})
}

func (cm *ChatMessages) GetLast() string {
	if len(*cm) == 0 {
		return "No messages in the chat history"
	}

	return (*cm)[len(*cm)-1].Msg.Content
}

// Chat sends the message history to the AI model and returns the response
// It handles any errors that might occur during the API call
func Chat(message []openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	// Create a new OpenAI client for this request
	client := NewOpenAiClient()

	// Send the chat completion request to the API
	rsp, err := client.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:    os.Getenv("DASH_SCOPE_MODEL"),
		Messages: message,
	})
	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	// Return the first (and typically only) message from the choices
	return rsp.Choices[0].Message
}

// ChatWithTools sends the message history to the AI model with tools and returns the response
// It handles any errors that might occur during the API call
func ChatWithTools(message []openai.ChatCompletionMessage, tools []openai.Tool) openai.ChatCompletionMessage {
	// Create a new OpenAI client for this request
	client := NewOpenAiClient()

	// Send the chat completion request to the API
	rsp, err := client.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:      os.Getenv("DASH_SCOPE_MODEL"),
		Messages:   message,
		Tools:      tools,
		ToolChoice: "auto",
	})

	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	// Return the first (and typically only) message from the choices
	return rsp.Choices[0].Message
}

func main() {
	query := "深圳现在的天气怎么样？请告诉我当前的温度和天气状况，以及是否适合出去游玩。"

	weatherTool := tools.WeatherToolName + ":" + tools.WeatherToolDescription + "\nparameters: \n" + tools.WeatherToolParameters
	toolsList := make([]string, 0)
	toolsList = append(toolsList, weatherTool)
	toolNames := make([]string, 0)
	toolNames = append(toolNames, tools.WeatherToolName)

	prompt := fmt.Sprintf(prompt.Template, toolsList, toolNames, query)
	fmt.Println("Prompt: ", prompt)

	MessageStore.AddUser(prompt)

	maxLoops := 5
	loopCount := 1

	for {
		fmt.Printf("\n---------------- The %d round response ----------------\n", loopCount)
		fmt.Println("# Message Store Debug:")
		messages := MessageStore.GetMessages()
		fmt.Printf("Number of messages: %d\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("Message %d: Role=%s, Content length=%d\n", i, msg.Role, len(msg.Content))
		}

		first_response := Chat(MessageStore.GetMessages())
		fmt.Println("# Response from llm:")
		fmt.Println(first_response.Content)
		fmt.Println()

		regexPattern := regexp.MustCompile(`Final Answer:\s*(.*)`)
		finalAnswer := regexPattern.FindStringSubmatch(first_response.Content)
		if len(finalAnswer) > 0 {
			fmt.Println("# Final Answer from llm:")
			fmt.Println(first_response.Content)
			fmt.Println("--------------------------------------------------------")
			break
		}

		// Check if max loops reached
		if loopCount >= maxLoops {
			fmt.Println("Exceeded maximum number of reasoning loops. Stopping execution.")
			break
		}

		MessageStore.AddAssistant(first_response)

		regexAction := regexp.MustCompile(`Action:\s*(.*?)(?:$|[\n\r])`)
		regexActionInput := regexp.MustCompile(`Action Input:\s*(.*?)(?:$|[\n\r])`)
		action := regexAction.FindStringSubmatch(first_response.Content)
		actionInput := regexActionInput.FindStringSubmatch(first_response.Content)

		if len(action) > 1 && len(actionInput) > 1 {
			result := ""

			if action[1] == tools.WeatherToolName {
				var weatherParams tools.WeatherParams
				err := json.Unmarshal([]byte(actionInput[1]), &weatherParams)
				if err != nil {
					result = "Error parsing weather parameters: " + err.Error()
				} else {
					result, _ = tools.GetWeather(weatherParams)
				}
			}
			fmt.Println("# Action Debug:")
			fmt.Println("Action:")
			fmt.Println(action[1])
			fmt.Println("Action Input:")
			fmt.Println(actionInput[1])
			fmt.Println("Result:")
			fmt.Println(result)

			// Add the observation as a user message instead of a tool message
			Observation := "Observation: " + result
			prompt = first_response.Content + Observation

			fmt.Printf("# The %d round user prompt:\n", loopCount)
			fmt.Println(prompt)
			MessageStore.AddUser(prompt)
		}
		loopCount++
	}
}
