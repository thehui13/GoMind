package aihelper

import (
	"GoMind/common/rag"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type StreamCallback func(msg string)

// AIModel 定义AI模型接口
type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	GetModelType() string
}

// =================== OpenAI 实现 ===================
type OpenAIModel struct {
	llm model.ToolCallingChatModel
}

func NewOpenAIModel(ctx context.Context) (*OpenAIModel, error) {
	key := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai model failed: %v", err)
	}
	return &OpenAIModel{llm: llm}, nil
}

func (o *OpenAIModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai generate failed: %v", err)
	}
	return resp, nil
}

func (o *OpenAIModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("openai stream failed: %v", err)
	}
	defer stream.Close()

	var fullResp strings.Builder

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("openai stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content) // 聚合

			cb(msg.Content) // 实时调用cb函数，方便主动发送给前端
		}
	}

	return fullResp.String(), nil //返回完整内容，方便后续存储
}

func (o *OpenAIModel) GetModelType() string { return "1" }

// =================== Ollama 实现 ===================

// OllamaModel Ollama模型实现
type OllamaModel struct {
	llm model.ToolCallingChatModel
}

func NewOllamaModel(ctx context.Context, baseURL, modelName string) (*OllamaModel, error) {
	llm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("create ollama model failed: %v", err)
	}
	return &OllamaModel{llm: llm}, nil
}

func (o *OllamaModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("ollama generate failed: %v", err)
	}
	return resp, nil
}

func (o *OllamaModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("ollama stream failed: %v", err)
	}
	defer stream.Close()
	var fullResp strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("openai stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content) // 聚合
			cb(msg.Content)                   // 实时调用cb函数，方便主动发送给前端
		}
	}
	return fullResp.String(), nil //返回完整内容，方便后续存储
}

func (o *OllamaModel) GetModelType() string { return "2" }

// AIToolCall 表示AI工具调用请求
type AIToolCall struct {
	IsToolCall bool                   `json:"isToolCall"`
	ToolName   string                 `json:"toolName"`
	Args       map[string]interface{} `json:"args"`
}

// =================== Gemini 实现（集成 RAG 和 MCP）===================

// GeminiModel Gemini模型实现，使用OpenAI兼容API调用Gemini，集成RAG和MCP功能
type GeminiModel struct {
	llm        model.ToolCallingChatModel
	username   string
	mcpClient  *client.Client
	mcpBaseURL string
}

func NewGeminiModel(ctx context.Context, username string) (*GeminiModel, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	modelName := os.Getenv("GEMINI_MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	baseURL := "https://generativelanguage.googleapis.com/v1beta/"

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini model failed: %v", err)
	}

	return &GeminiModel{
		llm:        llm,
		username:   username,
		mcpBaseURL: "http://localhost:8081/mcp",
	}, nil
}

// GenerateResponse 生成响应，集成RAG和MCP功能
func (g *GeminiModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	// 获取用户最后一条消息作为查询
	lastMessage := messages[len(messages)-1]
	query := lastMessage.Content

	// 1. 尝试RAG检索
	var ragPrompt string
	ragQuery, err := rag.NewRAGQuery(ctx, g.username)
	if err != nil {
		log.Printf("Failed to create RAG query (user may not have uploaded file): %v", err)
		ragPrompt = query // 没有RAG，使用原始查询
	} else {
		docs, err := ragQuery.RetrieveDocuments(ctx, query)
		if err != nil {
			log.Printf("Failed to retrieve documents: %v", err)
			ragPrompt = query
		} else {
			ragPrompt = rag.BuildRAGPrompt(query, docs)
		}
	}

	// 2. 尝试MCP工具调用
	return g.generateWithMCP(ctx, messages, ragPrompt)
}

