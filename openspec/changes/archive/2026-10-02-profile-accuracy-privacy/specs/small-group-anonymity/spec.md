# Spec Delta

## ADDED Requirements

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
