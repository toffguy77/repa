# Tasks

## 1. Schema and classification

- [x] 1.1 Add migration `010_question_tone.up.sql` creating the `question_tone` enum (`WARM`, `NEUTRAL`, `EDGY`), `questions.tone` defaulting to `NEUTRAL`, and `groups.kind_only` defaulting to false, plus the down migration; verify both apply and roll back cleanly against Postgres
- [x] 1.2 Classify the seeded questions in the same migration by matching curated phrase lists, so an existing database is corrected without re-seeding; verify the migration runs against a seeded database and leaves no question without a tone
- [x] 1.3 Add tone to the seeder so a fresh database arrives classified, using the same lists; verify a unit test asserts the seeder's classifier agrees with the migration's for a representative sample
- [x] 1.4 Add a test pinning the classification: assert the distribution across tones and spot-check named questions on both the warm and edgy lists; verify it fails if a question moves between lists

## 2. Selection and group setting

- [x] 2.1 Add the tone filter to the season question selection query so a kind-only group never receives an edgy question, filtering inside the query rather than after the limit; verify `make sqlc` regenerates and a query test asserts an edgy question is absent for a kind-only group
- [x] 2.2 Accept `kind_only` on group creation, defaulting to on for creators under 18 and off otherwise, with an explicit value always winning; verify unit tests cover a minor, an adult, and both explicit overrides
- [x] 2.3 Accept `kind_only` on group update (admin only) and expose it in group responses; verify handler tests cover the update, the non-admin refusal, and the field's presence
- [x] 2.4 Confirm changing the setting leaves an open season untouched; verify a unit test asserts the season's questions are unchanged after the setting flips
- [x] 2.5 Refuse a kind-only group whose categories offer no kind question, asking the bank rather than a hardcoded category list, at both creation and setting change; verify unit tests cover the refusal on create, the refusal on update, one kind category being enough, and an ordinary group being unaffected
- [x] 2.6 Update `docs/features/groups.md` and `docs/features/voting.md` with the tone, the setting, the default rule and the open-season rule; verify every documented rule has a test

## 3. Card ordering

- [x] 3.1 Order a card's attributes by percentage with tone breaking near-ties inside a 5-point band, never reordering past a clearly stronger result; implement it as anchor clusters rather than a comparison function, since the band is not transitive; verify unit tests cover an exact tie, a within-band difference, an out-of-band difference, and a chain of near-ties spanning more than the band
- [x] 3.2 Confirm tone-aware ordering removes nothing; verify a unit test asserts every received attribute is still present
- [x] 3.4 Put the rule in `internal/cardorder` and route every card path through it — the reveal API, the members-cards list, the hidden-attributes view, and the PNG worker — carrying tone through each feeding query; verify each path has its own test, since an unset tone reads as neutral and degrades to percentage order without failing anything
- [x] 3.3 Update `docs/features/cards.md` and `docs/features/reveal.md` with the ordering rule and the band; verify the documented band matches the constant

## 4. Moderation

- [x] 4.1 Have the question moderation step assign a tone to a user-submitted question, defaulting to neutral when it cannot be determined; verify unit tests cover a warm, an edgy and an undeterminable submission
- [x] 4.2 Update `docs/features/moderation.md` with tone assignment and its independence from the allow/reject verdict; verify the documented behaviour matches the tests

## 5. Mobile

- [x] 5.1 Add the kind-only switch to group creation, defaulted from the user's age and labelled with what it does to the questions, sending the field only when the user actually toggled it so the server's age default is not silently overridden; verify widget tests cover the default for a minor, an adult and an unknown birth year, the explanatory copy, and that an untouched switch sends nothing
- [x] 5.2 Expose the setting in the group's Freezed model and allow the admin to change it; verify model tests round-trip the field and a widget test asserts the admin-only control
- [x] 5.3 Update `docs/features/groups.md` mobile section; verify each documented element has a test

## 6. Integration verification

- [x] 6.1 Add e2e coverage: a kind-only group's season contains no edgy question, while an ordinary group's bank is unrestricted
- [x] 6.2 Add e2e coverage: an under-18 creator gets the setting on by default, an adult does not, and an explicit value wins
- [x] 6.3 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate question-tone --strict` passes
