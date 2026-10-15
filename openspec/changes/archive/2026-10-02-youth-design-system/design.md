# Design

## Context

See proposal.md — Why. What exists today: one light `ThemeData`, nine flat colour
constants, six static `TextStyle`s, and no tokens for spacing, radius, elevation, or
motion. Screens reach for `AppColors.x` directly, which is why a second theme is currently
impossible — a static const cannot vary by context.

The reference brief is the `designsystems.one` gallery; the systems worth borrowing method
from here are **shadcn/ui** (role-named tokens resolved from context, never value-named),
**Linear** (one timing per interaction class, fast easings), and **Geist** (layered dark
surfaces instead of shadows). The product constraint from `docs/specs/master_context.md`
stands: system font, hardcoded Russian strings, `#7C3AED` as the brand accent.

## Goals / Non-Goals

**Goals:**

- Make a second theme possible at all, by moving from static constants to context-resolved
  tokens.
- Make the compact, dark, high-contrast look the default a new screen falls into, rather
  than something each screen has to re-achieve.
- Prove the contrast guarantee with a test rather than with judgement.

**Non-Goals:**

- A user-facing theme switch in settings. The system setting is the only input; an in-app
  override is a separate change with its own persistence question.
- Per-screen pixel-perfect redesign of all 18 screens. Six get bespoke attention; the rest
  inherit.
- Sharing token values between Dart and Go at build time (see Risks).

## Decisions

### `ThemeExtension` for tokens, not a global singleton or an InheritedWidget of our own

Tokens live in `AppTokens extends ThemeExtension<AppTokens>`, registered on both
`ThemeData`s and read through a `BuildContext` extension:

```dart
extension TokensX on BuildContext {
  AppTokens get t => Theme.of(this).extension<AppTokens>()!;
}
```

Alternatives considered:

- *Keep static constants, add a second set for dark.* Rejected: every widget would need to
  branch on brightness itself, which is precisely the duplication that lets one theme rot.
- *A bespoke `InheritedWidget`.* Rejected: `ThemeExtension` already gives `lerp` (so theme
  changes animate) and participates in `Theme.of`, which the Material widgets we wrap are
  already reading.

The `!` is deliberate: a missing extension is a wiring bug that should fail loudly at the
first build, not silently fall back to defaults that look almost right.

### Dark is the default, and it is layered rather than shadowed

`themeMode: ThemeMode.system` with both `theme` and `darkTheme` supplied; Flutter picks dark
when the platform reports dark appearance.

Note on "dark-first": Flutter's `platformBrightness` is always either light or dark — there
is no third "unspecified" value to fall back from, so "dark when the system has no
preference" is not expressible and the spec does not claim it. Dark-first here means dark is
the *reference* palette: it is designed and contrast-checked first, the component kit is
tuned against it, and the shared card matches it. Light is a complete, equally tested
alternate rather than the baseline.

Depth in dark mode is expressed by **stacked surface tokens plus a hairline border**
(`canvas` → `surface` → `surfaceRaised`), not by shadows: a drop shadow on a near-black
canvas is invisible, so a shadow-based elevation scale would simply stop working in the
default theme. The light theme does use soft shadows, which is why `elevation` is a token
*triple* — surface colour, border colour, and shadow list — rather than a number. Each
theme fills in what reads in that theme.

### The palette is chosen to pass a contrast test, not to look right in isolation

Dark canvas is a near-black with a violet cast (`#0B0712`) so the brand hue feels native
rather than applied on top of grey. The accent lightens from the brand `#7C3AED` to
`#8B5CF6` in dark mode, because the brand value does not clear 4.5:1 for text drawn on the
dark canvas; `#7C3AED` is retained as the pressed/gradient-end value where it sits under
white text. The light theme uses `#7C3AED` unchanged.

A `energy` accent (lime) is added for streaks, crystals, and "new" markers. Violet + acid
lime is the current youth idiom and gives the UI a second voice without a second brand
colour; it is restricted to small areas by convention, and it carries a text token of its
own so it is never used as a text colour against the canvas.

Every (text token, surface token) pair the app actually uses is enumerated in a test that
computes the WCAG ratio. Values were adjusted until that test passed — the test is the
specification of "readable", not anyone's eye.