// generateWithMCP 使用MCP工具生成响应
func (g *GeminiModel) generateWithMCP(ctx context.Context, messages []*schema.Message, query string) (*schema.Message, error) {
	// 第一次调用AI：告诉AI使用固定的JSON格式
	firstPrompt := g.buildFirstPrompt(query)
	firstMessages := make([]*schema.Message, len(messages))
	copy(firstMessages, messages)
	firstMessages[len(firstMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: firstPrompt,
	}

	// 调用LLM生成第一次响应
	firstResp, err := g.llm.Generate(ctx, firstMessages)
	if err != nil {
		return nil, fmt.Errorf("gemini first generate failed: %v", err)
	}

	// 解析AI响应
	aiResult := firstResp.Content
	toolCall, err := g.parseAIResponse(aiResult)
	if err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		return firstResp, nil
	}

	// 情况1：AI不调用工具，直接返回响应
	if !toolCall.IsToolCall {
		return firstResp, nil
	}

	// 情况2：AI要调用工具
	var toolResult string

	// 获取MCP客户端
	mcpClient, mcpErr := g.getMCPClient(ctx)
	if mcpErr != nil {
		log.Printf("MCP client error: %v, using mock data", mcpErr)
		// MCP服务不可用，使用模拟数据
		toolResult = g.getMockWeatherData(toolCall.Args)
	} else {
		// 调用MCP工具
		toolResult, mcpErr = g.callMCPTool(ctx, mcpClient, toolCall.ToolName, toolCall.Args)
		if mcpErr != nil {
			log.Printf("MCP tool call failed: %v, using mock data", mcpErr)
			toolResult = g.getMockWeatherData(toolCall.Args)
		}
	}

	// 第二次调用AI：将工具结果告诉AI
	secondPrompt := g.buildSecondPrompt(query, toolCall.ToolName, toolCall.Args, toolResult)
	secondMessages := make([]*schema.Message, len(messages))
	copy(secondMessages, messages)
	secondMessages[len(secondMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: secondPrompt,
	}

	// 调用LLM生成最终响应
	finalResp, err := g.llm.Generate(ctx, secondMessages)
	if err != nil {
		return nil, fmt.Errorf("gemini second generate failed: %v", err)
	}

	return finalResp, nil
}

// StreamResponse 流式响应，集成RAG和MCP功能
func (g *GeminiModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	// 获取用户最后一条消息作为查询
	lastMessage := messages[len(messages)-1]
	query := lastMessage.Content

	// 1. 尝试RAG检索
	var ragPrompt string
	ragQuery, err := rag.NewRAGQuery(ctx, g.username)
	if err != nil {
		log.Printf("Failed to create RAG query (user may not have uploaded file): %v", err)
		ragPrompt = query
	} else {
		docs, err := ragQuery.RetrieveDocuments(ctx, query)
		if err != nil {
			log.Printf("Failed to retrieve documents: %v", err)
			ragPrompt = query
		} else {
			ragPrompt = rag.BuildRAGPrompt(query, docs)
		}
	}

	// 2. 尝试MCP工具调用
	return g.streamWithMCP(ctx, messages, ragPrompt, cb)
}

// streamWithMCP 使用MCP工具生成流式响应
func (g *GeminiModel) streamWithMCP(ctx context.Context, messages []*schema.Message, query string, cb StreamCallback) (string, error) {
	// 第一次调用AI：告诉AI使用固定的JSON格式
	firstPrompt := g.buildFirstPrompt(query)
	firstMessages := make([]*schema.Message, len(messages))
	copy(firstMessages, messages)
	firstMessages[len(firstMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: firstPrompt,
	}

	// 第一次调用使用同步接口（非流式）
	firstResp, err := g.llm.Generate(ctx, firstMessages)
	if err != nil {
		return "", fmt.Errorf("gemini first generate failed: %v", err)
	}

	aiResult := firstResp.Content
	toolCall, err := g.parseAIResponse(aiResult)
	if err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		// 通过回调发送响应
		cb(aiResult)
		return aiResult, nil
	}

	// 情况1：AI不调用工具，直接返回响应
	if !toolCall.IsToolCall {
		// 通过回调发送响应
		cb(aiResult)
		return aiResult, nil
	}

	// 情况2：AI要调用工具
	var toolResult string

	// 获取MCP客户端
	mcpClient, err := g.getMCPClient(ctx)
	if err != nil {
		log.Printf("MCP client error: %v, using mock data", err)
		toolResult = g.getMockWeatherData(toolCall.Args)
	} else {
		// 调用MCP工具
		toolResult, err = g.callMCPTool(ctx, mcpClient, toolCall.ToolName, toolCall.Args)
		if err != nil {
			log.Printf("MCP tool call failed: %v, using mock data", err)
			toolResult = g.getMockWeatherData(toolCall.Args)
		}
	}

	// 第二次调用AI：将工具结果告诉AI，使用流式接口
	secondPrompt := g.buildSecondPrompt(query, toolCall.ToolName, toolCall.Args, toolResult)
	secondMessages := make([]*schema.Message, len(messages))
	copy(secondMessages, messages)
	secondMessages[len(secondMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: secondPrompt,
	}

	// 调用LLM生成最终响应（流式）
	stream, err := g.llm.Stream(ctx, secondMessages)
	if err != nil {
		return "", fmt.Errorf("gemini second stream failed: %v", err)
	}
	defer stream.Close()

	var finalResp strings.Builder

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("gemini stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			finalResp.WriteString(msg.Content)
			cb(msg.Content)
		}
	}

	return finalResp.String(), nil
}

// getMCPClient 获取或创建MCP客户端
func (g *GeminiModel) getMCPClient(ctx context.Context) (*client.Client, error) {
	if g.mcpClient == nil {
		httpTransport, err := transport.NewStreamableHTTP(g.mcpBaseURL)
		if err != nil {
			return nil, fmt.Errorf("create mcp transport failed: %v", err)
		}

		g.mcpClient = client.NewClient(httpTransport)

		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = g.LATEST_PROTOCOL_VERSION()
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "Gemini-Go AIHelper Client",
			Version: "1.0.0",
		}
		initRequest.Params.Capabilities = mcp.ClientCapabilities{}

		if _, err := g.mcpClient.Initialize(ctx, initRequest); err != nil {
			return nil, fmt.Errorf("mcp client initialize failed: %v", err)
		}
	}
	return g.mcpClient, nil
}

