package webauthn

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.uber.org/zap"
)

type Handler struct {
	logger   *zap.Logger
	webauthn *webauthn.WebAuthn
}

// NewHandler creates a new WebAuthn handler
func NewHandler(logger *zap.Logger, config *webauthn.Config) (*Handler, error) {
	w, err := webauthn.New(config)
	if err != nil {
		return nil, err
	}

	return &Handler{
		logger:   logger,
		webauthn: w,
	}, nil
}

// BeginRegistration starts the WebAuthn registration process
func (h *Handler) BeginRegistration(c *gin.Context) {
	user := getUserFromContext(c) // This would be implemented to get user from your auth system

	options, session, err := h.webauthn.BeginRegistration(user)
	if err != nil {
		h.logger.Error("Failed to begin registration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin registration"})
		return
	}

	// Store the session data
	storeSessionData(c, session)

	c.JSON(http.StatusOK, options)
}

// FinishRegistration completes the WebAuthn registration process
func (h *Handler) FinishRegistration(c *gin.Context) {
	user := getUserFromContext(c)
	session := getSessionData(c)

	credential, err := h.webauthn.FinishRegistration(user, session, c.Request)
	if err != nil {
		h.logger.Error("Failed to finish registration", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to finish registration"})
		return
	}

	// Store the credential
	if err := storeCredential(user, credential); err != nil {
		h.logger.Error("Failed to store credential", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

// BeginLogin starts the WebAuthn authentication process
func (h *Handler) BeginLogin(c *gin.Context) {
	user := getUserFromContext(c)

	options, session, err := h.webauthn.BeginLogin(user)
	if err != nil {
		h.logger.Error("Failed to begin login", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin login"})
		return
	}

	// Store the session data
	storeSessionData(c, session)

	c.JSON(http.StatusOK, options)
}

// FinishLogin completes the WebAuthn authentication process
func (h *Handler) FinishLogin(c *gin.Context) {
	user := getUserFromContext(c)
	session := getSessionData(c)

	credential, err := h.webauthn.FinishLogin(user, session, c.Request)
	if err != nil {
		h.logger.Error("Failed to finish login", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to finish login"})
		return
	}

	// Verify the credential
	if err := verifyCredential(user, credential); err != nil {
		h.logger.Error("Failed to verify credential", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credential"})
		return
	}

	// Generate session token
	token, err := generateSessionToken(user)
	if err != nil {
		h.logger.Error("Failed to generate session token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// Helper functions (to be implemented based on your storage and auth system)
func getUserFromContext(c *gin.Context) interface{} {
	// TODO: Implement user retrieval from context
	return nil
}

func storeSessionData(c *gin.Context, session *webauthn.SessionData) {
	// TODO: Implement session storage
}

func getSessionData(c *gin.Context) *webauthn.SessionData {
	// TODO: Implement session retrieval
	return nil
}

func storeCredential(user interface{}, credential *webauthn.Credential) error {
	// TODO: Implement credential storage
	return nil
}

func verifyCredential(user interface{}, credential *webauthn.Credential) error {
	// TODO: Implement credential verification
	return nil
}

func generateSessionToken(user interface{}) (string, error) {
	// TODO: Implement session token generation
	return "", nil
} 