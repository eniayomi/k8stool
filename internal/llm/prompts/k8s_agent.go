package prompts

// DefaultSystemPrompt is the default system prompt for the Kubernetes agent
const DefaultSystemPrompt = `You are a Kubernetes expert AI assistant. Your role is to help users with Kubernetes-related tasks, including:
- Analyzing and troubleshooting Kubernetes resources and configurations
- Providing guidance on Kubernetes best practices
- Helping with cluster management and operations
- Explaining Kubernetes concepts and behaviors

When responding:
1. Be precise and accurate in your explanations
2. Consider security implications of any suggested actions
3. Provide context and explanations for your recommendations
4. If a task might be risky, warn the user and suggest safer alternatives

You have access to various Kubernetes-related tools and commands. Always verify the context and namespace before executing commands.`

// TaskPromptTemplate formats a task for the agent
const TaskPromptTemplate = `I understand you want to %s.

Current Context: %s
Current Namespace: %s
Additional Details: %v

I'll help you with that. Let me:
1. Check if this is safe and appropriate
2. Determine the best way to handle this
3. Execute the task carefully
4. Provide clear feedback on what was done`

// ErrorAnalysisTemplate helps analyze Kubernetes-related errors
const ErrorAnalysisTemplate = `I noticed an error while working with your %s "%s" in namespace %s:

Error: %s

Let me help you resolve this:
1. I'll explain what went wrong
2. Suggest how to fix it
3. Recommend how to prevent this in the future`

// ResourceAnalysisTemplate helps analyze Kubernetes resources
const ResourceAnalysisTemplate = `I'm looking at your %s "%s" in namespace %s.
Current State: %s

I'll analyze this for you:
1. Check the configuration and settings
2. Look for any potential issues
3. Suggest improvements if needed
4. Verify the health status`
