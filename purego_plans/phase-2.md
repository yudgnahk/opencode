# Phase 2: File Operations & Tools

**Duration**: Weeks 5-10 (6 weeks)  
**Team**: 2 engineers  
**Complexity**: 🟡 Medium  
**Dependencies**: Phase 1 complete

---

## Objectives

1. Implement file read/write/edit tools
2. Add glob pattern matching
3. Implement grep (ripgrep integration)
4. Bash command execution
5. Project management and detection
6. Path operations and ignore patterns

**Success Criterion**: Can browse files, search code, and execute shell commands via TUI

---

## Detailed Tasks

### Week 5-6: File Operations

#### Task 5.1: File Read Tool

Create `internal/tool/read.go`:

```go
package tool

import (
    "bufio"
    "fmt"
    "os"
)

type ReadRequest struct {
    FilePath string `json:"filePath" validate:"required"`
    Offset   int    `json:"offset"`
    Limit    int    `json:"limit"`
}

type ReadResponse struct {
    Content string   `json:"content"`
    Lines   []string `json:"lines"`
    Total   int      `json:"total"`
}

func (t *ToolExecutor) Read(req ReadRequest) (*ReadResponse, error) {
    file, err := os.Open(req.FilePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    // Default limits
    if req.Limit == 0 {
        req.Limit = 2000
    }

    scanner := bufio.NewScanner(file)
    var lines []string
    lineNum := 0

    for scanner.Scan() {
        if lineNum >= req.Offset && lineNum < req.Offset+req.Limit {
            // Format with line numbers (like cat -n)
            lines = append(lines, fmt.Sprintf("%5d\t%s", lineNum+1, scanner.Text()))
        }
        lineNum++
        if lineNum >= req.Offset+req.Limit {
            break
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return &ReadResponse{
        Content: string(bytes.Join(lines, []byte("\n"))),
        Lines:   lines,
        Total:   lineNum,
    }, nil
}
```

#### Task 5.2: File Write Tool

Create `internal/tool/write.go`:

```go
package tool

import (
    "fmt"
    "os"
    "path/filepath"
)

type WriteRequest struct {
    FilePath string `json:"filePath" validate:"required"`
    Content  string `json:"content" validate:"required"`
}

type WriteResponse struct {
    Success bool   `json:"success"`
    Path    string `json:"path"`
}

func (t *ToolExecutor) Write(req WriteRequest) (*WriteResponse, error) {
    // Ensure directory exists
    dir := filepath.Dir(req.FilePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create directory: %w", err)
    }

    // Write file
    if err := os.WriteFile(req.FilePath, []byte(req.Content), 0644); err != nil {
        return nil, fmt.Errorf("failed to write file: %w", err)
    }

    return &WriteResponse{
        Success: true,
        Path:    req.FilePath,
    }, nil
}
```

#### Task 5.3: File Edit Tool

Create `internal/tool/edit.go`:

```go
package tool

import (
    "fmt"
    "os"
    "strings"
)

type EditRequest struct {
    FilePath  string `json:"filePath" validate:"required"`
    OldString string `json:"oldString" validate:"required"`
    NewString string `json:"newString" validate:"required"`
    ReplaceAll bool  `json:"replaceAll"`
}

type EditResponse struct {
    Success   bool `json:"success"`
    Replaced  int  `json:"replaced"`
}

func (t *ToolExecutor) Edit(req EditRequest) (*EditResponse, error) {
    // Read file
    content, err := os.ReadFile(req.FilePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    contentStr := string(content)

    // Count occurrences
    count := strings.Count(contentStr, req.OldString)

    if count == 0 {
        return nil, fmt.Errorf("oldString not found in content")
    }

    if count > 1 && !req.ReplaceAll {
        return nil, fmt.Errorf("oldString found multiple times and requires more code context to uniquely identify the intended match")
    }

    // Perform replacement
    var newContent string
    if req.ReplaceAll {
        newContent = strings.ReplaceAll(contentStr, req.OldString, req.NewString)
    } else {
        newContent = strings.Replace(contentStr, req.OldString, req.NewString, 1)
    }

    // Write back
    if err := os.WriteFile(req.FilePath, []byte(newContent), 0644); err != nil {
        return nil, fmt.Errorf("failed to write file: %w", err)
    }

    replaced := 1
    if req.ReplaceAll {
        replaced = count
    }

    return &EditResponse{
        Success:  true,
        Replaced: replaced,
    }, nil
}
```

