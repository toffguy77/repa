# member-safety Specification

## Purpose
What a member can do to get away from another member in an anonymous rating product, and what the
system guarantees once they have.

## Requirements

### Requirement: A member can block another member

A member SHALL be able to block any other member they share a group with. A block SHALL take effect
immediately and SHALL be reversible by the member who created it.

#### Scenario: Blocking takes effect at once

- **WHEN** a member blocks another member
- **THEN** the block is in force without any further action

#### Scenario: Unblocking restores the previous state

- **GIVEN** a member who blocked someone
- **WHEN** they unblock that person
- **THEN** both appear to each other as before

#### Scenario: Blocking is idempotent

- **GIVEN** a member who already blocked someone
- **WHEN** they block them again
- **THEN** the request succeeds and nothing changes

#### Scenario: A member cannot block themselves

- **WHEN** a member attempts to block their own account
- **THEN** the request is refused

### Requirement: A block removes the other person from voting, in both directions

While a block exists, neither member SHALL appear as a voting target for the other. A block SHALL be
mutual in effect even though only one person created it — the product must never ask someone to rate a
person who blocked them.

#### Scenario: A blocked member is not offered as a target

- **GIVEN** I blocked another member of my group
- **WHEN** I open a voting session
- **THEN** that member is not among the targets

#### Scenario: The blocker is not offered to the blocked member either

- **GIVEN** another member blocked me
- **WHEN** I open a voting session
- **THEN** that member is not among my targets

#### Scenario: A vote for a blocked member is refused

- **GIVEN** a block between me and another member
- **WHEN** I attempt to cast a vote for them anyway
- **THEN** the vote is refused

#### Scenario: Votes cast before the block remain

- **GIVEN** I voted for someone and then blocked them
- **WHEN** the season reveals
- **THEN** the earlier vote still counts, because retroactively editing results would expose that a
  block happened

### Requirement: A block hides the other member's results

While a block exists, neither member SHALL see the other's reputation card, and neither SHALL appear in
the other's view of the group's results.

#### Scenario: A blocked member's card is not shown

- **GIVEN** a block between two members and a revealed season
- **WHEN** either opens the members' cards
- **THEN** the other's card is absent

#### Scenario: My own card is unaffected

- **GIVEN** a block
- **WHEN** I open my own card
- **THEN** it shows my own results in full

### Requirement: A member can report another member

A member SHALL be able to report another member with an optional reason. A report SHALL reach the same
administrative queue that content reports use, and SHALL NOT notify the reported member.

#### Scenario: A report reaches the queue

- **WHEN** a member reports another member
- **THEN** the report appears in the administrative queue with both identities and the reason

#### Scenario: The reported member is not told

- **WHEN** a member is reported
- **THEN** they receive no notification about it

#### Scenario: Reporting is limited to once per person

- **GIVEN** a member who already reported someone
- **WHEN** they report the same person again
- **THEN** the request succeeds without creating a second report

#### Scenario: Reporting does not require blocking

- **WHEN** a member reports someone without blocking them
- **THEN** the report is created and no block is

### Requirement: An admin can remove a member

A group admin SHALL be able to remove a member from the group. The removed member SHALL lose access to
the group immediately.

#### Scenario: A removed member loses access

- **WHEN** an admin removes a member
- **THEN** that member can no longer read the group or vote in it

#### Scenario: Only the admin can remove

- **WHEN** a non-admin attempts to remove a member
- **THEN** the request is refused and the member remains

#### Scenario: An admin cannot remove themselves this way

- **WHEN** an admin attempts to remove their own account from the group
- **THEN** the request is refused and they are directed to leave instead

### Requirement: Removal and permanent departure cannot be undone by an invite

A member who was removed, or who left permanently, SHALL NOT be able to re-join that group with an
invite. Ordinary leaving SHALL remain reversible.

#### Scenario: A removed member cannot re-join

- **GIVEN** a member removed by the admin
- **WHEN** they open the group's invite link
- **THEN** joining is refused

#### Scenario: A permanent departure cannot be reversed by a link

- **GIVEN** a member who left permanently
- **WHEN** they use the invite link
- **THEN** joining is refused

#### Scenario: An ordinary leave can be reversed

- **GIVEN** a member who left normally
- **WHEN** they use the invite link
- **THEN** they re-join

#### Scenario: A regenerated invite does not undo a removal

- **GIVEN** a removed member and an admin who regenerated the invite
- **WHEN** the removed member uses the new link
- **THEN** joining is still refused
