# Tasks

## 1. Withhold the figure

- [x] 1.1 Make `StatsDto.GuessAccuracy` a `*float64` with `omitempty`, populated in `GetProfile` only when the viewer is the viewed member; verify unit tests cover own profile (present), another member's (absent), and a member with no stats row (absent, and not zero)
- [x] 1.2 Confirm the omission happens in the service rather than the handler, so a future caller does not get the figure by default; verify a test calls the service directly and asserts the field is nil for a different viewer
- [x] 1.3 Confirm the remaining statistics are unaffected; verify a test asserts seasons played, both streaks, and both vote totals are still present on another member's profile
- [x] 1.4 Assert the wire form, not just the struct; verify a handler test decodes the JSON and asserts the `guess_accuracy` key is absent for another member and present for oneself
- [x] 1.5 Update `docs/features/profile.md` with who may read the figure and why, naming both inference routes (published winners, and differencing the rolling average); verify the documented rule has a test

## 2. Mobile

- [x] 2.1 Make `StatsDto.guessAccuracy` nullable and hide the accuracy tile when it is absent; verify model tests round-trip both shapes and a widget test covers the tile's presence on one's own profile and absence on another's
- [x] 2.2 Update `docs/features/profile.md` mobile section; verify each documented element has a test

## 3. Record the rule where it will be read again

- [x] 3.1 State the convention in `CLAUDE.md` as "prefer publishing an order over a measurement", with the reason, so it applies to the next per-member statistic rather than only to this field; verify the wording names what a figure can be combined with
- [x] 3.2 Update `docs/specs/master_context.md` anonymity rules with the same; verify it matches `CLAUDE.md`

## 4. Integration verification

- [x] 4.1 Add e2e coverage: member A reading member B's profile gets no `guess_accuracy`, A reading their own gets it, and a non-member is refused
- [x] 4.2 Add e2e coverage: the chronicle standing and the profile agree — neither exposes a figure for another member
- [x] 4.4 Fix `StatCard` reading `MediaQuery` from `initState` via `context.motion` — a defect introduced by the design-token work, which threw in any test mounting the card and in release left the controller's duration ignoring a change in the reduce-motion preference; verify a widget test mounts the card with no exception and one with `disableAnimations: true`
- [x] 4.3 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate profile-accuracy-privacy --strict` passes
