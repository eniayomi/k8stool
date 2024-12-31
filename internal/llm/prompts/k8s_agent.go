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
const TaskPromptTemplate = `Task Type: %s
Description: %s
Current Context: %s
Current Namespace: %s
Additional Parameters: %v

Please analyze the situation and provide:
1. Your understanding of the task
2. Proposed solution or response
3. Any potential risks or considerations
4. Step-by-step actions (if applicable)`

// ErrorAnalysisTemplate helps analyze Kubernetes-related errors
const ErrorAnalysisTemplate = `Error: %s
Resource Type: %s
Resource Name: %s
Namespace: %s

Please analyze this error and provide:
1. Root cause analysis
2. Potential solutions
3. Prevention measures for the future`

// ResourceAnalysisTemplate helps analyze Kubernetes resources
const ResourceAnalysisTemplate = `Resource Type: %s
Resource Name: %s
Namespace: %s
Current State: %s

Please analyze this resource and provide:
1. Configuration assessment
2. Potential issues or improvements
3. Best practice recommendations
4. Health status evaluation`
