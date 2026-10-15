# group-invites Specification

## Purpose
How a person gets into a Repa group: the shape of an invite code, what the system accepts
when someone types or pastes one, and the guarantees it makes about uniqueness and
revocation.

## Requirements

### Requirement: An invite code is short enough to say out loud

A newly generated invite code SHALL be 6 characters long and SHALL be drawn from an alphabet
that excludes characters people confuse when transcribing: the digits `0` and `1`, and the
letters `O`, `I`, and `L`.

#### Scenario: A new group's code is transcribable

- **WHEN** a group is created
- **THEN** its invite code is 6 characters long
- **AND** contains no `0`, `1`, `O`, `I`, or `L`

#### Scenario: Codes are not sequential or guessable from one another

- **WHEN** many groups are created in sequence
- **THEN** their invite codes are not derivable from each other

### Requirement: Input is normalised before lookup

The system SHALL accept an invite code regardless of letter case and surrounding whitespace,
and SHALL also accept a full invite URL in place of a bare code. The user SHALL NOT have to
know which form is expected.

#### Scenario: Lowercase input finds the group

- **GIVEN** a group whose invite code is `AB2CD3`
- **WHEN** a user submits `ab2cd3`
- **THEN** the group is found

#### Scenario: Surrounding whitespace is ignored

- **WHEN** a user submits `  AB2CD3 `
- **THEN** the group is found

#### Scenario: A pasted link is accepted

- **GIVEN** a user pastes `https://repa.app/join/AB2CD3` into the join field
- **WHEN** the app looks the group up
- **THEN** the group is found
- **AND** the user does not have to strip the link down to the code themselves

#### Scenario: A legacy code keeps its own separators

- **GIVEN** a group whose invite code is a 36-character hyphenated identifier
- **WHEN** a user submits that code
- **THEN** the group is found, because the hyphens are part of the stored value rather than display separators

#### Scenario: A wrong code is refused clearly

- **WHEN** a user submits a code no group has
- **THEN** the request fails with a not-found error rather than joining some other group

### Requirement: Codes are unique and verified at generation

The system SHALL guarantee that no two groups share an invite code, comparing
case-insensitively, and SHALL verify uniqueness at the moment of generation rather than
relying on probability alone.

#### Scenario: A collision is retried, not accepted

- **GIVEN** a freshly generated code that already belongs to another group
- **WHEN** the system generates an invite code
- **THEN** it generates another until the code is free
- **AND** the created group has a code no other group holds

#### Scenario: Case-insensitive uniqueness

- **GIVEN** a group with code `AB2CD3`
- **WHEN** a code `ab2cd3` would be generated
- **THEN** it is treated as taken

#### Scenario: Group creation fails rather than issuing a duplicate

- **GIVEN** repeated collisions that exhaust the generator's attempts
- **WHEN** a group is created
- **THEN** creation fails with an error and no group is created with a duplicate code

### Requirement: Regenerating an invite revokes the previous code

When a group admin regenerates the invite link, the system SHALL issue a new short code and
the previous code SHALL stop working immediately.

#### Scenario: The old code stops working

- **GIVEN** a group with code `AB2CD3`
- **WHEN** the admin regenerates the invite link
- **THEN** a different 6-character code is returned
- **AND** joining with `AB2CD3` fails with a not-found error

#### Scenario: Only the admin can revoke

- **WHEN** a non-admin member requests a new invite link
- **THEN** the request is refused and the existing code keeps working

### Requirement: Links already shared keep working

Invite codes issued before this change SHALL continue to resolve, so links already sent to
people are not broken by the new format.

#### Scenario: A legacy code still joins

- **GIVEN** a group whose invite code is a 36-character identifier from before this change
- **WHEN** a user opens that invite link
- **THEN** the group is found and the user can join

#### Scenario: A legacy code is replaced on regeneration

- **GIVEN** a group still holding a legacy code
- **WHEN** the admin regenerates the invite link
- **THEN** the new code is in the short format

### Requirement: The app presents the code as something to read aloud

The app SHALL display an invite code in a form built for transcription — visually grouped,
unambiguous, and copyable — and SHALL tell the user they can share either the link or the
code.

#### Scenario: The code is shown alongside the link

- **WHEN** a member opens the invite sheet for their group
- **THEN** both the short code and the full link are visible
- **AND** each can be copied independently

#### Scenario: The join field accepts either form

- **WHEN** a user opens the join screen
- **THEN** the field's hint indicates that a code or a link is accepted

### Requirement: An invite can be delivered by scanning

An invite SHALL be deliverable as a scannable code in addition to a link and a typed code, so
a person looking at someone's screen or a printed card can join without typing anything.

#### Scenario: Scanning a card's code opens the join flow

- **GIVEN** a card showing a group's scannable code
- **WHEN** a person scans it with a phone camera
- **THEN** the invite link opens and offers to join that group

#### Scenario: The scannable code encodes the same invite as the text code

- **WHEN** a card is generated for a group
- **THEN** the scannable code resolves to the same group as the printed code next to it

### Requirement: An invite does not override a removal

An invite SHALL NOT admit a person who was removed from that group or who left it permanently,
regardless of which invite code they present.

#### Scenario: A banned person is refused with a distinct error

- **GIVEN** a person who was removed from a group
- **WHEN** they attempt to join with a valid invite code
- **THEN** the request fails with an error distinguishing this from an unknown code

#### Scenario: Other groups are unaffected

- **GIVEN** a person removed from one group
- **WHEN** they join a different group with its own invite
- **THEN** they join normally
