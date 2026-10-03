# Spec Delta

## ADDED Requirements

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
