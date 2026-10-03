# Design System

## Overview

Repa's visual language is a token layer plus a small component kit. Tokens are resolved from
the ambient theme, never from static constants — that is what makes two complete palettes
possible. Names describe the **role** a value plays, not what it looks like, so changing a
hue never requires renaming a token.

Reference points: **shadcn/ui** (context-resolved, role-named tokens), **Linear** (one
timing per interaction class), **Geist** (layered dark surfaces instead of shadows).

The product targets 14–22-year-olds. The look is deliberately dark, compact and
motion-forward; density comes from spacing and type, never from shrinking touch targets.

## Access

```dart
final t = context.t;              // AppTokens for the active theme
t.color.accent                    // role-named colour
AppTokens.space.lg                // 16
AppTokens.radius.card             // BorderRadius.circular(20)
context.motion(MotionClass.tap)   // Duration, zero under reduced motion
```

`context.t` throws if the extension is missing. That is deliberate: a missing extension is a
wiring bug that should fail at the first build, not fall back to values that look almost
right.

## Themes

Both themes are complete — every colour role is defined in both. `main.dart` supplies
`theme`, `darkTheme`, and `themeMode: ThemeMode.system`, so the platform appearance selects
the palette.

**Dark is the reference palette:** it is designed and contrast-checked first, the component
kit is tuned against it, and the shared reputation card matches it. Light is a complete,
equally tested alternate rather than the baseline. Flutter's `platformBrightness` is always
either light or dark — there is no third "unspecified" value — so "dark when the system has
no preference" is not something the app can express, and it does not claim to.

## Colour tokens

| Role | Dark | Light | Used for |
|---|---|---|---|
| `canvas` | `#0B0712` | `#FFFFFF` | Page background |
| `surface` | `#161022` | `#F6F4FB` | A panel on the canvas |
| `surfaceRaised` | `#1F1733` | `#FFFFFF` | A panel on a panel; sheets, dialogs |
| `border` | `#2E2446` | `#E6E1F0` | Hairline separators, panel outlines |
| `accent` | `#9B6DFF` | `#7C3AED` | Brand as **foreground**: icons, links, emphasis |
| `accentFill` | `#7C3AED` | `#7C3AED` | Brand as a **filled control background** |
| `accentFillPressed` | `#6228CC` | `#6228CC` | Pressed state of a filled control |
| `onAccentFill` | `#FFFFFF` | `#FFFFFF` | Text/icons on `accentFill` |
| `energy` | `#C8F751` | `#4D7C0F` | Second voice: streaks, crystals, "new" |
| `energyFill` | `#C8F751` | `#4D7C0F` | Filled energy background |
| `onEnergyFill` | `#17210A` | `#FFFFFF` | Text on `energyFill` |
| `textPrimary` | `#F6F3FF` | `#150E24` | Body and headings |
| `textSecondary` | `#AFA3CC` | `#5C5375` | De-emphasised text |
| `success` | `#3DDC97` | `#047857` | Positive state |
| `warning` | `#FFC45A` | `#A16207` | Non-blocking caution |
| `danger` | `#FF7A8A` | `#DC2626` | Destructive or failed |
| `scrim` | `#000000` @70% | `#000000` @40% | Behind modals and sheets |

**Why `accent` and `accentFill` are separate roles:** a filled control's background must be
dark enough for white text (≥ 4.5:1), while a foreground accent must be light enough to read
on a near-black canvas (≥ 3:1). In dark mode no single value satisfies both — `#7C3AED` fails
as a foreground on the dark canvas for text, and `#9B6DFF` fails to carry white text. The
split is the fix, not a redundancy.

**`energy` is restricted to small areas** by convention — streak badges, crystal amounts,
"new" markers. Violet plus acid lime is what keeps the UI from reading as corporate; used
over large areas it stops being a highlight.

### Contrast

`mobile/test/core/theme/contrast_test.dart` enumerates every (text token, surface token)
pairing the app renders and asserts the WCAG 2.1 ratio: **4.5:1** for text, **3:1** for large
text and the boundary of interactive controls, in **both** palettes. Palette values were
adjusted until that test passed. The test is the definition of "readable" here — not anyone's
eye. Adding a new pairing to the app means adding it to that list.

## Elevation

A level is a **triple** — surface colour, outline colour, shadow list — not a number, because
dark mode expresses depth by stacking surfaces while light mode casts shadows. A drop shadow
on a near-black canvas is invisible, so a numeric elevation scale would silently stop working
in the default theme.

| Level | Dark | Light |
|---|---|---|
| `level0` | canvas surface, hairline outline, **no shadow** | white, hairline outline, no shadow |
| `level1` (card) | `#161022`, hairline outline, **no shadow** | white + soft 12px shadow |
| `level2` (sheet, dialog) | `#1F1733`, hairline outline, **no shadow** | white + 28px shadow |

## Spacing

One geometric scale, 4-logical-pixel base: `xs 4`, `sm 8`, `md 12`, `lg 16`, `xl 24`,
`xxl 32`, `huge 48`. Deliberately dense at the small end. Screens default to the 12/16 steps
where a conventional layout would use 16/24. A value not on the scale is not representable —
use the nearest step.

