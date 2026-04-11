import type { Plugin } from "@opencode-ai/plugin"

// Token estimation: ~4 chars per token
const CHARS_PER_TOKEN = 4
const MAX_CONTEXT_TOKENS = 100_000
const MAX_CONTEXT_CHARS = MAX_CONTEXT_TOKENS * CHARS_PER_TOKEN
const MAX_ADVISOR_CALLS = 3

// Patterns to sanitize before sending to advisor
const SECRET_PATTERNS = [
  /(?:api[_-]?key|apikey|secret|token|password|passwd|credential|auth)[\s]*[=:]\s*["']?[A-Za-z0-9_\-/.+=]{16,}["']?/gi,
  /(?:sk|pk|rk|ak)-[A-Za-z0-9]{20,}/g,
  /ghp_[A-Za-z0-9]{36,}/g,
  /eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}/g,
]

function sanitizeSecrets(text: string): string {
  let result = text
  for (const pattern of SECRET_PATTERNS) {
    result = result.replace(pattern, "[REDACTED]")
  }
  return result
}

interface FlatMessage {
  role: string
  text: string
}

function truncateTranscript(messages: FlatMessage[]): string {
  const full = messages.map((m) => `[${m.role}]: ${m.text}`).join("\n\n")

  if (full.length <= MAX_CONTEXT_CHARS) {
    return sanitizeSecrets(full)
  }

  // Keep first 5 messages + last portion that fits within budget
  const firstMessages = messages.slice(0, 5)
  const firstPart = firstMessages
    .map((m) => `[${m.role}]: ${m.text}`)
    .join("\n\n")

  const remainingBudget = MAX_CONTEXT_CHARS - firstPart.length - 200
  if (remainingBudget <= 0) {
    return sanitizeSecrets(firstPart)
  }

  const lastPart = full.slice(-remainingBudget)
  const truncatedTokens = Math.round(remainingBudget / CHARS_PER_TOKEN)
  const result = `${firstPart}\n\n[... transcript truncated — showing last ~${truncatedTokens} tokens ...]\n\n${lastPart}`

  return sanitizeSecrets(result)
}

const advisorPlugin: Plugin = async (ctx) => {
  const client = ctx.client
  const callCounts = new Map<string, number>()

  return {
    name: "advisor",
    tools: {
      advisor: {
        description:
          "Consult a senior strategic advisor for guidance on complex decisions. " +
          "Use BEFORE starting substantial work, when stuck after 2+ attempts, " +
          "before declaring done on complex tasks, or before changing approach. " +
          "The advisor sees your full conversation history and provides strategic direction.",
        parameters: {
          type: "object" as const,
          properties: {
            question: {
              type: "string",
              description:
                "What you need guidance on. Be specific: include what you have tried, " +
                "what failed, and what you are considering.",
            },
          },
          required: ["question"],
        },
        async execute(
          args: { question: string },
          toolCtx: { sessionID: string },
        ): Promise<string> {
          const sessionID = toolCtx.sessionID

          // Enforce max calls per session
          const currentCount = callCounts.get(sessionID) ?? 0
          if (currentCount >= MAX_ADVISOR_CALLS) {
            return (
              `Advisor limit reached (${MAX_ADVISOR_CALLS} calls per session). ` +
              "Continue with your best judgment. If still stuck, ask the user for guidance."
            )
          }
          callCounts.set(sessionID, currentCount + 1)

          try {
            // 1. Read full transcript of current session
            const messagesResponse = await client.session.messages({
              path: { id: sessionID },
            })

            // 2. Flatten messages into readable transcript
            const flatMessages: FlatMessage[] = []
            if (Array.isArray(messagesResponse)) {
              for (const msg of messagesResponse) {
                const role =
                  (msg as Record<string, unknown>).info &&
                  typeof (msg as Record<string, unknown>).info === "object"
                    ? ((msg as Record<string, unknown>).info as Record<string, unknown>)
                        .role ?? "unknown"
                    : "unknown"
                const textParts: string[] = []
                const parts = (msg as Record<string, unknown>).parts
                if (Array.isArray(parts)) {
                  for (const part of parts) {
                    const p = part as Record<string, unknown>
                    if (p.type === "text" && typeof p.text === "string") {
                      textParts.push(p.text)
                    } else if (p.type === "tool-invocation") {
                      const toolName =
                        typeof p.toolName === "string"
                          ? p.toolName
                          : "unknown-tool"
                      const input = p.input
                        ? JSON.stringify(p.input).slice(0, 500)
                        : ""
                      const output =
                        p.state === "completed" && p.output
                          ? JSON.stringify(p.output).slice(0, 500)
                          : ""
                      textParts.push(
                        `[Tool: ${toolName}] Input: ${input}${output ? ` Output: ${output}` : ""}`,
                      )
                    }
                  }
                }
                if (textParts.length > 0) {
                  flatMessages.push({
                    role: String(role),
                    text: textParts.join("\n"),
                  })
                }
              }
            }

            // 3. Build truncated transcript
            const transcript = truncateTranscript(flatMessages)

            // 4. Create temporary advisor session
            const advisorSession = await client.session.create({
              body: { title: `advisor-${Date.now()}` },
            })

            const advisorSessionId =
              advisorSession &&
              typeof advisorSession === "object" &&
              "id" in advisorSession
                ? (advisorSession as Record<string, unknown>).id
                : undefined

            if (!advisorSessionId || typeof advisorSessionId !== "string") {
              return "Advisor unavailable: could not create advisor session. Continue with your best judgment."
            }

            // 5. Send transcript + question to advisor
            await client.session.prompt({
              path: { id: advisorSessionId },
              body: {
                parts: [
                  {
                    type: "text" as const,
                    text: [
                      "## Full Session Transcript",
                      "",
                      transcript,
                      "",
                      "## Question",
                      "",
                      args.question,
                      "",
                      "Respond with a clear, actionable plan in under 100 words using enumerated steps.",
                    ].join("\n"),
                  },
                ],
              },
            })

            // 6. Read the advisor's response
            const advisorMessages = await client.session.messages({
              path: { id: advisorSessionId },
            })

            let advisorResponse = ""
            if (Array.isArray(advisorMessages)) {
              for (const msg of advisorMessages) {
                const m = msg as Record<string, unknown>
                const info = m.info as Record<string, unknown> | undefined
                if (info?.role === "assistant" && Array.isArray(m.parts)) {
                  for (const part of m.parts as Record<string, unknown>[]) {
                    if (
                      part.type === "text" &&
                      typeof part.text === "string"
                    ) {
                      advisorResponse += part.text
                    }
                  }
                }
              }
            }

            if (!advisorResponse) {
              return "Advisor returned empty response. Continue with your best judgment."
            }

            const callNum = callCounts.get(sessionID) ?? 1
            return `[Advisor guidance — call ${callNum}/${MAX_ADVISOR_CALLS}]\n\n${advisorResponse}`
          } catch (error) {
            const message =
              error instanceof Error ? error.message : String(error)
            return `Advisor unavailable: ${message}. Continue with your best judgment.`
          }
        },
      },
    },
  }
}

export default advisorPlugin
