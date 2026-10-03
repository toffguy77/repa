# Spec Delta

## ADDED Requirements

### Requirement: A push only promises what the app can show

A notification that refers to information about the recipient SHALL only be sent when the screen it
opens will actually contain that information.

#### Scenario: The "someone answered about you" push requires a non-zero count

- **GIVEN** a member about whom nobody has voted
- **WHEN** the mid-week signal push runs
- **THEN** that member is not sent the push

#### Scenario: The teaser push requires a teaser

- **GIVEN** a member with no votes about them
- **WHEN** the teaser push runs
- **THEN** that member is not sent the push

#### Scenario: A member with something to see is still notified

- **GIVEN** a member about whom two people have voted
- **WHEN** the mid-week signal push runs
- **THEN** they are sent the push

#### Scenario: The push opens the screen that holds the information

- **WHEN** a member taps the mid-week signal push
- **THEN** the screen that opens shows how many people have answered about them
