# Proposal

## Why

Repa's visual language is a light-mode Material default with nine flat colours
(`mobile/lib/core/theme/app_colors.dart`), six text styles
(`mobile/lib/core/theme/app_text_styles.dart`), and a single `ThemeData`
(`mobile/lib/core/theme/app_theme.dart:5`) wired as `theme: AppTheme.light` with no dark
theme (`mobile/lib/main.dart:44`). There are no spacing, radius, elevation, or motion
tokens at all, so the 18 screens carry 64 ad-hoc `BorderRadius.circular(...)` literals and
61 raw `Colors.white`/`Colors.black` references — every new surface re-invents its own
proportions.

That is a problem beyond tidiness. The product targets 14–22-year-olds and competes for
attention with Gas, BeReal, and Discord, all of which are dark, dense, and
motion-forward; a white Material screen reads as a utility app. The shareable reputation
card — the only organic acquisition channel — is already a dark purple 1080×1920 artifact
(`backend/internal/service/cards/template.go`), so the app and its own hero asset do not
even look like the same product.

## What Changes

- **BREAKING (internal API):** `AppColors` and `AppTextStyles` static constants are
  replaced by a token layer resolved from `Theme.of(context)`. Every screen reads colour
  and type from the theme rather than from global constants, so a second theme is possible
  at all.
- A documented token set is introduced — colour (with dark and light palettes), type
  scale, spacing, radius, border, elevation, and motion (duration + curve) — each token
  named by role rather than by value.
- The app becomes **dark-first**: dark is the default theme and light is the alternate,
  with `themeMode` following the system setting. Both palettes are held to WCAG AA
  contrast for text and interactive elements.
- A small component kit is introduced for the patterns the app repeats: primary/secondary/
  ghost buttons, card surface, chip, progress bar, bottom sheet, list tile, stat/percentage
  readout, and section header.
- Six high-traffic surfaces are restyled against the tokens: onboarding, home/groups list,
  group detail, voting, reveal, and the crystal shop. Remaining screens inherit the new
  look through `ThemeData` without bespoke work.
- Motion is standardised: one duration/curve pair per interaction class (tap feedback,
  surface transition, reveal emphasis), replacing the per-widget `flutter_animate` timings
  chosen ad hoc.

## Capabilities

### New Capabilities

- `design-tokens`: the named, documented set of visual and motion primitives the app is
  built from, and the guarantees they carry (contrast, scale, a single source of truth).
- `app-theming`: how a theme is selected and applied — dark-first default, system
  following, and the requirement that no surface hardcodes a colour or radius.

### Modified Capabilities

None. The existing `season-lifecycle` and `small-group-anonymity` specs describe backend
behaviour and are untouched; this change restyles the screens that render them without
altering what they say.

## Impact

- **Mobile:** new `lib/core/theme/` token files and `lib/core/widgets/` component kit;
  edits across all 18 screens to drop hardcoded values; `main.dart` theme wiring.
- **Tests:** existing widget tests that assert on `AppColors.*` constants or on raw
  `Colors.white` need updating; new tests for token contrast and theme resolution.
- **Backend:** the card template (`internal/service/cards/template.go`) is brought onto
  the same palette so the shared PNG matches the app.
- **Docs:** a new `docs/features/design-system.md`; the UI sections of existing feature
  docs stop naming raw colours.
- **No API, schema, or business-logic changes.**

## Non-goals

- Replacing Material as the widget foundation, or adopting a third-party component library
  — the kit wraps Material rather than competing with it.
- Custom fonts. The MVP ships the system font (per `docs/specs/master_context.md`); the
  type scale is defined so a brand face can be swapped in later without touching screens.
- Localisation or copy rewrites — strings stay hardcoded Russian.
- Redesigning the shared card's *layout*; only its palette is aligned with the app.
- Animated illustrations, Lottie, or custom shaders.
