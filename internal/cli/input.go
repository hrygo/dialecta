package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// InputReader handles reading material input from various sources
type InputReader struct {
	stdin io.Reader
	out   io.Writer
}

// NewInputReader creates a new InputReader
func NewInputReader(stdin io.Reader, out io.Writer) *InputReader {
	return &InputReader{
		stdin: stdin,
		out:   out,
	}
}

// DefaultInputReader creates an InputReader using os.Stdin and os.Stdout
func DefaultInputReader() *InputReader {
	return NewInputReader(os.Stdin, os.Stdout)
}

// ReadFile reads material from a file
func (r *InputReader) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return string(data), nil
}

// ReadFileOrText tries to read as file first, if file doesn't exist, treats input as text
// - If path is a valid file: reads file content
// - If path is a directory: returns error
// - If path doesn't exist: treats as text content
// - If access denied: returns error
func (r *InputReader) ReadFileOrText(source string) (string, error) {
	info, err := os.Stat(source)
	if err == nil && !info.IsDir() {
		// File exists and is not a directory
		return r.ReadFile(source)
	}
	if err == nil && info.IsDir() {
		// Path is a directory
		return "", fmt.Errorf("无法读取目录: %s", source)
	}
	if err != nil && !os.IsNotExist(err) {
		// Other errors (e.g., permission denied)
		return "", fmt.Errorf("无法访问: %w", err)
	}
	// File doesn't exist, treat source as text content
	return source, nil
}

// ReadStdin reads material from stdin (pipe mode)
func (r *InputReader) ReadStdin() (string, error) {
	data, err := io.ReadAll(r.stdin)
	if err != nil {
		return "", fmt.Errorf("读取标准输入失败: %w", err)
	}
	return string(data), nil
}

// ReadInteractive reads material interactively from the user
// The user can finish input by entering two consecutive empty lines
func (r *InputReader) ReadInteractive() (string, error) {
	return r.ReadInteractiveWithContext()
}

// ReadInteractiveWithContext reads both question and optional context file
func (r *InputReader) ReadInteractiveWithContext() (string, error) {
	fmt.Fprintln(r.out)
	fmt.Fprintf(r.out, "%s%s┌─ 📝 INTERACTIVE INPUT ───────────────────────────────────────┐%s\n", ColorBrightCyan, ColorBold, ColorReset)
	fmt.Fprintf(r.out, "%s│%s  You can provide a question and an optional context file.      %s│%s\n", ColorBrightCyan, ColorReset, ColorBrightCyan, ColorReset)
	fmt.Fprintf(r.out, "%s%s└──────────────────────────────────────────────────────────────┘%s\n", ColorBrightCyan, ColorBold, ColorReset)
	fmt.Fprintln(r.out)

	// Step 1: Ask for the question/instruction
	fmt.Fprintf(r.out, "%s%s① Enter your question or instruction:%s\n", ColorBrightYellow, ColorBold, ColorReset)
	fmt.Fprintf(r.out, "%s   (Press ENTER twice to finish)%s\n\n", ColorDim, ColorReset)
	fmt.Fprintf(r.out, "%s%s▸ %s", ColorBrightGreen, ColorBold, ColorReset)

	question, err := r.readMultiLineInput()
	if err != nil {
		return "", err
	}

	// Step 2: Ask for optional context file
	fmt.Fprintln(r.out)
	fmt.Fprintf(r.out, "%s%s② Enter context file path (optional, press ENTER to skip):%s\n", ColorBrightYellow, ColorBold, ColorReset)
	fmt.Fprintf(r.out, "%s%s▸ %s", ColorBrightGreen, ColorBold, ColorReset)

	scanner := bufio.NewScanner(r.stdin)
	var contextFile string
	if scanner.Scan() {
		contextFile = strings.TrimSpace(scanner.Text())
	}

	// Combine question and context
	return r.combineQuestionAndContext(question, contextFile)
}

// readMultiLineInput reads multiple lines until two consecutive empty lines
func (r *InputReader) readMultiLineInput() (string, error) {
	var lines []string
	scanner := bufio.NewScanner(r.stdin)
	emptyCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		lines = append(lines, line)
		// Print line number prompt for next line
		if emptyCount < 2 {
			fmt.Fprintf(r.out, "%s%s▸ %s", ColorBrightGreen, ColorBold, ColorReset)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}

	// Trim trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n"), nil
}

// combineQuestionAndContext combines question and optional context file into structured material
func (r *InputReader) combineQuestionAndContext(question, contextFile string) (string, error) {
	if contextFile == "" {
		// No context file, return question only
		fmt.Fprintf(r.out, "\n%s%s✓ Question received (%d characters)%s\n\n", ColorBrightGreen, ColorBold, len(question), ColorReset)
		return question, nil
	}

	// Try to read context file
	contextContent, err := r.ReadFileOrText(contextFile)
	if err != nil {
		return "", fmt.Errorf("读取上下文文件失败: %w", err)
	}

	// Determine if contextFile was treated as file or text
	info, statErr := os.Stat(contextFile)
	isFile := statErr == nil && !info.IsDir()

	// Structured combination
	var material strings.Builder
	material.WriteString("# 用户问题\n\n")
	material.WriteString(question)
	material.WriteString("\n\n")
	material.WriteString("---\n\n")

	if isFile {
		material.WriteString(fmt.Sprintf("# 上下文文件：%s\n\n", contextFile))
		fmt.Fprintf(r.out, "\n%s%s✓ Question + Context file loaded%s\n", ColorBrightGreen, ColorBold, ColorReset)
		fmt.Fprintf(r.out, "  %s• Question: %d characters%s\n", ColorDim, len(question), ColorReset)
		fmt.Fprintf(r.out, "  %s• Context file: %s (%d characters)%s\n\n", ColorDim, contextFile, len(contextContent), ColorReset)
	} else {
		material.WriteString("# 上下文内容\n\n")
		fmt.Fprintf(r.out, "\n%s%s✓ Question + Context text received%s\n", ColorBrightGreen, ColorBold, ColorReset)
		fmt.Fprintf(r.out, "  %s• Question: %d characters%s\n", ColorDim, len(question), ColorReset)
		fmt.Fprintf(r.out, "  %s• Context: %d characters%s\n\n", ColorDim, len(contextContent), ColorReset)
	}

	material.WriteString(contextContent)

	return material.String(), nil
}

