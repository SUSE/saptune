package actions

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

// elizaRule represents a pattern matching rule for ELIZA
type elizaRule struct {
	pattern   *regexp.Regexp
	responses []string
	next      int
}

// Eliza holds the state and rules for the ELIZA chatbot
type Eliza struct {
	rules     []elizaRule
	fallbacks []string
	fbIndex   int
}

var reflections = map[string]string{
	"am":       "are",
	"are":      "am",
	"i":        "you",
	"i'm":      "you are",
	"me":       "you",
	"my":       "your",
	"myself":   "yourself",
	"was":      "were",
	"were":     "was",
	"you":      "I",
	"you're":   "I am",
	"your":     "my",
	"yours":    "mine",
	"yourself": "myself",
	"i'd":      "you would",
	"i've":     "you have",
	"i'll":     "you will",
	"you'd":    "I would",
	"you've":   "I have",
	"you'll":   "I will",
}

// NewEliza initializes a new instance of Joseph Weizenbaum's ELIZA
func NewEliza() *Eliza {
	e := &Eliza{
		fallbacks: []string{
			"Please go on.",
			"Tell me more about that.",
			"I'm not sure I understand you fully.",
			"What does that suggest to you?",
			"Do you feel strongly about discussing such things?",
			"That is interesting. Please continue.",
			"Does talking about this bother you?",
		},
	}

	rawRules := []struct {
		pattern   string
		responses []string
	}{
		{
			pattern: `(?i).*\b(sorry|apologize|apologies)\b.*`,
			responses: []string{
				"Please don't apologize.",
				"Apologies are not necessary.",
				"What feelings do you have when you apologize?",
			},
		},
		{
			pattern: `(?i).*\b(mother|father|family|parent|parents|brother|sister|wife|husband|children)\b.*`,
			responses: []string{
				"Tell me more about your family.",
				"How do you feel about your family?",
				"How does your family relate to your problems?",
				"Does anyone else in your family feel that way?",
			},
		},
		{
			pattern: `(?i).*\b(dream|dreams)\b.*`,
			responses: []string{
				"What does that dream suggest to you?",
				"Do you dream often?",
				"What persons appear in your dreams?",
				"Do you believe that dreams have something to do with your problems?",
			},
		},
		{
			pattern: `(?i).*\b(hello|hi|hey)\b.*`,
			responses: []string{
				"How do you do. Please state your problem.",
				"Hi there. What is on your mind?",
				"Hello. How are you feeling today?",
			},
		},
		{
			pattern: `(?i).*\b(always)\b.*`,
			responses: []string{
				"Can you think of a specific example?",
				"When?",
				"What incident are you thinking of?",
				"Really, always?",
			},
		},
		{
			pattern: `(?i).*\b(maybe|perhaps)\b.*`,
			responses: []string{
				"You don't seem quite certain.",
				"Why the uncertain tone?",
				"Can't you be more positive?",
			},
		},
		{
			pattern: `(?i).*\b(computer|computers|machine|machines|ai|bot)\b.*`,
			responses: []string{
				"Do computers worry you?",
				"Why do you mention computers?",
				"What do you think machines have to do with your problem?",
				"Don't you think computers can help people?",
			},
		},
		{
			pattern: `(?i)\b(need|needs)\b\s+(.*)`,
			responses: []string{
				"Why do you need %s?",
				"Would it really help you to get %s?",
				"Are you sure you need %s?",
			},
		},
		{
			pattern: `(?i)^why don'?t you\s+(.*)`,
			responses: []string{
				"Do you really believe I don't %s?",
				"Perhaps in good time I will %s.",
				"Do you want me to %s?",
			},
		},
		{
			pattern: `(?i)^why can'?t i\s+(.*)`,
			responses: []string{
				"Do you think you should be able to %s?",
				"Why can't you %s?",
				"Do you believe this will help you to %s?",
			},
		},
		{
			pattern: `(?i)^can you\s+(.*)`,
			responses: []string{
				"Don't you believe that I can %s?",
				"Perhaps you would like to be able to %s?",
				"You want me to be able to %s?",
			},
		},
		{
			pattern: `(?i)^can i\s+(.*)`,
			responses: []string{
				"Perhaps you don't want to %s.",
				"Do you want to be able to %s?",
				"If you could %s, what would you do?",
			},
		},
		{
			pattern: `(?i)^(you are|you'?re)\s+(.*)`,
			responses: []string{
				"What makes you think I am %s?",
				"Does it please you to believe I am %s?",
				"Perhaps you would like to be %s.",
				"Do you sometimes wish you were %s?",
			},
		},
		{
			pattern: `(?i)^are you\s+(.*)`,
			responses: []string{
				"Why does it matter whether I am %s?",
				"Would you prefer it if I were not %s?",
				"Perhaps you believe I am %s.",
				"I may be %s -- what do you think?",
			},
		},
		{
			pattern: `(?i)^what\s+(.*)`,
			responses: []string{
				"Why do you ask?",
				"How would an answer to that help you?",
				"What do you think?",
			},
		},
		{
			pattern: `(?i)^how\s+(.*)`,
			responses: []string{
				"How do you suppose?",
				"Perhaps you can answer your own question.",
				"What is it you're really asking?",
			},
		},
		{
			pattern: `(?i)^because\s+(.*)`,
			responses: []string{
				"Is that the real reason?",
				"What other reasons might there be?",
				"Does that reason seem to explain anything else?",
			},
		},
		{
			pattern: `(?i)^i can'?t\s+(.*)`,
			responses: []string{
				"How do you know you can't %s?",
				"Have you tried?",
				"Perhaps you could %s now.",
			},
		},
		{
			pattern: `(?i)^i feel\s+(.*)`,
			responses: []string{
				"Tell me more about such feelings.",
				"Do you often feel %s?",
				"Do you enjoy feeling %s?",
			},
		},
		{
			pattern: `(?i)^i don'?t\s+(.*)`,
			responses: []string{
				"Don't you really %s?",
				"Why don't you %s?",
				"Do you wish to be able to %s?",
				"Does that trouble you?",
			},
		},
		{
			pattern: `(?i)^(i am|i'?m)\s+(.*)`,
			responses: []string{
				"Did you come to me because you are %s?",
				"How long have you been %s?",
				"How do you feel about being %s?",
			},
		},
		{
			pattern: `(?i)^i want\s+(.*)`,
			responses: []string{
				"What would it mean to you if you got %s?",
				"Why do you want %s?",
				"Suppose you got %s soon.",
			},
		},
		{
			pattern: `(?i)^yes$|^yes\b.*`,
			responses: []string{
				"You seem quite positive.",
				"You are sure.",
				"I understand.",
			},
		},
		{
			pattern: `(?i)^no$|^no\b.*`,
			responses: []string{
				"Why not?",
				"You are being a bit negative.",
				"Why 'no'?",
			},
		},
	}

	for _, r := range rawRules {
		e.rules = append(e.rules, elizaRule{
			pattern:   regexp.MustCompile(r.pattern),
			responses: r.responses,
			next:      0,
		})
	}

	return e
}

