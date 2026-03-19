---
phase: quick
plan: 260319-tck
type: execute
wave: 1
depends_on: []
files_modified:
  - README.md
autonomous: true
requirements: []
must_haves:
  truths:
    - "README mentions get-shit-done (GSD) as the AI coding workflow used to build the project"
    - "README documents available Docker image tags (latest and SHA-based)"
  artifacts:
    - path: "README.md"
      provides: "Updated project documentation"
      contains: "get-shit-done"
  key_links: []
---

<objective>
Update README.md with a GSD attribution section and Docker image tag documentation.

Purpose: Help users understand available Docker image tags for pulling from ghcr.io, and credit the AI coding workflow used to build the project.
Output: Updated README.md
</objective>

<context>
@README.md
@.github/workflows/ci.yml
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add GSD mention and Docker image tags to README</name>
  <files>README.md</files>
  <action>
Make the following additions to README.md:

1. **Add a "Built With" or "Acknowledgments" section** before the License section (near bottom of file). Include a brief mention that this project was built using [get-shit-done](https://github.com/lmignot/get-shit-done), an AI coding workflow for Claude Code. Keep it to 1-2 sentences -- factual, not promotional.

2. **Add a "Docker Image Tags" subsection** inside the existing "## Docker" section, after the introductory bullet list (after line 237, before "### Building the Image"). Document the two tag strategies from ci.yml:
   - `latest` -- always points to the most recent build from the `main` branch
   - `<sha>` -- short Git commit SHA (e.g., `abc1234`) for pinning to a specific build

   Include a small example showing how to pull a specific tag:
   ```
   docker pull ghcr.io/lm/karaclean:latest
   docker pull ghcr.io/lm/karaclean:<sha>
   ```

   Note that pinning to a SHA tag is recommended for production to avoid unexpected changes.

3. **Verify the personal note** already exists on line 13. Do NOT duplicate it. If it is missing for some reason, add it after the introductory paragraph (after "...evaluating every bookmark against your rules each cycle.") with the exact text: "Note: this project has been an exploration of AI coding tools for me. Although I do use karaclean with my own Karakeep instance, use it at your own risk!"
  </action>
  <verify>
    <automated>grep -q "get-shit-done" README.md && grep -q "latest" README.md && grep -q "sha" README.md && echo "PASS" || echo "FAIL"</automated>
  </verify>
  <done>README.md contains GSD attribution, documents both Docker image tag strategies (latest and SHA), and has the personal note</done>
</task>

</tasks>

<verification>
- README.md renders correctly in markdown preview
- GSD section exists with link to repository
- Docker image tags subsection documents `latest` and SHA tags
- Personal note present (already exists, not duplicated)
</verification>

<success_criteria>
- grep for "get-shit-done" in README.md returns a match
- grep for "ghcr.io/lm/karaclean:latest" in README.md returns a match
- Docker tags section explains both latest and SHA pinning
- No duplicate personal note
</success_criteria>

<output>
No summary file needed for quick plans.
</output>
