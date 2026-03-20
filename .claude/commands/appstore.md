---
description: Use when the user wants to update App Store metadata, screenshots, submit for review, or manage App Store Connect. Triggers on /appstore command.
---

# App Store Operations

Update metadata, screenshots, and submit for App Store review.

## Flow

1. **Ask what to do:**
   - Update metadata (descriptions, keywords, release notes)
   - Update screenshots
   - Submit for review
   - Check review status

2. **For metadata updates:**
   - Edit metadata files in the fastlane metadata directory (path TBD when fastlane is configured)
   - Available files: name.txt, subtitle.txt, description.txt, keywords.txt (max 100 chars), promotional_text.txt, release_notes.txt
   - Commit changes

3. **For screenshot updates:**
   - Generate/compose screenshots (tooling TBD)
   - Sizes: 6.7" (1290x2796), 6.5" (1242x2688), iPad 13" (2048x2732)

4. **For submission:**
   - Run fastlane submission (command TBD when fastlane is configured)
   - Requires App Store Connect API credentials from env vars

5. **For status check:**
   - Use App Store Connect API with JWT auth
   - Credentials should be in env vars (ASC_KEY_ID, ASC_ISSUER_ID, ASC_API_KEY_PATH)

## Metadata constraints

- App name: max 30 characters
- Subtitle: max 30 characters
- Keywords: max 100 characters (comma-separated)
- Description: max 4000 characters
- Release notes: max 4000 characters