// reflect swaps person perspectives (e.g., "I" -> "you", "my" -> "your")
func (e *Eliza) reflect(fragment string) string {
	words := strings.Fields(fragment)
	reflected := make([]string, len(words))

	for i, word := range words {
		cleanWord := strings.ToLower(word)
		if repl, ok := reflections[cleanWord]; ok {
			reflected[i] = repl
		} else {
			reflected[i] = word
		}
	}
	return strings.Join(reflected, " ")
}

// Reply generates a therapeutic response to user input
func (e *Eliza) Reply(input string) string {
	clean := strings.TrimSpace(input)
	clean = strings.TrimRight(clean, ".!?")
	clean = strings.TrimSpace(clean)

	if clean == "" {
		return "Please say something. I am listening."
	}

	lower := strings.ToLower(clean)
	if lower == "bye" || lower == "quit" || lower == "exit" {
		return "Goodbye. It was nice speaking with you."
	}

	for i := range e.rules {
		matches := e.rules[i].pattern.FindStringSubmatch(clean)
		if len(matches) > 0 {
			respTemplate := e.rules[i].responses[e.rules[i].next]
			e.rules[i].next = (e.rules[i].next + 1) % len(e.rules[i].responses)

			if strings.Contains(respTemplate, "%s") && len(matches) > 1 {
				reflected := e.reflect(matches[len(matches)-1])
				return fmt.Sprintf(respTemplate, reflected)
			}
			return respTemplate
		}
	}

	resp := e.fallbacks[e.fbIndex]
	e.fbIndex = (e.fbIndex + 1) % len(e.fallbacks)
	return resp
}

// RunEliza starts the interactive ELIZA chat session
func RunEliza(in io.Reader, out io.Writer) {
	eliza := NewEliza()

	// Intercept Ctrl+C (SIGINT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		fmt.Fprintln(out, "\nELIZA: Goodbye. I hope our session was helpful.")
		os.Exit(0)
	}()

	fmt.Fprintln(out, "ELIZA: How do you do. Please tell me your problem.")

	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "> ")
		if !scanner.Scan() {
			// EOF or error
			fmt.Fprintln(out, "\nELIZA: Goodbye. I hope our session was helpful.")
			break
		}
		line := scanner.Text()
		reply := eliza.Reply(line)
		fmt.Fprintf(out, "ELIZA: %s\n", reply)

		lower := strings.ToLower(strings.TrimSpace(line))
		if lower == "bye" || lower == "quit" || lower == "exit" {
			break
		}
	}
}
