package main

import (
	"context"
	"fmt"
	"log"
	// "math"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ollama/ollama/api"
)

var messageHistory = &History{{
	Role:    "system",
	Content: "Be brief. Format all response with markdown",
}}
