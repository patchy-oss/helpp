package main

import (
	"context"
	"errors"
	"flag"
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

type DetailsLevel int

const (
	Default DetailsLevel = iota
	Detailed
	MoreDetailed
	FullDetails
)

func main() {
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %v [-d|-dd|-ddd] [QUESTION...]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "NOTE: You can specify only one of d flags.\n")
		os.Exit(1)
	}
	var dFlag, ddFlag, dddFlag bool
	flag.BoolVar(&dFlag, "d", false, "ask model to include more details.")
	flag.BoolVar(&ddFlag, "dd", false, "ask model to include even more details.")
	flag.BoolVar(&dddFlag, "ddd", false, "ask model to include as much details as possible (default Gemini settings).")
	flag.Parse()

	modelReplyDetailsLevel := Default
	switch {
	case dFlag:
		modelReplyDetailsLevel = Detailed
	case ddFlag:
		modelReplyDetailsLevel = MoreDetailed
	case dddFlag:
		modelReplyDetailsLevel = FullDetails
	}

	envFilePath, err := initUserEnv()
	exitError(err)

	if len(flag.Args()) == 0 {
		flag.Usage()
	}

	apiConfig, err := newApiConfigFromEnv(envFilePath, modelReplyDetailsLevel)
	exitError(err)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiConfig.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	exitError(err)

	result, err := client.Models.GenerateContent(
		ctx,
		apiConfig.modelName,
		genai.Text(strings.Join(os.Args[1:], " ")),
		apiConfig.modelConfig,
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
	apiKey      string
	modelName   string
	modelConfig *genai.GenerateContentConfig
}

func newApiConfigFromEnv(envFilePath string, detailsLevel DetailsLevel) (ApiConfig, error) {
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

	var systemInstructions *genai.Content
	switch detailsLevel {
	case Default:
		systemInstructions = genai.NewContentFromText(
			"You are a CLI helping tool, give me very short (1-2 lines) answers to user questions,"+
				" with an example command/code snippet if applicable.",
			genai.RoleUser,
		)
	case Detailed:
		systemInstructions = genai.NewContentFromText(
			"You are a CLI helping tool, give me short (3-5 lines) answer to user question,"+
				" with an example command/code snippet if applicable.",
			genai.RoleUser,
		)
	case MoreDetailed:
		systemInstructions = genai.NewContentFromText(
			"You are a CLI helping tool, give me somewhat detailed (7-10 lines) answer to user question,"+
				" with an example command/code snippets",
			genai.RoleUser,
		)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstructions,
	}

	return ApiConfig{apiKey, modelName, config}, nil
}

func exitError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
