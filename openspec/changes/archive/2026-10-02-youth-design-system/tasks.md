# Tasks

## 1. Token layer

- [x] 1.1 Add `mobile/lib/core/theme/app_tokens.dart` with `AppTokens extends ThemeExtension<AppTokens>` carrying colour, spacing, radius, border, elevation, and motion groups plus `copyWith`/`lerp`, and a `BuildContext.t` getter; verify `flutter analyze` passes and a unit test asserts `lerp` returns the target at t=1
- [x] 1.2 Define the dark and light colour palettes as `AppTokens.dark` / `AppTokens.light` with role-named fields (canvas, surface, surfaceRaised, border, accent, accentPressed, onAccent, energy, onEnergy, textPrimary, textSecondary, textInverse, success, warning, danger); verify a test asserts both palettes define every field (no nulls, no shared instance)
- [x] 1.3 Add `mobile/test/core/theme/contrast_test.dart` enumerating every (text token, surface token) pair the app uses and asserting WCAG ratios — 4.5:1 for body text, 3:1 for large text and control boundaries — in both palettes; adjust palette values until it passes
- [x] 1.4 Define the type scale in `app_tokens.dart` (display1, display2, heading1, heading2, body, bodyStrong, caption, label, numeric) with explicit size/weight/height/letterSpacing, font family resolved in one place, and tabular figures on `numeric`; verify a test asserts the family swaps for every role when changed in that one place
- [x] 1.5 Define spacing (4/8/12/16/24/32/48), radius (6/10/14/20/28/full), border widths, and the elevation triples (surface colour + border colour + shadow list, shadows empty in dark); verify a test asserts the spacing scale has no value that is not a multiple of 4 and that dark elevations carry no shadows
- [x] 1.6 Define motion tokens (tap 120ms/easeOut, surface 220ms/easeOutCubic, emphasis 420ms/easeOutBack) and a `motionDuration(context, class)` helper that returns `Duration.zero` when `MediaQuery.disableAnimationsOf` is true; verify tests cover both the normal and the reduced-motion result

## 2. Theme wiring

- [x] 2.1 Rewrite `mobile/lib/core/theme/app_theme.dart` to build `ThemeData` for both brightnesses from `AppTokens`, registering the extension and configuring appBar, elevatedButton, outlinedButton, textButton, input, bottomSheet, card, chip, progressIndicator, and snackBar themes from tokens; verify a test resolves `AppTokens` from each theme and asserts the scaffold background matches the canvas token
- [x] 2.2 Wire `theme`, `darkTheme`, and `themeMode: ThemeMode.system` in `mobile/lib/main.dart` so the platform appearance selects the theme; verify a widget test pumps the app under each `platformBrightness` and asserts the resolved canvas colour
- [x] 2.3 Add a widget test that rebuilds a screen when `platformBrightness` changes mid-session and asserts the new palette applies without the widget being recreated; verify it fails if `themeMode` is hardcoded
- [x] 2.4 Create `docs/features/design-system.md` documenting every token group, both palettes with hex values, the role-naming rule, the 44px target rule, the reduced-motion rule, and the Dart↔Go card palette pairing with its known drift limitation; verify every token named in the doc exists in `app_tokens.dart`

## 3. Component kit

- [x] 3.1 Add `AppButton` (primary/secondary/ghost variants, loading and disabled states) wrapping Material buttons, binding tokens and enforcing a 44px minimum target; verify widget tests cover each variant's enabled/disabled/loading rendering and assert the target size
- [x] 3.2 Add `AppCard` and `AppSectionHeader` using the surface/border/elevation and heading tokens; verify widget tests assert they take the active theme's tokens under both brightnesses with no per-call overrides
- [x] 3.3 Add `AppChip`, `AppListTile`, and `AppProgressBar`; verify widget tests assert the list tile's row height is at least 44px at compact density and that the progress bar renders both determinate and indeterminate states
- [x] 3.4 Add `AppStat` (large numeral + label, tabular figures) and an `AppSheet` helper for bottom sheets; verify widget tests assert the numeral uses the `numeric`/display role and that the sheet uses the sheet radius token
- [x] 3.5 Document the kit in `docs/features/design-system.md` with one line per component stating what it wraps and which tokens it binds; verify each documented component exists and is exported from a single barrel file

