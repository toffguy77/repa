# Spec Delta

## ADDED Requirements

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