## Radius

Radii diverge by role rather than being uniform; tight controls against soft cards is what
makes the density read as intentional instead of cramped.

| Token | Value | Used for |
|---|---|---|
| `xs` | 6 | Chip, badge, tag |
| `sm` | 10 | Button, input |
| `md` | 14 | List tile, inline notice |
| `lg` | 20 | Card |
| `xl` | 28 | Bottom sheet, modal (top corners only) |
| `full` | 999 | Pill |

## Type scale

System font (per `docs/specs/master_context.md`), resolved via
`AppTokens.text.fontFamily` in exactly one place so a brand face can replace it without
touching a screen.

| Role | Size/Line | Weight | Tracking | Used for |
|---|---|---|---|---|
| `display1` | 40/44 | 800 | −1.2 | Reveal percentages, hero numerals |
| `display2` | 32/36 | 800 | −0.8 | Screen hero text |
| `heading1` | 24/28 | 700 | −0.4 | Screen titles |
| `heading2` | 19/24 | 700 | −0.2 | Section and card titles |
| `body` | 15/22 | 400 | — | Body text |
| `bodyStrong` | 15/22 | 600 | — | Buttons, emphasised body |
| `caption` | 13/18 | 400 | — | Secondary and helper text |
| `label` | 11/14 | 700 | +0.8 | Uppercase micro-labels |
| `numeric` | 15/20 | 700 | tabular | Counters, percentages |

Body is 15 and caption 13 — one step below convention. `numeric` uses tabular figures so a
changing counter does not shift layout.

## Motion

Screens pick an interaction **class**, never a number, so two comparable interactions can
never animate at different speeds.

| Class | Duration | Curve | Used for |
|---|---|---|---|
| `tap` | 120ms | `easeOut` | Press feedback, selection, small state flips |
| `surface` | 220ms | `easeOutCubic` | A surface appearing, route transition, list reflow |
| `emphasis` | 420ms | `easeOutBack` | The Reveal — the only place motion is the point |

`context.motion(class)` returns `Duration.zero` when `MediaQuery.disableAnimationsOf` is
true. Durations collapse rather than widgets disappearing, so layout cannot shift between the
two modes, and no information is ever conveyed by movement alone.

## Touch targets

Every interactive component keeps a **44 logical pixel** minimum target in both dimensions,
however small it is drawn. This is enforced inside the component kit, not left to each
screen. It is the one place density is not allowed to win.

## Component kit

`mobile/lib/core/widgets/kit/` — each component wraps the Material widget that already
handles semantics, focus and ink. The kit binds tokens and enforces the touch target; it does
not re-implement interaction.

| Component | Wraps | Binds |
|---|---|---|
| `AppButton` (primary / secondary / ghost) | `ElevatedButton` / `OutlinedButton` / `TextButton` | `accentFill`, `onAccentFill`, `radius.control`, `bodyStrong`, 44px target |
| `AppCard` | `Container` + `Material` | `elevation.level1`, `radius.card`, `border.hairline` |
| `AppSectionHeader` | `Row` + `Text` | `heading2`, `label`, `space.md` |
| `AppChip` | `Material` + `InkWell` | `radius.chip`, `caption`, `accentFill` when selected |
| `AppListTile` | `InkWell` | `radius.tile`, `bodyStrong`/`caption`, 44px row |
| `AppProgressBar` | `LinearProgressIndicator` | `accent`, `surface` track, `radius.full` |
| `AppStat` | `Column` + `Text` | `display2`/`numeric` + `label` |
| `AppSheet` | `showModalBottomSheet` | `elevation.level2`, `radius.sheet`, `scrim` |

`ThemeData` is also configured for the wrapped Material widgets (app bar, buttons, inputs,
card, chip, sheet, snack bar, list tile, tabs, FAB, progress), so a screen that was never
restyled by hand still picks up the look.

## Shared reputation card (Go)

The server-rendered PNG (`backend/internal/service/cards/template.go`) mirrors the **dark**
palette in a single `cardPalette` block. There is no build step shared between Dart and Go,
so these two definitions must be changed together:

| Card value | Mirrors |
|---|---|
| background gradient | `canvas` → `surface` (dark) |
| accent / bar fill | `accentFill` → `accent` (dark) |
| primary text | `textPrimary` (dark) |
| secondary text | `textSecondary` (dark) |

**Known limitation:** a Dart-side change does not break the Go test, so the pairing above is
the only guard. A generated shared token file is the right fix and is deliberately out of
scope — it is a build-system change, not a design one.

## Rules

1. No colour, radius, or duration literal in a screen. If a value is missing, add a token.
2. Token names say what a value is *for*, never what it looks like.
3. A new (text, surface) pairing goes into the contrast test before it goes into a screen.
4. Interactive targets stay at 44px regardless of visual size.
5. Decorative motion must survive reduced-motion by collapsing, not by disappearing.
