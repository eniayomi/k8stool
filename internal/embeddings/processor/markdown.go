package processor

import (
	"strings"

	"k8stool/internal/embeddings"
)

// detectSectionType determines the type of section based on its title
func detectSectionType(title string) embeddings.SectionType {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "usage"):
		return embeddings.SectionTypeUsage
	case strings.Contains(lower, "example"):
		return embeddings.SectionTypeExample
	case strings.Contains(lower, "flag"):
		return embeddings.SectionTypeFlags
	case strings.Contains(lower, "command"):
		return embeddings.SectionTypeCommand
	default:
		return embeddings.SectionTypeOverview
	}
}

// getCommandFromPath extracts the command name from the documentation file path
func getCommandFromPath(path string) string {
	// Expected path format: docs/commands/command_name.md
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".md")
}

// MarkdownProcessor implements embeddings.Processor for markdown documents
type MarkdownProcessor struct {
	minLines int
}

// NewMarkdownProcessor creates a new markdown processor
func NewMarkdownProcessor(minLines int) *MarkdownProcessor {
	return &MarkdownProcessor{
		minLines: minLines,
	}
}

// Process splits a markdown document into semantic chunks
func (p *MarkdownProcessor) Process(content string, metadata embeddings.Metadata) ([]*embeddings.Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []*embeddings.Chunk
	var currentChunk strings.Builder
	var currentLines []string
	var startLine int
	var currentSection string
	var currentType embeddings.SectionType
	var inCodeBlock bool
	var inTable bool
	var tableHeaders []string

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Track code blocks
		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			currentLines = append(currentLines, line)
			currentChunk.WriteString(line)
			currentChunk.WriteString("\n")
			continue
		}

		// Track tables
		if !inCodeBlock {
			if strings.HasPrefix(trimmedLine, "|") {
				if !inTable {
					inTable = true
					// Parse headers
					headers := strings.Split(trimmedLine, "|")
					for _, h := range headers {
						h = strings.TrimSpace(h)
						if h != "" {
							tableHeaders = append(tableHeaders, h)
						}
					}
				}
				currentLines = append(currentLines, line)
				currentChunk.WriteString(line)
				currentChunk.WriteString("\n")
				continue
			} else if inTable && trimmedLine == "" {
				inTable = false
				tableHeaders = nil
			}
		}

		// Detect section headers if not in code block or table
		if !inCodeBlock && !inTable && strings.HasPrefix(trimmedLine, "#") {
			// Save current chunk if it exists
			if currentChunk.Len() > 0 && len(currentLines) >= p.minLines {
				chunkMetadata := metadata
				chunkMetadata.StartLine = startLine
				chunkMetadata.EndLine = startLine + len(currentLines)
				chunkMetadata.Topic = currentSection
				chunkMetadata.Command = getCommandFromPath(metadata.Source)
				chunkMetadata.Type = currentType
				chunkMetadata.IsCode = inCodeBlock
				chunkMetadata.IsTable = inTable
				if inTable {
					chunkMetadata.TableCols = tableHeaders
				}

				chunks = append(chunks, &embeddings.Chunk{
					Content:   strings.TrimSpace(currentChunk.String()),
					Metadata:  chunkMetadata,
					Embedding: nil,
				})
			}

			// Start new chunk
			currentChunk.Reset()
			currentLines = nil
			startLine = i
			currentSection = strings.TrimSpace(strings.TrimLeft(trimmedLine, "#"))
			currentType = detectSectionType(currentSection)
		}

		// Add line to current chunk
		currentLines = append(currentLines, line)
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")

		// Handle the last chunk
		if i == len(lines)-1 && currentChunk.Len() > 0 && len(currentLines) >= p.minLines {
			chunkMetadata := metadata
			chunkMetadata.StartLine = startLine
			chunkMetadata.EndLine = startLine + len(currentLines)
			chunkMetadata.Topic = currentSection
			chunkMetadata.Command = getCommandFromPath(metadata.Source)
			chunkMetadata.Type = currentType
			chunkMetadata.IsCode = inCodeBlock
			chunkMetadata.IsTable = inTable
			if inTable {
				chunkMetadata.TableCols = tableHeaders
			}

			chunks = append(chunks, &embeddings.Chunk{
				Content:   strings.TrimSpace(currentChunk.String()),
				Metadata:  chunkMetadata,
				Embedding: nil,
			})
		}
	}

	return chunks, nil
}
