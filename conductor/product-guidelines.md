# Product Guidelines

## Tone and Voice

### Technical and Precise
All documentation and user-facing content should use clear, concise language focused on accuracy and developer efficiency. Avoid marketing fluff, unnecessary adjectives, and vague statements.

**Do:**
- "Conductor reads `product.md` to establish project context"
- "Run `/conductor:new-track` to create a specification and plan"

**Don't:**
- "Conductor amazingly transforms your development experience"
- "Simply run the easy-to-use new-track command"

## Brand Messaging

### Primary Tagline
**"Measure twice, code once."**

This captures the core philosophy: invest time in planning and specification before writing code to reduce rework and improve quality.

### Technical Positioning
**"Context-driven development"**

Emphasizes the unique approach of treating project context as a managed artifact that persists across all AI interactions.

### Role Description
**"Your AI's project manager"**

Highlights Conductor's function as an orchestration layer that guides AI agents through structured workflows.

## Terminology Conventions

### Conductor-Specific Terms
Use these terms consistently throughout all documentation:

| Term | Definition | Avoid |
|------|------------|-------|
| **Track** | A high-level unit of work (feature or bug fix) | "feature branch", "ticket", "issue" |
| **Phase** | A logical grouping of tasks within a track | "stage", "step", "milestone" |
| **Spec** | The specification document (`spec.md`) | "requirements doc", "PRD" |
| **Plan** | The implementation plan (`plan.md`) | "task list", "to-do list" |

### Code Formatting
Always use monospace/code formatting for:
- Commands: `/conductor:setup`, `/conductor:implement`
- File paths: `conductor/product.md`, `conductor/tracks/`
- Directory names: `conductor/`, `tracks/`
- Status markers: `[ ]`, `[x]`

## Interactive Questions Format

When presenting questions to users, follow the AskUserQuestion tool specification:

### Question Structure
- **header**: Very short label, maximum 12 characters (e.g., "Tech Stack", "Workflow")
- **options**: Between 2-4 options per question
- **label**: Display text, 1-5 words per option
- **description**: Brief explanation of what the option means or implies

### Multi-Select vs Single Choice
- Use `multiSelect: true` for additive questions (brainstorming, scope definition)
- Use `multiSelect: false` for exclusive choices (single commitment decisions)

### Option Ordering
- Place the recommended option **first** in the list
- Append `(Recommended)` to the recommended option's label
- The "Other" option is provided automatically by the system—do not include it manually

### Example Format
```
Question: "Which testing approach should we use?"
Header: "Testing"

Options:
1. label: "TDD (Recommended)"
   description: "Write tests before implementation code"
2. label: "Test after"
   description: "Write tests after feature implementation"
3. label: "No tests"
   description: "Skip automated testing for this track"
```

## Visual and Structural Guidelines

### Markdown-First
All generated documents use GitHub-flavored Markdown with consistent heading hierarchy:
- `#` for document title (one per file)
- `##` for major sections
- `###` for subsections
- `####` sparingly, for detailed breakdowns

### Status Indicators
Use checkbox syntax for task tracking in plans:
- `- [ ]` for pending tasks
- `- [x]` for completed tasks

Nested tasks use indentation:
```markdown
- [ ] Task: Implement user authentication
    - [ ] Write unit tests
    - [ ] Implement login endpoint
    - [ ] Add session management
```

### Section Separation
Use horizontal rules (`---`) to separate major document sections, particularly between:
- Document metadata and content
- Distinct logical sections
- Phase boundaries in plans

## Error Messages and User Feedback

### Actionable Guidance
Error messages must tell users what to do next, not just what went wrong.

**Do:**
- "No track found. Run `/conductor:new-track` to create one."

**Don't:**
- "Error: Track not found."

### Reference Documentation
Point users to relevant commands or files when issues occur:
- "See `conductor/workflow.md` for configuration options"
- "Run `/conductor:status` to view current progress"

### Minimal Verbosity
Keep messages concise. Provide essential information without lengthy explanations.

**Do:**
- "Track `auth_20240115` marked complete. Run `/conductor:new-track` for next feature."

**Don't:**
- "Congratulations! The track you were working on, which was identified as `auth_20240115`, has now been successfully marked as complete in the system. If you would like to continue working on your project, you may want to consider running the `/conductor:new-track` command to begin planning your next feature or bug fix."