// buildFirstPrompt 构建第一次调用的提示词
func (g *GeminiModel) buildFirstPrompt(query string) string {
	return fmt.Sprintf(`你是一个智能助手，可以调用工具来获取信息。

可用工具:
- get_weather: 获取指定城市的天气信息，参数: city（城市名称，支持中文和英文，如北京、Shanghai等）
- 知识库检索: 已自动为你检索相关文档内容

重要规则:
1. 如果需要调用工具，必须严格返回以下JSON格式：
{
  "isToolCall": true,
  "toolName": "工具名称",
  "args": {"参数名": "参数值"}
}
2. 如果不需要调用工具，直接返回自然语言回答
3. 请根据用户问题决定是否需要调用工具

用户问题: %s

请根据需要调用适当的工具，然后给出综合的回答。`, query)
}

// buildSecondPrompt 构建第二次调用的提示词
func (g *GeminiModel) buildSecondPrompt(query, toolName string, args map[string]interface{}, toolResult string) string {
	return fmt.Sprintf(`你是一个智能助手，可以调用工具来获取信息。

工具执行结果:
工具名称: %s
工具参数: %v
工具结果: %s

用户问题: %s

请根据工具结果和用户问题，给出最终的综合回答。`, toolName, args, toolResult, query)
}

// parseAIResponse 解析AI响应，检查是否包含工具调用
func (g *GeminiModel) parseAIResponse(response string) (*AIToolCall, error) {
	// 去除 markdown 代码块包装
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	var toolCall AIToolCall
	if err := json.Unmarshal([]byte(response), &toolCall); err == nil {
		return &toolCall, nil
	}

	if strings.Contains(response, "get_weather") {
		city := g.extractCityFromResponse(response)
		if city != "" {
			return &AIToolCall{
				IsToolCall: true,
				ToolName:   "get_weather",
				Args:       map[string]interface{}{"city": city},
			}, nil
		}
	}

	return &AIToolCall{IsToolCall: false}, nil
}

// callMCPTool 调用MCP工具
func (g *GeminiModel) callMCPTool(ctx context.Context, client *client.Client, toolName string, args map[string]interface{}) (string, error) {
	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	result, err := client.CallTool(ctx, callToolRequest)
	if err != nil {
		return "", fmt.Errorf("mcp tool call failed: %v", err)
	}

	var text string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			text += textContent.Text + "\n"
		}
	}

	return text, nil
}

// getMockWeatherData 获取模拟天气数据（当MCP服务不可用时使用）
func (g *GeminiModel) getMockWeatherData(args map[string]interface{}) string {
	city, _ := args["city"].(string)
	if city == "" {
		city = "未知城市"
	}

	mockData := map[string]string{
		"北京": "北京今日天气：晴，温度22-30°C，空气质量优",
		"上海": "上海今日天气：多云，温度24-32°C，空气质量良",
		"杭州": "杭州今日天气：晴转多云，温度25-33°C，空气质量优",
		"深圳": "深圳今日天气：阵雨，温度26-31°C，空气质量良",
		"广州": "广州今日天气：阴，温度27-34°C，空气质量良",
		"成都": "成都今日天气：小雨，温度18-24°C，空气质量轻度污染",
	}

	if data, ok := mockData[city]; ok {
		return data
	}

	return fmt.Sprintf("%s今日天气：晴，温度23-31°C，空气质量良", city)
}

// extractCityFromResponse 从响应中提取城市名称
func (g *GeminiModel) extractCityFromResponse(response string) string {
	var toolCall AIToolCall
	if err := json.Unmarshal([]byte(response), &toolCall); err == nil {
		if args, ok := toolCall.Args["city"].(string); ok {
			return args
		}
	}
	return ""
}

// GetModelType 获取模型类型
func (g *GeminiModel) GetModelType() string { return "3" }

// LATEST_PROTOCOL_VERSION MCP协议版本
func (g *GeminiModel) LATEST_PROTOCOL_VERSION() string {
	return "2024-01-01"
}
