# Spec Delta

## ADDED Requirements

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