### A compact scale, with the tightness in spacing and type rather than in touch targets

Base unit 4. Body type drops from 16 to 15 and caption from 14 to 13; an uppercase
`label` role at 11/700/+0.8 tracking supplies the dense, "tiny" texture. Screens default to
the 12/16 steps where they previously used 16/24.

Radius deliberately diverges by role rather than being uniform: controls get small radii
(10) and cards get large ones (20), sheets larger still (28). The contrast between tight
controls and soft cards is what makes the density read as intentional instead of cramped.

Interactive targets stay at 44 logical pixels regardless of how small the control is drawn.
This is the one place density is not allowed to win, and it is enforced in the component kit
(each interactive component wraps its visual in a minimum-size box) rather than left to each
screen.

### Motion: three classes, one timing each, and a reduced-motion escape hatch

`tap` (120ms, easeOut), `surface` (220ms, easeOutCubic), `emphasis` (420ms, easeOutBack).
Screens pick a class, never a number. The Reveal is the only consumer of `emphasis`, which
is what makes it feel like the event the product is built around.

`MediaQuery.disableAnimationsOf(context)` gates decorative animation. Reduced motion
collapses durations to zero rather than removing the widgets, so layout cannot shift between
the two modes.

### The component kit wraps Material, it does not replace it

Eight components (`AppButton` with primary/secondary/ghost variants, `AppCard`, `AppChip`,
`AppProgressBar`, `AppSheet`, `AppListTile`, `AppStat`, `AppSectionHeader`) each wrap the
Material widget that already handles semantics, focus, and ink. The kit's job is to bind
tokens and enforce the touch target, not to re-implement interaction. This keeps the
accessibility behaviour Flutter already gives us and keeps the diff per screen small.

`ThemeData` is also configured for the wrapped widgets (`elevatedButtonTheme`,
`inputDecorationTheme`, `appBarTheme`, `bottomSheetTheme`, …) so screens that were never
restyled still pick up the new look. That is what makes "six screens bespoke, twelve
inherited" a coherent outcome rather than a half-finished one.

### The card template duplicates the palette, with the duplication made explicit

The server renders the shared PNG in Go; there is no build step shared with Dart. Rather
than pretend otherwise, the Go template's colours move into a single
`cardPalette` block whose comment names the Dart tokens it mirrors, and a test asserts the
specific hex values. If someone changes the Dart token without the Go one, that test still
passes — so the real guard is the documented pairing in
`docs/features/design-system.md`, and this is stated as a known limitation rather than a
solved problem.

## Risks / Trade-offs

- **Dart and Go palettes can drift** → the duplication is isolated to one block per side and
  documented as a pair; a shared generated token file is the right fix but is a build-system
  change out of scope here.
- **Dark-first inverts the look for existing users** → there are no production users yet
  (the app is pre-release per `docs/specs/T24_release_checklist.md`), so this is the cheapest
  moment to make the change.
- **Tests asserting `AppColors.*` or `Colors.white` will break** → they are updated to assert
  on roles via the theme; a test that hardcodes a hex value is asserting the wrong thing and
  is rewritten rather than patched.
- **Twelve screens are only theme-restyled, not redesigned** → they will look consistent but
  not composed. Accepted deliberately: the alternative is a change too large to review, and
  the token layer makes each later polish pass cheap.
- **Removing `AppColors`/`AppTextStyles` touches every screen at once** → the edits are
  mechanical (constant → token) and covered by `flutter analyze` plus the existing 143
  widget tests, which must stay green throughout.

## Migration Plan

1. Add the token layer and both themes alongside the existing constants; nothing breaks yet.
2. Add the contrast and token tests; adjust values until green.
3. Add the component kit, built on tokens.
4. Convert screens group by group (core widgets → list/home → group/voting → reveal → shop →
   remainder), keeping `flutter analyze` and `flutter test` green after each group.
5. Delete `AppColors` and `AppTextStyles` once no references remain — the deletion is the
   proof that the migration is complete.
6. Align the Go card template and its test.
7. Rollback: the change is additive until step 5, so reverting the theme wiring in
   `main.dart` restores the previous look without touching screens.
