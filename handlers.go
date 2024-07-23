package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minherz/smartagent/metadata"
	"github.com/minherz/smartagent/utils"
)

const (
	ModelNameEnvVar = "GEMINI_MODEL_NAME"
	// from https://cloud.google.com/vertex-ai/generative-ai/docs/learn/model-versions
	DefaultModelName = "gemini-1.5-flash-001"
)

// ReturnStatus provides addition information about response
type ReturnStatus struct {
	Error   string      `json:"error,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type AskRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
}

type AskResponse struct {
	SessionID string `json:"sessionId,omitempty"`
	Response  string `json:"response,omitempty"`
}

type Agent struct {
	VertexClient *genai.Client
	Model        *genai.GenerativeModel
	Sessions     map[string]*ChatSession
}

type ChatSession struct {
	ID      string
	Session *genai.ChatSession
}

func NewAgent(ctx context.Context) (*Agent, error) {
	var (
		projectID, region string
		err               error
	)
	if projectID, err = metadata.ProjectID(ctx); err != nil {
		return nil, fmt.Errorf("could not retrieve current project ID: %w", err)
	}
	if region, err = metadata.Region(ctx); err != nil {
		return nil, fmt.Errorf("could not retrieve current region: %w", err)
	}
	agent := &Agent{Sessions: make(map[string]*ChatSession)}
	if agent.VertexClient, err = genai.NewClient(ctx, projectID, region); err != nil {
		return nil, fmt.Errorf("could not initialize Vertex AI client: %w", err)
	}
	modelName := utils.GetenvWithDefault(ModelNameEnvVar, DefaultModelName)
	agent.Model = agent.VertexClient.GenerativeModel(modelName)
	return agent, nil
}

func (a *Agent) Close() {
	if a.VertexClient != nil {
		a.Close()
	}
}

func (a *Agent) OnAsk(ectx echo.Context) error {
	input := &AskRequest{}
	if err := ectx.Bind(&input); err != nil {
		Logger.Error("failed to parse input", "error", fmt.Sprintf("%v", err))
		return ectx.JSON(http.StatusBadRequest, ReturnStatus{Error: fmt.Sprintf("invalid input: %q", err)})
	}
	if input.Prompt == "" {
		Logger.Error("prompt is empty")
		return ectx.JSON(http.StatusBadRequest, ReturnStatus{Error: "prompt is empty"})
	}
	if input.SessionID == "" {
		if ID, err := uuid.NewRandom(); err != nil {
			Logger.Error("failed to generate session ID", "error", fmt.Sprintf("%v", err))
			return ectx.JSON(http.StatusInternalServerError, ReturnStatus{Error: fmt.Sprintf("failed to generate session ID: %q", err)})
		} else {
			input.SessionID = ID.String()
		}
		session := &ChatSession{ID: input.SessionID, Session: a.Model.StartChat()}
		// TODO: check for already existing session
		a.Sessions[input.SessionID] = session
	}
	s := a.Sessions[input.SessionID]
	response, err := s.Session.SendMessage(ectx.Request().Context(), genai.Text(input.Prompt))
	if err != nil {
		Logger.Error("chat response error", "error", fmt.Sprintf("%v", err))
		return ectx.JSON(http.StatusInternalServerError, ReturnStatus{Error: fmt.Sprintf("chat response error: %q", err)})
	}
	if len(response.Candidates) == 0 {
		return ectx.JSON(http.StatusOK, ReturnStatus{Payload: AskResponse{SessionID: input.SessionID, Response: "<empty>"}})
	}
	return ectx.JSON(http.StatusOK, ReturnStatus{Payload: AskResponse{SessionID: input.SessionID, Response: composeResponse(response.Candidates[0])}})
}

func composeResponse(candidate *genai.Candidate) string {
	totalParts := len(candidate.Content.Parts)
	if totalParts == 0 {
		return "<empty>"
	}
	// convert []genai.Part to []string through []genai.Text
	// TODO: skip genai.Part that aren't genai.Text
	texts := make([]string, totalParts)
	for i := range candidate.Content.Parts {
		texts[i] = string(candidate.Content.Parts[i].(genai.Text))
	}
	// concatenate as strings
	return strings.Join(texts, ". ")
}