## 4. Core shared widgets migration

- [x] 4.1 Convert `mobile/lib/core/widgets/` (empty state, error state, skeleton loader, reveal countdown) to tokens and the kit, removing every colour and radius literal; verify the four existing widget test files still pass and `grep` finds no `Color(0x` or `Colors.white` in `lib/core/widgets/`
- [x] 4.2 Convert `lib/core/theme` consumers that are not screens (router transitions, analytics-free helpers) to motion tokens; verify `flutter analyze` passes and no `Duration(milliseconds:` literal remains in `lib/core/`

## 5. Bespoke surfaces

- [x] 5.1 Restyle onboarding (`lib/features/onboarding/`) against tokens: dark canvas, display type, emphasis motion on slide entrance; verify existing onboarding behaviour is unchanged by running the full mobile suite and that the screen renders under both themes in a widget test
- [x] 5.2 Restyle home/groups list (`home_screen.dart`, `group_card.dart`) with compact density and the card/stat components; verify `home_screen_test.dart` still passes and a new test asserts both themes render
- [x] 5.3 Restyle group detail (`group_screen.dart`) including the season card, `PendingRevealNotice`, and `SmallGroupNotice` so the notices use semantic tokens rather than `AppColors.warning`; verify the notice widget tests still pass after updating them to assert roles instead of hex values
- [x] 5.4 Restyle voting (`voting_screen.dart`, `question_card.dart`, `participant_card.dart`, `voting_complete_screen.dart`) with the tap/surface motion classes; verify the 31 existing voting tests still pass
- [x] 5.5 Restyle reveal (`reveal_screen.dart`, `reputation_card.dart`, `detector_sheet.dart`, `achievement_popup.dart`) using display/numeric type and the emphasis motion class; verify the reveal and detector widget tests still pass and the detector's unavailable state still renders its explanation
- [x] 5.6 Restyle the crystal shop (`lib/features/crystals/`) with the card, chip, and stat components and the energy accent for crystal amounts; verify a widget test renders the shop under both themes

## 6. Remaining screens and constant removal

- [x] 6.1 Convert the remaining screens (auth phone/otp/profile-setup, join group, create group, profile, member profile, settings, force update, telegram setup, question vote) from `AppColors`/`AppTextStyles` to tokens; verify `flutter analyze` passes and `grep -r "AppColors\.\|AppTextStyles\." lib/` returns nothing
- [x] 6.2 Delete `app_colors.dart` and `app_text_styles.dart`; verify `flutter analyze` and the full `flutter test` suite pass with no references remaining
- [x] 6.3 Update the tests that asserted on removed constants or raw colours to assert on token roles; verify the full mobile suite passes and no test file contains a `Color(0x` literal outside the token and contrast tests
- [x] 6.4 Update the UI sections of `docs/features/*.md` that name raw colours to name token roles instead; verify no feature doc mentions `AppColors` or a hex value except `design-system.md`

## 7. Shared card palette

- [x] 7.1 Extract the colour values in `backend/internal/service/cards/template.go` into a single `cardPalette` block whose comment names the Dart tokens it mirrors, and align them with the dark palette; verify `go build ./...` passes and the rendered HTML contains the new values
- [x] 7.2 Update `backend/internal/service/cards/template_test.go` to assert the palette hex values and that the template still escapes user content; verify `go test ./internal/service/cards/...` passes
- [x] 7.3 Update `docs/features/cards.md` with the new palette and a pointer to the Dart tokens it must stay in step with; verify the documented hex values match `template.go`

## 8. Integration verification

- [x] 8.1 Render every screen under both themes in a smoke test and assert no widget throws and no text uses a colour that fails the contrast pairs asserted in 1.3
- [x] 8.2 Run `flutter analyze`, the full `flutter test` suite, and `go test ./...`; confirm `openspec validate youth-design-system --strict` passes
