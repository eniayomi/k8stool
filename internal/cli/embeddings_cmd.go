package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8stool/internal/embeddings"
	"k8stool/internal/embeddings/generator"
	"k8stool/internal/embeddings/processor"
	"k8stool/internal/embeddings/store"

	"github.com/spf13/cobra"
)

// NewEmbeddingsCmd creates a new embeddings command
func NewEmbeddingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "embeddings",
		Short: "Manage documentation embeddings",
		Long: `Generate and manage embeddings for k8stool documentation.
These embeddings are used to provide better context for the AI agent.`,
	}

	cmd.AddCommand(newEmbeddingsGenerateCmd())
	return cmd
}

func newEmbeddingsGenerateCmd() *cobra.Command {
	var (
		apiKey  string
		docsDir string
		outFile string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate embeddings from documentation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if apiKey == "" {
				// Try to get from environment
				apiKey = os.Getenv("OPENAI_API_KEY")
				if apiKey == "" {
					return fmt.Errorf("OpenAI API key is required. Set --api-key flag or OPENAI_API_KEY environment variable")
				}
			}

			// Check if docs directory exists
			if docsDir == "" {
				// Try current directory's docs folder first
				if _, err := os.Stat("docs"); err == nil {
					docsDir = "docs"
				} else {
					// Try executable's directory
					exePath, err := os.Executable()
					if err != nil {
						return fmt.Errorf("failed to get executable path: %w", err)
					}
					exeDir := filepath.Dir(exePath)
					possiblePath := filepath.Join(exeDir, "docs")
					if _, err := os.Stat(possiblePath); err == nil {
						docsDir = possiblePath
					} else {
						return fmt.Errorf("docs directory not found. Use --docs-dir to specify the path")
					}
				}
			}

			// Verify the directory exists and is readable
			if info, err := os.Stat(docsDir); err != nil || !info.IsDir() {
				return fmt.Errorf("invalid docs directory: %s", docsDir)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create components
			store := store.NewFileStore(apiKey)
			proc := processor.NewMarkdownProcessor(3) // Minimum 3 lines per chunk
			gen := generator.NewOpenAIGenerator(apiKey)

			// Process all markdown files in docs directory
			err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip directories and non-markdown files
				if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
					return nil
				}

				// Read the file
				content, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read %s: %w", path, err)
				}

				// Create metadata
				metadata := embeddings.Metadata{
					Source:    strings.TrimPrefix(path, docsDir+"/"),
					Command:   strings.TrimSuffix(info.Name(), ".md"),
					Topic:     "", // Will be set by processor
					Type:      embeddings.SectionTypeOverview,
					IsCode:    false,
					IsTable:   false,
					TableCols: nil,
				}

				// Process the document into chunks
				chunks, err := proc.Process(string(content), metadata)
				if err != nil {
					return fmt.Errorf("failed to process %s: %w", path, err)
				}

				// Generate embeddings for each chunk
				for _, chunk := range chunks {
					embedding, err := gen.Generate(chunk.Content)
					if err != nil {
						return fmt.Errorf("failed to generate embedding for chunk in %s: %w", path, err)
					}
					chunk.Embedding = embedding

					// Store the chunk
					if err := store.Store(chunk); err != nil {
						return fmt.Errorf("failed to store chunk from %s: %w", path, err)
					}
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to generate embeddings: %w", err)
			}

			// Save the store
			if err := store.Save(outFile); err != nil {
				return fmt.Errorf("failed to save embeddings: %w", err)
			}

			fmt.Printf("Successfully generated embeddings and saved to %s\n", outFile)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&apiKey, "api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
	cmd.Flags().StringVar(&docsDir, "docs-dir", "", "Path to documentation directory (default: ./docs or <executable_dir>/docs)")
	cmd.Flags().StringVar(&outFile, "out", "embeddings.json", "Output file for embeddings")

	return cmd
}

// getTopicFromPath extracts the topic from a file path
func getTopicFromPath(path string) string {
	// For command docs (e.g., docs/commands/pods.md), use the command name
	if strings.Contains(path, "commands/") {
		base := filepath.Base(path)
		return strings.TrimSuffix(base, ".md")
	}

	// For other docs, use the directory name
	dir := filepath.Dir(path)
	return filepath.Base(dir)
}
