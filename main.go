package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

const CONFIG_DIR = ".config/helpp"
const CONFIG_PERMISSION = 0755
const ENV_FILE = "env"
const ENV_PERMISSION = 0600

func init() {
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")
}

func main() {
	envFilePath, err := initUserEnv()
	exitError(err)

	if len(os.Args) < 2 {
		log.Fatalln("Not enough arguments, please provide a question")
	}

	apiConfig, err := newApiConfigFromEnv(envFilePath)
	exitError(err)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiConfig.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	exitError(err)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(
			"You are a CLI helping tool, give me very short (1-2 lines) answers to user questions,"+
				" with an example command/code snippet if applicable.",
			genai.RoleUser,
		),
	}

	result, err := client.Models.GenerateContent(
		ctx,
		apiConfig.modelName,
		genai.Text(strings.Join(os.Args[1:], " ")),
		config,
	)
	exitError(err)

	fmt.Println(result.Text())
}

func initUserEnv() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDirPath := filepath.Join(homeDir, CONFIG_DIR)
	envFilePath := filepath.Join(configDirPath, ENV_FILE)

	// Config exists, early return
	_, err = os.Stat(envFilePath)
	if !errors.Is(err, os.ErrNotExist) {
		return envFilePath, nil
	}

	if err := os.MkdirAll(configDirPath, CONFIG_PERMISSION); err != nil {
		return "", err
	}

	if err := setDefaultEnv(envFilePath); err != nil {
		return "", err
	}

	return envFilePath, nil
}

const ENV_API_KEY = "GEMINI_API_KEY"
const ENV_MODEL_NAME = "GEMINI_MODEL_NAME"

func setDefaultEnv(envFilePath string) error {
	defaultEnv := map[string]string{
		ENV_API_KEY:    "",
		ENV_MODEL_NAME: "gemini-2.0-flash",
	}
	env, err := godotenv.Marshal(defaultEnv)
	if err != nil {
		return err
	}

	return os.WriteFile(envFilePath, []byte(env), ENV_PERMISSION)
}

type ApiConfig struct {
	apiKey    string
	modelName string
}

func newApiConfigFromEnv(envFilePath string) (ApiConfig, error) {
	err := godotenv.Load(envFilePath)
	if err != nil {
		return ApiConfig{},
			fmt.Errorf("Error loading env file, please check the config path (%v)", envFilePath)
	}

	apiKey := os.Getenv(ENV_API_KEY)
	modelName := os.Getenv(ENV_MODEL_NAME)
	if apiKey == "" || modelName == "" {
		return ApiConfig{},
			fmt.Errorf("Either %v and/or %v isn't set, please check your config file (%v)", ENV_API_KEY, ENV_MODEL_NAME, envFilePath)
	}

	return ApiConfig{apiKey, modelName}, nil
}

func exitError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
