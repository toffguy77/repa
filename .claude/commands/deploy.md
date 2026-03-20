---
description: Use when the user wants to deploy a new build, bump version, or ship. Triggers on /deploy command.
---

# Deploy

Bump build number, verify, push. CI handles the rest.

## Flow

1. **Check prerequisites:**
   - Current branch is `main` (warn if not)
   - Working tree is clean (`git status` shows no changes)
   - Backend: `cd backend && npm run lint && npm test`
   - Mobile: `cd mobile && flutter analyze && flutter test`

2. **Bump build number** in `mobile/pubspec.yaml`:
   - Parse current version (e.g., `1.0.0+5`)
   - Increment only the build number after `+` (e.g., `1.0.0+6`)
   - Do NOT change the version string — that is set manually by the developer

3. **Commit and push:**
   ```bash
   git add mobile/pubspec.yaml
   git commit -m "chore: bump build number to {new_build_number}"
   git push origin main
   ```

4. **Confirm:** Tell user the push was made and CI will handle deployment.

## Notes

- If tests or lint fail, fix issues first. Do not deploy broken code.
- If the branch is not main, ask user if they want to proceed anyway.
- Backend deployment details TBD — will be configured when infra is set up.
