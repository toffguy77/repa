# small-group-anonymity Specification

## Purpose
Protects the product's core promise — that votes are anonymous — in groups small enough
that aggregated percentages or a voter list would identify individual voters, and
guarantees members are never shown a meaningless reputation card.

## Requirements

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

### Requirement: The small-group floor applies to every detector rung

In a group below the detector's minimum size, **no** rung of the detector SHALL be purchasable — not
the hint and not the full list — because a partial reveal drawn from two or three possible voters
identifies someone just as surely as a list does.

#### Scenario: A hint is refused in a small group

- **WHEN** a member of a four-person group attempts to buy a hint
- **THEN** the request fails with the group-too-small error
- **AND** their balance is unchanged

#### Scenario: The free count is still shown in a small group

- **GIVEN** a four-person group whose season has revealed
- **WHEN** a member opens the detector
- **THEN** they still see how many people voted, since a count names nobody

### Requirement: Blocks count against the anonymity floor

The group-size thresholds that protect anonymity SHALL be evaluated against the number of members a
person can actually be rated by, not the raw membership count — otherwise a group of five with three
blocks presents itself as safely anonymous while behaving like a group of two.

#### Scenario: A blocked-down group shows the small-group warning

- **GIVEN** a group of five members where I have blocked two of them
- **WHEN** I open the group
- **THEN** I see the small-group anonymity warning

#### Scenario: The detector respects the effective size

- **GIVEN** a group whose effective size for me is below the detector's floor
- **WHEN** I attempt to buy any paid detector rung
- **THEN** the request is refused with the group-too-small error

### Requirement: The chronicle carries the small-group warning with it

A chronicle entry SHALL be presented with the same warning the live results carry when the percentages
it shows were computed over too few voters to hide who voted. The chronicle SHALL NOT become a way to
read small-group results without the caveat that applies to them.

The judgement is made on the number of voters the percentage was computed over, recorded with the
result, rather than on the group's membership at the time: a share of two voters identifies them
whatever the group's size was, and a group's historic membership is not recoverable once members leave.

#### Scenario: A historic entry computed over too few voters is marked

- **GIVEN** a chronicle entry whose percentages were computed over four voters
- **WHEN** a member reads it
- **THEN** it is marked as coming from a result too small for the percentages to hide who voted

#### Scenario: A historic entry with enough voters is not marked

- **GIVEN** a chronicle entry whose percentages were computed over twelve voters
- **WHEN** a member reads it
- **THEN** no warning is attached

#### Scenario: The warning follows the result, not the group's size today

- **GIVEN** a group that has grown from four members to twelve
- **WHEN** a member reads an entry from a season only three members voted in
- **THEN** the entry is still marked, because the warning is about the result that was produced

#### Scenario: The threshold is the one the detector uses

- **WHEN** the chronicle decides whether to mark an entry
- **THEN** it uses the same minimum the detector is gated on, rather than a second number

### Requirement: A derived figure counts as exposing a vote

A per-member statistic SHALL be treated as exposing that member's votes whenever the API also publishes
something it can be combined with to recover them. Removing vote identifiers from a response does not by
itself satisfy anonymity.

In particular, a figure measuring how often a member's votes matched the group's result SHALL NOT be
disclosed to any member other than that member, because the winner of each question is published.

#### Scenario: A member's match rate is not shown to another member

- **GIVEN** a member viewing another member's profile in their group
- **WHEN** the profile is returned
- **THEN** it does not contain a figure for how often that member's votes matched the group's result

#### Scenario: A member sees their own figure

- **GIVEN** a member viewing their own profile
- **WHEN** the profile is returned
- **THEN** the figure is present, because it describes votes they cast themselves

#### Scenario: Withheld is distinguishable from zero

- **GIVEN** a member viewing another member's profile
- **WHEN** the profile is returned
- **THEN** the figure is absent rather than reported as zero, so it cannot be read as "never matched
  anything"

#### Scenario: Statistics that are not match counts stay visible

- **GIVEN** a member viewing another member's profile
- **WHEN** the profile is returned
- **THEN** how many seasons they played, their voting streak, and how many votes they cast and received
  are all still present, because none of them counts matches

### Requirement: An order may be published where a measurement may not

Where the product needs to compare members on how well they predict the group, it SHALL publish their
relative order rather than the underlying figures.

#### Scenario: The group standing ranks without measuring

- **WHEN** the group's standing of who predicts it best is returned
- **THEN** it gives each member's position and how many seasons it is based on, and no figure
