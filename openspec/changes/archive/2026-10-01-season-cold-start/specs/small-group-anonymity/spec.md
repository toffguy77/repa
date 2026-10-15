# Spec Delta

## Purpose

Protects the product's core promise — that votes are anonymous — in groups small enough
that aggregated percentages or a voter list would identify individual voters, and
guarantees members are never shown a meaningless reputation card.

## ADDED Requirements

### Requirement: Detector is unavailable in groups below five members

The detector SHALL be unavailable while a group has fewer than 5 members, because a voter
list drawn from at most 4 possible voters identifies individuals. The request SHALL fail
with a dedicated error and SHALL NOT deduct crystals.

#### Scenario: Purchase is refused in a small group

- **WHEN** a member of a 4-person group attempts to buy a detector for a revealed season
- **THEN** the request fails with error code `GROUP_TOO_SMALL`
- **AND** the member's crystal balance is unchanged
- **AND** no detector record is created

#### Scenario: Detector is offered once the group grows

- **GIVEN** a group that has grown to 5 members
- **WHEN** a member buys a detector for a revealed season
- **THEN** the purchase succeeds and the voter list is returned

#### Scenario: Detector status reports unavailability

- **WHEN** a member of a 4-person group reads detector status for a revealed season
- **THEN** the response indicates the detector is unavailable for this group size rather than merely unpurchased

### Requirement: Members are warned while anonymity is weak

While a group has fewer than 5 members, the system SHALL tell members that results in a
group this small are not fully anonymous, so they can decide what to vote before the
Reveal rather than after.

#### Scenario: Notice is shown in a small group

- **GIVEN** a group with 3 members
- **WHEN** a member opens the group
- **THEN** the app shows a notice that in groups under 5 people results can reveal who voted how

#### Scenario: Notice disappears once the group is large enough

- **GIVEN** a group with 5 members
- **WHEN** a member opens the group
- **THEN** no small-group anonymity notice is shown

#### Scenario: Notice is shown before voting starts

- **GIVEN** a group with 2 members
- **WHEN** a member starts the voting session
- **THEN** the small-group anonymity notice is visible before the first vote is cast

### Requirement: No member is ever shown an empty card

A reputation card SHALL only exist for a season that legitimately revealed. The system
SHALL NOT present a card with no attributes, and SHALL NOT generate card images for a
season that did not reveal.

#### Scenario: No card image for an unrevealed season

- **GIVEN** a season that was postponed because the participation floor was not met
- **WHEN** card generation for that season would otherwise run
- **THEN** no card images are generated and no card records are created

#### Scenario: Waiting state instead of an empty card

- **GIVEN** a member of a group whose season has not revealed
- **WHEN** the member opens the reveal screen
- **THEN** the app shows a waiting state explaining what the Reveal is waiting for, not an empty card