---

### Week 7: Pattern Matching (Glob & Grep)

#### Task 7.1: Glob Tool

Create `internal/tool/glob.go`:

```go
package tool

import (
    "path/filepath"
    "sort"
    "strings"

    "github.com/gobwas/glob"
)

type GlobRequest struct {
    Pattern string `json:"pattern" validate:"required"`
    Path    string `json:"path"` // Optional, defaults to CWD
}

type GlobResponse struct {
    Files []string `json:"files"`
}

func (t *ToolExecutor) Glob(req GlobRequest) (*GlobResponse, error) {
    basePath := req.Path
    if basePath == "" {
        basePath = "."
    }

    // Compile glob pattern
    g, err := glob.Compile(req.Pattern, '/')
    if err != nil {
        return nil, err
    }

    var files []string

    // Walk directory tree
    err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // Skip errors
        }

        // Skip directories
        if info.IsDir() {
            return nil
        }

        // Check if path matches pattern
        relPath, _ := filepath.Rel(basePath, path)
        if g.Match(relPath) {
            files = append(files, path)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // Sort by modification time (newest first)
    sort.Slice(files, func(i, j int) bool {
        infoI, _ := os.Stat(files[i])
        infoJ, _ := os.Stat(files[j])
        return infoI.ModTime().After(infoJ.ModTime())
    })

    return &GlobResponse{Files: files}, nil
}
```

Install glob library:

```bash
go get github.com/gobwas/glob
```

#### Task 7.2: Grep Tool (Ripgrep Integration)

Create `internal/tool/grep.go`:

```go
package tool

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os/exec"
)

type GrepRequest struct {
    Pattern string `json:"pattern" validate:"required"`
    Path    string `json:"path"`    // Optional
    Include string `json:"include"` // File pattern like "*.go"
}

type GrepResponse struct {
    Matches []GrepMatch `json:"matches"`
}

type GrepMatch struct {
    File  string `json:"file"`
    Line  int    `json:"line"`
    Text  string `json:"text"`
}

func (t *ToolExecutor) Grep(req GrepRequest) (*GrepResponse, error) {
    // Check if ripgrep is available
    if _, err := exec.LookPath("rg"); err != nil {
        return nil, fmt.Errorf("ripgrep not found: %w", err)
    }

    args := []string{
        "--json",         // JSON output
        "--no-heading",   // Don't group by file
        req.Pattern,
    }

    if req.Include != "" {
        args = append(args, "--glob", req.Include)
    }

    if req.Path != "" {
        args = append(args, req.Path)
    }

    cmd := exec.Command("rg", args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // Run command (ignore exit code - rg exits 1 if no matches)
    _ = cmd.Run()

    // Parse JSON output
    var matches []GrepMatch
    scanner := bufio.NewScanner(&stdout)
    for scanner.Scan() {
        var result struct {
            Type string `json:"type"`
            Data struct {
                Path struct {
                    Text string `json:"text"`
                } `json:"path"`
                Lines struct {
                    Text string `json:"text"`
                } `json:"lines"`
                LineNumber int `json:"line_number"`
            } `json:"data"`
        }

        if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
            continue
        }

        if result.Type == "match" {
            matches = append(matches, GrepMatch{
                File: result.Data.Path.Text,
                Line: result.Data.LineNumber,
                Text: result.Data.Lines.Text,
            })
        }
    }

    return &GrepResponse{Matches: matches}, nil
}
```

---

### Week 8: Bash Execution

#### Task 8.1: Bash Tool

Create `internal/tool/bash.go`:

```go
package tool

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "time"
)

type BashRequest struct {
    Command string `json:"command" validate:"required"`
    Timeout int    `json:"timeout"` // milliseconds
}

type BashResponse struct {
    Output   string `json:"output"`
    Error    string `json:"error"`
    ExitCode int    `json:"exitCode"`
    Duration int64  `json:"duration"` // milliseconds
}

func (t *ToolExecutor) Bash(req BashRequest) (*BashResponse, error) {
    // Default timeout: 2 minutes
    timeout := 120000
    if req.Timeout > 0 {
        timeout = req.Timeout
    }
    if timeout > 600000 { // Max 10 minutes
        timeout = 600000
    }

    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
    defer cancel()

    // Execute command with shell
    cmd := exec.CommandContext(ctx, "sh", "-c", req.Command)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    start := time.Now()
    err := cmd.Run()
    duration := time.Since(start).Milliseconds()

    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        } else {
            return nil, fmt.Errorf("command execution failed: %w", err)
        }
    }

    // Truncate output if too large (30,000 chars max)
    output := stdout.String()
    if len(output) > 30000 {
        output = output[:30000] + "\n[Output truncated...]"
    }

    return &BashResponse{
        Output:   output,
        Error:    stderr.String(),
        ExitCode: exitCode,
        Duration: duration,
    }, nil
}
```

---

### Week 9-10: Project Management

#### Task 9.1: Project Detection

Create `internal/project/project.go`:

```go
package project

import (
    "os"
    "path/filepath"
)

type Project struct {
    ID   string `json:"id"`
    Path string `json:"path"`
    Name string `json:"name"`
    Git  *GitInfo `json:"git,omitempty"`
}

type GitInfo struct {
    Root     string `json:"root"`
    Branch   string `json:"branch"`
    Worktree bool   `json:"worktree"`
}

func Detect(path string) (*Project, error) {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return nil, err
    }

    name := filepath.Base(absPath)

    project := &Project{
        ID:   generateID(absPath),
        Path: absPath,
        Name: name,
    }

    // Detect git repository
    if gitInfo := detectGit(absPath); gitInfo != nil {
        project.Git = gitInfo
    }

    return project, nil
}

func detectGit(path string) *GitInfo {
    gitDir := findGitDir(path)
    if gitDir == "" {
        return nil
    }

    // Check if worktree
    worktree := false
    if info, err := os.Stat(filepath.Join(gitDir, "commondir")); err == nil && !info.IsDir() {
        worktree = true
    }

    // Get current branch
    branch := getCurrentBranch(gitDir)

    return &GitInfo{
        Root:     filepath.Dir(gitDir),
        Branch:   branch,
        Worktree: worktree,
    }
}

func findGitDir(path string) string {
    current := path
    for {
        gitPath := filepath.Join(current, ".git")
        if _, err := os.Stat(gitPath); err == nil {
            return gitPath
        }

        parent := filepath.Dir(current)
        if parent == current {
            return ""
        }
        current = parent
    }
}

func getCurrentBranch(gitDir string) string {
    headFile := filepath.Join(gitDir, "HEAD")
    content, err := os.ReadFile(headFile)
    if err != nil {
        return "unknown"
    }

    // Parse "ref: refs/heads/main"
    line := strings.TrimSpace(string(content))
    if strings.HasPrefix(line, "ref: ") {
        ref := strings.TrimPrefix(line, "ref: ")
        return filepath.Base(ref)
    }

    return "detached"
}

func generateID(path string) string {
    hash := sha256.Sum256([]byte(path))
    return fmt.Sprintf("%x", hash[:8])
}
```

#### Task 9.2: Ignore Patterns

Create `internal/file/ignore.go`:

```go
package file

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"

    "github.com/gobwas/glob"
)

type IgnorePatterns struct {
    patterns []glob.Glob
}

func LoadIgnorePatterns(projectPath string) (*IgnorePatterns, error) {
    ig := &IgnorePatterns{
        patterns: make([]glob.Glob, 0),
    }

    // Load .gitignore
    gitignorePath := filepath.Join(projectPath, ".gitignore")
    if err := ig.loadFromFile(gitignorePath); err == nil {
        // Successfully loaded
    }

    // Add default patterns
    ig.addPattern("node_modules/**")
    ig.addPattern(".git/**")
    ig.addPattern("dist/**")
    ig.addPattern("build/**")
    ig.addPattern("*.log")

    return ig, nil
}

func (ig *IgnorePatterns) loadFromFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        ig.addPattern(line)
    }

    return scanner.Err()
}

func (ig *IgnorePatterns) addPattern(pattern string) {
    g, err := glob.Compile(pattern, '/')
    if err == nil {
        ig.patterns = append(ig.patterns, g)
    }
}

func (ig *IgnorePatterns) ShouldIgnore(path string) bool {
    for _, pattern := range ig.patterns {
        if pattern.Match(path) {
            return true
        }
    }
    return false
}
```

