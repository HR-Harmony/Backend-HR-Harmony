package controllers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"os"
	"strings"
)

type BotResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

type HarmonyUsecase interface {
	RecommendHarmony(userInput, openAIKey string) (string, error)
}

type harmonyUsecase struct{}

func NewHarmonyUsecase() HarmonyUsecase {
	return &harmonyUsecase{}
}

func (uc *harmonyUsecase) RecommendHarmony(userInput, openAIKey string) (string, error) {
	ctx := context.Background()
	client := openai.NewClient(openAIKey)
	model := openai.GPT3Dot5Turbo
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Halo, perkenalkan saya sistem untuk rekomendasi tempat wisata",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userInput,
		},
	}

	resp, err := uc.getCompletionFromMessages(ctx, client, messages, model)
	if err != nil {
		return "", err
	}
	answer := resp.Choices[0].Message.Content
	return answer, nil
}

func (uc *harmonyUsecase) getCompletionFromMessages(
	ctx context.Context,
	client *openai.Client,
	messages []openai.ChatCompletionMessage,
	model string,
) (openai.ChatCompletionResponse, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	return resp, err
}

func RecommendTraining(c echo.Context, harmonyUsecase HarmonyUsecase) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Authorization token is missing"})
	}

	authParts := strings.SplitN(tokenString, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Invalid token format"})
	}

	tokenString = authParts[1]

	var requestData map[string]interface{}
	err := c.Bind(&requestData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Invalid JSON format"})
	}

	userInput, ok := requestData["message"].(string)
	if !ok || userInput == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Invalid or missing 'message' in the request"})
	}

	// Check if the user input contains keywords related to tourism
	trainingKeywords := []string{"materi", "materi apa saja", "materi apa yang ada", "materi apa yang tersedia", "materi apa yang bisa dipelajari", "materi apa yang bisa", "soal", "pelatihan", "kode", "coding", "belajar", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "belajar apa saja", "belajar apa yang ada", "belajar apa yang tersedia", "belajar apa yang bisa dipelajari", "belajar apa yang bisa", "belajar apa", "divisi", "bagian", "perusahaan"}
	containsTrainingKeyword := false
	for _, keyword := range trainingKeywords {
		if strings.Contains(strings.ToLower(userInput), keyword) {
			containsTrainingKeyword = true
			break
		}
	}

	if !containsTrainingKeyword {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Maaf, HRBot hanya bisa menjawab pertanyaan seputar pelatihan. Silakan coba lagi dengan pertanyaan yang berbeda"})
	}

	userInput = fmt.Sprintf("Training: %s", userInput)

	answer, err := harmonyUsecase.RecommendHarmony(userInput, os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		errorMessage := "Failed to generate question about training"
		if strings.Contains(err.Error(), "rate limits exceeded") {
			errorMessage = "Rate limits exceeded. Please try again later."
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": true, "message": errorMessage})
	}

	responseData := BotResponse{
		Status: "success",
		Data:   answer,
	}

	return c.JSON(http.StatusOK, responseData)
}
