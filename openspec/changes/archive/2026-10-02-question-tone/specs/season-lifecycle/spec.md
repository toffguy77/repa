# Spec Delta

## ADDED Requirements

### Requirement: Season question selection respects the group's tone setting

Selecting a season's questions SHALL honour the group's tone setting as well as its categories, so a
group that asked for kind questions never receives an edgy one.

#### Scenario: Tone and category are both applied

- **GIVEN** a kind-only group with a single category enabled
- **WHEN** a season is created
- **THEN** every selected question is in that category and is not edgy

#### Scenario: Rotation still applies

- **GIVEN** a kind-only group whose recent seasons used certain questions
- **WHEN** a new season is created
- **THEN** those recent questions are still excluded, as for any group