---

## Integration

### Update Server Handlers

Modify `internal/server/handlers.go` to use real implementations:

```go
type Server struct {
    storage  *storage.Storage
    tools    *tool.ToolExecutor
    projects *project.Manager
}

func (s *Server) handleFilesRead(w http.ResponseWriter, r *http.Request) {
    var req tool.ReadRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    resp, err := s.tools.Read(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(resp)
}

// Similar for Write, Edit, Glob, Grep, Bash...
```

---

## Testing

### Unit Tests

- [ ] Test file read with various offsets/limits
- [ ] Test file write creates directories
- [ ] Test edit with single/multiple replacements
- [ ] Test glob pattern matching
- [ ] Test grep with ripgrep
- [ ] Test bash execution and timeouts
- [ ] Test project detection (git/non-git)
- [ ] Test ignore patterns

### Integration Tests

- [ ] End-to-end file operations from API
- [ ] Test bash command execution via API
- [ ] Test project initialization
- [ ] Test glob/grep with real project

### Manual Testing with TUI

- [ ] Browse files in TUI
- [ ] Edit files and verify changes
- [ ] Search code with grep
- [ ] Execute bash commands
- [ ] Verify project detection

---

## Deliverables

- [x] File read/write/edit tools implemented
- [x] Glob pattern matching functional
- [x] Grep (ripgrep integration) working
- [x] Bash execution with timeout support
- [x] Project detection and git integration
- [x] Ignore pattern handling (.gitignore)
- [x] API handlers connected to tools
- [x] Comprehensive test coverage

---

## Success Metrics

- [ ] Can read files with line numbers (like cat -n)
- [ ] Can write new files (creates directories)
- [ ] Can edit files with exact string replacement
- [ ] Glob returns files matching pattern
- [ ] Grep finds code matches across project
- [ ] Bash executes commands with output/error
- [ ] Project auto-detected with git info
- [ ] Ignore patterns working (.gitignore honored)

---

## Phase 2 Checklist

### File Tools

- [ ] Implement Read tool
- [ ] Implement Write tool
- [ ] Implement Edit tool
- [ ] Add validation for file paths
- [ ] Handle file permission errors
- [ ] Support line offset/limit in Read
- [ ] Truncate long lines (>2000 chars)

### Pattern Matching

- [ ] Implement Glob tool
- [ ] Install glob library
- [ ] Support recursive directory walking
- [ ] Sort results by modification time
- [ ] Implement Grep tool
- [ ] Integrate with ripgrep
- [ ] Parse JSON output from ripgrep
- [ ] Support include patterns (\*.go)

### Bash Execution

- [ ] Implement Bash tool
- [ ] Add timeout support (default 2min, max 10min)
- [ ] Capture stdout and stderr
- [ ] Return exit code
- [ ] Truncate large output (>30,000 chars)
- [ ] Context cancellation

### Project Management

- [ ] Implement project detection
- [ ] Detect git repositories
- [ ] Get current branch name
- [ ] Detect git worktrees
- [ ] Generate project IDs
- [ ] Load .gitignore patterns
- [ ] Implement ignore matching
- [ ] Add default ignore patterns

### API Integration

- [ ] Wire tools to API handlers
- [ ] Add request validation
- [ ] Add error handling
- [ ] Return appropriate HTTP status codes
- [ ] Update Server struct with dependencies

### Testing

- [ ] Unit tests for each tool
- [ ] Test edge cases (missing files, permissions)
- [ ] Test bash timeout behavior
- [ ] Test ignore pattern matching
- [ ] Integration tests with API
- [ ] Manual TUI testing

---

## Next Phase

**Phase 3**: Session Management (Weeks 11-16)

- Session CRUD operations
- Message storage
- Event streaming
- Session state management
