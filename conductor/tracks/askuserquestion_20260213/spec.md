# Specification: Improve AskUserQuestion Integration

## Overview

Update all Conductor skills to use the AskUserQuestion tool format consistently, aligning interactive prompts with the specification defined in `product-guidelines.md`.

## Problem Statement

The current Conductor skills use a text-based question format with lettered options (A, B, C, D, E). While functional, this format does not leverage Claude Code's native `AskUserQuestion` tool, which provides:
- Structured question presentation
- Multi-select support
- Automatic "Other" option handling
- Consistent UI across all Claude Code interactions

## Goals

1. **Consistency**: All skills present questions in the same format
2. **Native Integration**: Leverage Claude Code's built-in tooling
3. **Better UX**: Provide clearer option descriptions and recommendations
4. **Maintainability**: Centralize question format guidelines in `product-guidelines.md`

## Scope

### In Scope

- Update `/conductor:setup` skill to use AskUserQuestion format guidelines
- Update `/conductor:new-track` skill to use AskUserQuestion format guidelines
- Update `/conductor:implement` skill (if interactive questions exist)
- Update `/conductor:review` skill (if interactive questions exist)
- Document the format specification in `product-guidelines.md` (already done)

### Out of Scope

- Changes to `/conductor:status` (read-only, no questions)
- Changes to `/conductor:revert` (confirmation-only, minimal interaction)
- Runtime code changes (plugin has no runtime code)

## Requirements

### Functional Requirements

1. **FR-1**: All multi-choice questions must specify:
   - `header`: Max 12 characters
   - `options`: 2-4 items with label (1-5 words) and description
   - `multiSelect`: true/false based on question type

2. **FR-2**: Recommended options must be listed first with "(Recommended)" suffix

3. **FR-3**: Do not include "Other" option explicitly (system provides it automatically)

4. **FR-4**: Use `multiSelect: true` for additive questions (brainstorming, scope)

5. **FR-5**: Use `multiSelect: false` for exclusive choice questions (single commitment)

### Non-Functional Requirements

1. **NFR-1**: No changes to skill behavior or logic
2. **NFR-2**: Backward compatible with existing conductor projects
3. **NFR-3**: Documentation must be clear and actionable

## Acceptance Criteria

- [ ] All SKILL.md files updated with AskUserQuestion format guidelines
- [ ] Questions follow the structure defined in `product-guidelines.md`
- [ ] No references to lettered options (A, B, C, D, E) in question instructions
- [ ] Each skill tested manually to verify question presentation

## Technical Notes

- Skills are Markdown files with YAML frontmatter
- Changes are documentation/instruction changes only
- No version bump required (internal improvement)
