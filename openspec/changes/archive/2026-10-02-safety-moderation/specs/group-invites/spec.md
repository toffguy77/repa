# Spec Delta

## ADDED Requirements

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
