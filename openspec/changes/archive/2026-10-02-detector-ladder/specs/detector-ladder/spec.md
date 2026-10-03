# Spec Delta

## Purpose

The rungs of the detector — what a member can learn about who voted for them at each price — and
the anonymity limits that hold at every rung.

## ADDED Requirements

### Requirement: The voter count is free

A member SHALL see how many people voted about them in a revealed season without paying anything.
The count SHALL NOT identify anyone.

#### Scenario: The count is visible without a purchase

- **GIVEN** a member of a group whose season has revealed
- **WHEN** they open the detector without buying anything
- **THEN** they see how many people voted about them
- **AND** no names, avatars, or initials are shown

#### Scenario: The count does not reveal who abstained

- **WHEN** a member reads the free count
- **THEN** they learn how many voted, not which members did not

### Requirement: A hint reveals one voter partially

For a price below the full list, a member SHALL be able to reveal one voter partially: enough to
guess, not enough to know. A hint SHALL show that voter's avatar and the first character of their
name, and SHALL NOT show the full name.

#### Scenario: Buying a hint reveals a partial identity

- **GIVEN** a member with enough crystals and a revealed season with voters
- **WHEN** they buy a hint
- **THEN** they receive one voter's avatar and the first character of that voter's name
- **AND** the full name is not returned

#### Scenario: A hint costs less than the full list

- **WHEN** the prices of the hint and the full list are compared
- **THEN** the hint costs less

#### Scenario: Hints do not repeat

- **GIVEN** a member who has already revealed one voter with a hint
- **WHEN** they buy another hint
- **THEN** a different voter is revealed

#### Scenario: Hints run out when every voter is revealed

- **GIVEN** a member who has revealed every voter with hints
- **WHEN** they attempt another hint
- **THEN** the request is refused and no crystals are spent

#### Scenario: A hint with no voters is refused

- **GIVEN** a revealed season in which nobody voted about the member
- **WHEN** they attempt a hint
- **THEN** the request is refused and no crystals are spent

#### Scenario: Insufficient balance refuses the hint without spending

- **GIVEN** a member whose balance is below the hint price
- **WHEN** they attempt a hint
- **THEN** the request fails, their balance is unchanged, and no voter is revealed

### Requirement: The full list remains the top rung

Buying the full voter list SHALL still return every voter, at the existing price, and SHALL remain
purchasable whether or not hints were bought first.

#### Scenario: The full list returns every voter

- **GIVEN** a member who has bought the full list
- **WHEN** they read the detector
- **THEN** every voter is listed

#### Scenario: Hints bought first do not block the full list

- **GIVEN** a member who revealed two voters with hints
- **WHEN** they buy the full list
- **THEN** the purchase succeeds and every voter is listed

#### Scenario: The full list is bought once per season

- **GIVEN** a member who already bought the full list for a season
- **WHEN** they attempt to buy it again
- **THEN** the request is refused and no crystals are spent

### Requirement: The ladder's state is visible

The detector response SHALL tell the client which rungs have been bought, what each remaining rung
costs, and whether a further hint is available, so the app can show the ladder rather than a single
button.

#### Scenario: An untouched ladder reports its prices

- **GIVEN** a member who has bought nothing
- **WHEN** they read the detector
- **THEN** the response includes the free voter count, the hint price, the full-list price, and that
  a hint is available

#### Scenario: A partially climbed ladder reports what is left

- **GIVEN** a member who has revealed one of three voters with a hint
- **WHEN** they read the detector
- **THEN** the response includes the revealed hint and that further hints are available

#### Scenario: A completed ladder reports nothing left to buy

- **GIVEN** a member who has bought the full list
- **WHEN** they read the detector
- **THEN** the response indicates no further purchase is available

### Requirement: No rung ever binds a voter to an answer

No rung of the ladder SHALL reveal what any voter answered, or connect a voter to a question. This
holds for the free count, every hint, and the full list.

#### Scenario: A hint carries no answer

- **WHEN** a hint is returned
- **THEN** it contains no question, attribute, or answer

#### Scenario: The full list carries no answers

- **WHEN** the full voter list is returned
- **THEN** it contains no question, attribute, or answer