// ReadMaterial reads material based on the input mode
// - If interactive is true, reads interactively
// - If source is "-", reads from stdin
// - Otherwise, tries to read as file path first, then treats as text content
func (r *InputReader) ReadMaterial(source string, interactive bool) (string, error) {
	if interactive {
		return r.ReadInteractive()
	}
	if source == "-" {
		return r.ReadStdin()
	}
	// Smart detection: try file first, then treat as text
	return r.ReadFileOrText(source)
}

// ValidateMaterial checks if the material is valid (non-empty)
func ValidateMaterial(material string) error {
	if strings.TrimSpace(material) == "" {
		return fmt.Errorf("材料内容为空")
	}
	return nil
}

// ModelCombination represents a model combination choice
type ModelCombination struct {
	ID            string
	Name          string
	JudgeProvider string
	ProProvider   string
	ConProvider   string
}

// GetModelCombinations returns available model combinations
// Strategy: Judge matches the stronger side (Pro or dominant model)
func GetModelCombinations() []ModelCombination {
	return []ModelCombination{
		{ID: "1", Name: "All Gemini", JudgeProvider: "gemini", ProProvider: "gemini", ConProvider: "gemini"},
		{ID: "2", Name: "Gemini Judge, DeepSeek Debate", JudgeProvider: "gemini", ProProvider: "deepseek", ConProvider: "deepseek"},
		{ID: "3", Name: "Gemini Judge, DeepSeek vs Qwen", JudgeProvider: "gemini", ProProvider: "deepseek", ConProvider: "dashscope"},
		{ID: "4", Name: "Gemini Judge, Qwen Debate", JudgeProvider: "gemini", ProProvider: "dashscope", ConProvider: "dashscope"},
		{ID: "5", Name: "All DeepSeek", JudgeProvider: "deepseek", ProProvider: "deepseek", ConProvider: "deepseek"},
		{ID: "6", Name: "DeepSeek Judge, DeepSeek vs Qwen", JudgeProvider: "deepseek", ProProvider: "deepseek", ConProvider: "dashscope"},
		{ID: "7", Name: "DeepSeek Judge, Qwen Debate", JudgeProvider: "deepseek", ProProvider: "dashscope", ConProvider: "dashscope"},
		{ID: "8", Name: "All Qwen", JudgeProvider: "dashscope", ProProvider: "dashscope", ConProvider: "dashscope"},
	}
}

// SelectModelCombination prompts user to select a model combination
func (r *InputReader) SelectModelCombination() (*ModelCombination, error) {
	combinations := GetModelCombinations()

	fmt.Fprintln(r.out)
	fmt.Fprintf(r.out, "%s%s┌─ 🔀 MODEL COMBINATION SELECTION ─────────────────────────────┐%s\n", ColorBrightCyan, ColorBold, ColorReset)
	fmt.Fprintf(r.out, "%s│%s  Select a model combination for the debate:                 %s│%s\n", ColorBrightCyan, ColorReset, ColorBrightCyan, ColorReset)
	fmt.Fprintf(r.out, "%s%s└──────────────────────────────────────────────────────────────┘%s\n", ColorBrightCyan, ColorBold, ColorReset)
	fmt.Fprintln(r.out)

	// Display combinations with Judge, Pro, Con info
	for i, comb := range combinations {
		// Add group headers
		if i == 0 {
			fmt.Fprintf(r.out, "  %s🌟 Gemini Judge:%s\n", ColorBrightMagenta, ColorReset)
		} else if i == 4 {
			fmt.Fprintf(r.out, "\n  %s⚡ DeepSeek Judge:%s\n", ColorBrightBlue, ColorReset)
		} else if i == 7 {
			fmt.Fprintf(r.out, "\n  %s🔷 Qwen Judge:%s\n", ColorBrightCyan, ColorReset)
		}

		fmt.Fprintf(r.out, "    %s%s[%s]%s %s%-40s%s\n",
			ColorBrightYellow, ColorBold, comb.ID, ColorReset,
			ColorBrightWhite, comb.Name, ColorReset)
	}

	fmt.Fprintln(r.out)
	fmt.Fprintf(r.out, "%s%s▸ Enter your choice (1-8): %s", ColorBrightGreen, ColorBold, ColorReset)

	scanner := bufio.NewScanner(r.stdin)
	if !scanner.Scan() {
		return nil, fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())
	for _, comb := range combinations {
		if comb.ID == choice {
			fmt.Fprintf(r.out, "\n%s%s✓ Selected: %s%s\n", ColorBrightGreen, ColorBold, comb.Name, ColorReset)
			return &comb, nil
		}
	}

	return nil, fmt.Errorf("无效的选择: %s", choice)
}
