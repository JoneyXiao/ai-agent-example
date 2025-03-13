# AI Agent - ReAct

## Overview

This project implements an AI agent using the ReAct (Reasoning and Acting) paradigm, which enables the AI to reason about tasks and take actions using external tools. It leverages large language models to process natural language queries and interact with real-world data through a structured reasoning framework.

In this implementation, the agent can:
1. Understand natural language queries
2. Reason about how to solve a problem
3. Take appropriate actions using available tools
4. Observe the results of those actions
5. Continue reasoning until reaching a final answer

## Features

- **ReAct Paradigm**: Implements the Reasoning-Action-Observation loop
- **Weather Tool Integration**: Allows querying weather information for specific locations
- **Flexible Architecture**: Easily extendable to add more tools
- **Conversation History Management**: Maintains context throughout interactions
- **Regex-based Response Parsing**: Extracts actions and final answers from LLM responses

## Requirements

- Go 1.21.3 or higher
- OpenAI API key or DashScope API key (Alibaba Cloud)

## Setup

1. Clone the repository:
   ```
   git clone https://github.com/your-username/agent-example.git
   cd agent-example
   ```

2. Copy the environment template and add your API keys:
   ```
   cp .env.template .env
   ```

3. Edit the `.env` file with your API keys:
   ```
   DASH_SCOPE_API_KEY="your-dashscope-api-key"
   DASH_SCOPE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
   DASH_SCOPE_MODEL="qwen-turbo"
   ```

4. Install dependencies:
   ```
   go mod tidy
   ```

## Usage

Run the application:
```
go run main.go
```

The default query asks about the current weather in Shenzhen and whether it's suitable for outdoor activities. You can modify the query in the `main.go` file.

## How It Works

1. The agent receives a natural language query
2. It follows the ReAct format:
   - **Thought**: Reasoning about what to do
   - **Action**: Which tool to use
   - **Action Input**: Parameters for the tool
   - **Observation**: Result from the tool
   - (Steps repeat until a conclusion is reached)
   - **Final Answer**: The response to the original query

## Project Structure

- `main.go`: Main application logic and ReAct implementation
- `tools/weather.go`: Weather tool implementation
- `prompt/prompt.go`: Template for structuring agent prompts
- `.env`: Configuration file for API keys

## Extending with New Tools

To add a new tool:
1. Create a new file in the `tools` directory
2. Implement the tool interface
3. Register the tool in `main.go`

## License

This project is licensed under the terms of the included LICENSE file.

## Acknowledgements

- ReAct paper: [ReAct: Synergizing Reasoning and Acting in Language Models](https://arxiv.org/abs/2210.03629)
- DashScope API (Alibaba Cloud)
