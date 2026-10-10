# Repositories

This project is **one system spread across several git repositories**. Conductor plans and tracks it as a single project: one `conductor/` directory, one tracks registry and one plan per track, while the code lives in the repositories below.

**Track branch:** `feat/<track_id>`

| Name | Path | Base branch |
|------|------|-------------|
| <name> | <path relative to the project root> | <base branch> |

## Multi-Repo Protocol

Every Conductor command follows these rules whenever this file exists. They override any single-repository git instruction in the Conductor commands and in `workflow.md`.

1. **The project root is not a repository.** Never run a bare `git` command in the project root. Always target a repository explicitly: `git -C <path> ...`.
2. **`conductor/` is its own local repository.** All Conductor bookkeeping (setup files, tracks registry, specs, plans, metadata, archive) is committed with `git -C conductor ...`. It has no remote unless the user adds one. Code is never committed there.
3. **Code goes to the repository that owns the file.** A file belongs to the repository whose path is its prefix. When one step changes files in several repositories, make one commit per repository, each with the same message. Run each repository's own tests from its own path.
4. **One track branch everywhere.** A track uses exactly one branch name, the **Track branch** above with `<track_id>` filled in, in every repository it touches.
    - Before the first change in a repository, run `git -C <path> status --porcelain`. If it is not empty, stop and ask the user whether to commit, stash or stop. Never stash, reset or discard on your own.
    - Then create the branch from the repository's base branch (`git -C <path> switch -c <track branch> <base branch>`), or switch to it if it already exists (`git -C <path> switch <track branch>`).
    - Add the repository's name to the `"repos"` list in the track's `metadata.json`.
    - Never commit track work to a base branch.
5. **Commit references name their repository.** Write `<name>@<sha7>`, e.g. `api@a1b2c3d`. A task or checkpoint that spans repositories lists every reference, separated by spaces: `api@a1b2c3d portal@d4e5f6a`. Bookkeeping commits in `conductor/` are not referenced in plans.
6. **Git notes** are attached in the repository that holds the commit.
7. **Reviews and reverts work per repository.** Group the plan's references by repository name and handle each repository on its own: its own revision range, its own `git diff`, its own `git revert`, all with `git -C <path>`.
8. **One pull request per repository.** When the track is complete, every repository in the track's `"repos"` has one track branch that becomes one pull request, all with the same title. Each description names the track and links the sibling pull requests. Never push a branch or open a pull request without the user's explicit confirmation.
