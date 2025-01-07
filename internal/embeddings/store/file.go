package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"k8stool/internal/embeddings"
	"k8stool/internal/embeddings/generator"
)

// FileStore implements embeddings.EmbeddingStore using a simple file-based approach
type FileStore struct {
	chunks    []*embeddings.Chunk
	generator *generator.OpenAIGenerator
}

// NewFileStore creates a new file-based embedding store
func NewFileStore(apiKey string) *FileStore {
	return &FileStore{
		chunks:    make([]*embeddings.Chunk, 0),
		generator: generator.NewOpenAIGenerator(apiKey),
	}
}

// Store saves a chunk and its embedding
func (s *FileStore) Store(chunk *embeddings.Chunk) error {
	s.chunks = append(s.chunks, chunk)
	return nil
}

// Search finds the most relevant chunks for a query using cosine similarity
func (s *FileStore) Search(query string, limit int) ([]*embeddings.Chunk, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.generator.Generate(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Extract command from query if it exists
	queryLower := strings.ToLower(query)
	var targetCommand string
	if strings.Contains(queryLower, "logs command") {
		targetCommand = "logs"
	} else if strings.Contains(queryLower, "pod command") {
		targetCommand = "pods"
	} else if strings.Contains(queryLower, "deployment command") {
		targetCommand = "deployments"
	}

	// Detect if query is about specific section types
	var targetTypes []embeddings.SectionType
	if strings.Contains(queryLower, "usage") || strings.Contains(queryLower, "how to use") {
		targetTypes = append(targetTypes, embeddings.SectionTypeUsage)
		targetTypes = append(targetTypes, embeddings.SectionTypeExample)
	} else if strings.Contains(queryLower, "example") {
		targetTypes = append(targetTypes, embeddings.SectionTypeExample)
	} else if strings.Contains(queryLower, "flag") || strings.Contains(queryLower, "option") {
		targetTypes = append(targetTypes, embeddings.SectionTypeFlags)
	}

	// If no specific types requested but asking about usage, default to a standard order
	if len(targetTypes) == 0 && strings.Contains(queryLower, "how") {
		targetTypes = []embeddings.SectionType{
			embeddings.SectionTypeUsage,
			embeddings.SectionTypeExample,
			embeddings.SectionTypeFlags,
		}
	}

	// Calculate similarities and apply metadata boosts
	results := make([]searchResult, 0, len(s.chunks))
	for _, chunk := range s.chunks {
		similarity := cosineSimilarity(queryEmbedding, chunk.Embedding)

		// Apply metadata-based boosts
		score := similarity

		// Boost command-specific content significantly
		if targetCommand != "" && chunk.Metadata.Command == targetCommand {
			score *= 2.0 // 100% boost for matching command
		}

		// Boost sections of target types
		if len(targetTypes) > 0 {
			for i, targetType := range targetTypes {
				if chunk.Metadata.Type == targetType {
					// Earlier types in the list get higher boosts
					boost := 1.5 - (float32(i) * 0.1)
					score *= boost
					break
				}
			}
		}

		// Add to results if score is significant
		if score > 0.1 { // Threshold to filter out low-relevance chunks
			results = append(results, searchResult{
				chunk:      chunk,
				similarity: score,
			})
		}
	}

	// Sort by similarity (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})

	// Take top N results
	if limit > len(results) {
		limit = len(results)
	}

	// Ensure we get sections in a logical order
	var finalChunks []*embeddings.Chunk
	seenTypes := make(map[embeddings.SectionType]bool)

	// First, add chunks from the target command in the desired order
	for _, targetType := range targetTypes {
		for _, result := range results {
			chunk := result.chunk
			if chunk.Metadata.Command == targetCommand && chunk.Metadata.Type == targetType && !seenTypes[targetType] {
				finalChunks = append(finalChunks, chunk)
				seenTypes[targetType] = true
				break
			}
		}
	}

	// If we still need more chunks, add other relevant ones
	if len(finalChunks) < limit {
		for _, result := range results {
			chunk := result.chunk
			if !seenTypes[chunk.Metadata.Type] {
				finalChunks = append(finalChunks, chunk)
				seenTypes[chunk.Metadata.Type] = true
				if len(finalChunks) >= limit {
					break
				}
			}
		}
	}

	return finalChunks, nil
}

// Load initializes the store from a file
func (s *FileStore) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.chunks)
}

// Save persists the store to a file
func (s *FileStore) Save(path string) error {
	data, err := json.Marshal(s.chunks)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// searchResult represents a chunk with its similarity score
type searchResult struct {
	chunk      *embeddings.Chunk
	similarity float32
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt calculates the square root of a float32
func sqrt(x float32) float32 {
	return float32(float64(x))
}
