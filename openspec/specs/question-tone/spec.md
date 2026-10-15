# question-tone Specification

## Purpose
What tone a question has, how a group constrains which tones it receives, and how tone decides what a
reputation card leads with.

## Requirements

### Requirement: Every question has a tone

Each question SHALL carry a tone of warm, neutral, or edgy. A question SHALL NOT exist without one.

#### Scenario: Seeded questions are classified

- **WHEN** the question bank is seeded
- **THEN** every question has a tone

#### Scenario: A user-created question gets a tone

- **WHEN** a member submits a custom question
- **THEN** the stored question has a tone

#### Scenario: An unclassifiable question is neutral, not rejected

- **GIVEN** a question whose tone cannot be determined
- **WHEN** it is stored
- **THEN** its tone is neutral and the question is still usable

### Requirement: A group can restrict itself to kind questions

A group SHALL have a setting that restricts its seasons to warm and neutral questions. While it is on,
an edgy question SHALL NOT be selected for that group's seasons.

#### Scenario: A kind-only group receives no edgy questions

- **GIVEN** a group with the kind-only setting on
- **WHEN** a season is created for it
- **THEN** none of the season's questions are edgy

#### Scenario: A group without the setting receives the full bank

- **GIVEN** a group with the kind-only setting off
- **WHEN** a season is created for it
- **THEN** questions of any tone may be selected

#### Scenario: The setting can be changed by the admin

- **GIVEN** a group with the setting off
- **WHEN** the admin turns it on
- **THEN** subsequent seasons exclude edgy questions

#### Scenario: Changing the setting does not rewrite an open season

- **GIVEN** an open season containing an edgy question
- **WHEN** the admin turns the setting on
- **THEN** the open season is unchanged and members keep answering what they were asked

#### Scenario: A kind-only group still gets a full season

- **GIVEN** a kind-only group
- **WHEN** a season is created
- **THEN** it has the normal number of questions, drawn from the warm and neutral ones

### Requirement: School-age groups are kind-only by default

When a group is created by someone under 18, the kind-only setting SHALL default to on. The creator
SHALL still be able to turn it off.

#### Scenario: An under-18 creator gets the setting on

- **WHEN** a user under 18 creates a group without specifying the setting
- **THEN** the group is kind-only

#### Scenario: An adult creator gets the setting off

- **WHEN** a user over 18 creates a group without specifying the setting
- **THEN** the group is not kind-only

#### Scenario: An explicit choice wins over the default

- **WHEN** a creator explicitly sets the value
- **THEN** that value is used regardless of their age

### Requirement: A kind-only group cannot be left without questions

Some categories exist in order to provoke, so every question in them is edgy. A combination of the
kind-only setting and categories that would leave the group with nothing to ask SHALL be refused,
rather than producing an empty season. The refusal SHALL name the setting and point at the categories,
so the member can see which of the two to change.

#### Scenario: Creating a kind-only group whose categories are all edgy is refused

- **GIVEN** a user creating a group with the kind-only setting on
- **WHEN** every category they chose offers only edgy questions
- **THEN** creation is refused with an explanation naming the setting and pointing at the categories

#### Scenario: Turning the setting on is refused when it would empty the bank

- **GIVEN** a group whose categories offer only edgy questions
- **WHEN** the admin turns the kind-only setting on
- **THEN** the change is refused and the setting stays off

#### Scenario: One kind category is enough

- **GIVEN** a user creating a kind-only group
- **WHEN** they choose one category that offers kind questions alongside categories that do not
- **THEN** creation succeeds and the season draws from the kind category

#### Scenario: An ordinary group is unaffected

- **GIVEN** a user creating a group with the kind-only setting off
- **WHEN** every category they chose offers only edgy questions
- **THEN** creation succeeds

### Requirement: A card leads with something worth sharing

When two of a member's attributes have comparable vote counts, the card SHALL lead with the warmer one,
so the artifact a person is invited to share is one they would want to.

#### Scenario: A warm attribute outranks an edgy one at the same count

- **GIVEN** a member with a warm attribute and an edgy attribute at equal vote counts
- **WHEN** their card is built
- **THEN** the warm attribute is first

#### Scenario: A clearly stronger result still wins

- **GIVEN** a member whose edgy attribute has substantially more votes than any warm one
- **WHEN** their card is built
- **THEN** the edgy attribute is first, because the card must still be truthful

#### Scenario: Tone does not remove attributes

- **WHEN** a card is built
- **THEN** every attribute the member received is present, in tone-aware order

#### Scenario: Every view of a card uses the same order

- **GIVEN** a member whose attributes include a near-tie between a warm and an edgy one
- **WHEN** their card is shown in the app, listed among the group's cards, expanded by the
  hidden-attributes purchase, and rendered as the shared image
- **THEN** all four lead with the same attribute

### Requirement: The setting states what it does

The app SHALL describe the kind-only setting in terms of what the group will and will not be asked,
rather than as an unexplained toggle.

#### Scenario: The setting is explained where it is chosen

- **WHEN** a user creates a group
- **THEN** the setting's effect on the questions is stated next to it
