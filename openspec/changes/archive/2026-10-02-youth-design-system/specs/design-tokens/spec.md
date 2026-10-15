# Spec Delta

## Purpose

The named set of visual and motion primitives every Repa screen is built from — colour,
type, spacing, radius, elevation, and motion — together with the guarantees they carry, so
a surface never invents its own proportions or contrast.

## ADDED Requirements

### Requirement: Every visual value comes from a named token

Screens and widgets SHALL obtain colour, type, spacing, radius, and motion values from the
token layer. No screen SHALL hardcode a colour literal, a radius literal, or an animation
duration.

#### Scenario: A surface uses a token instead of a literal

- **WHEN** a developer needs a card's corner radius
- **THEN** a named radius token is available for it
- **AND** no colour, radius, or duration literal appears in the screen's own code

#### Scenario: Tokens are reachable without passing them down

- **WHEN** a widget deep in the tree needs the surface colour
- **THEN** it can resolve the token from the ambient theme without receiving it as a parameter

### Requirement: Tokens are named by role, not by value

Token names SHALL describe what a value is for, not what it looks like, so a palette change
does not require renaming. A token named after its appearance is a defect.

#### Scenario: Role-named colour survives a palette change

- **WHEN** the accent hue changes
- **THEN** the token that carries it keeps its name
- **AND** every surface using it updates without edits

#### Scenario: Semantic colours are distinguishable from brand colours

- **WHEN** a developer needs the colour for a destructive action
- **THEN** a semantic token exists for that meaning, separate from the brand accent

### Requirement: Colour tokens meet WCAG AA contrast

Every text colour token SHALL reach a contrast ratio of at least 4.5:1 against the surface
tokens it is specified for, and at least 3:1 for large text and for the boundary of
interactive controls. This SHALL hold in both the dark and the light palette.

#### Scenario: Body text on every surface is readable

- **WHEN** each text token is measured against each surface token it is paired with
- **THEN** the ratio is at least 4.5:1 in both palettes

#### Scenario: Secondary text is still readable

- **WHEN** the de-emphasised text token is measured against its surfaces
- **THEN** the ratio is at least 4.5:1, not merely visually "dimmer"

#### Scenario: Accent is usable as a control background

- **WHEN** the on-accent text token is measured against the accent token
- **THEN** the ratio is at least 4.5:1

### Requirement: A single compact spacing scale

Spacing SHALL come from one geometric scale with a 4-logical-pixel base. The product's
visual density is deliberately tight: the scale SHALL provide dense steps for the small
sizes rather than jumping from one generous value to the next.

#### Scenario: Dense steps exist

- **WHEN** a developer needs the gap between a label and its value
- **THEN** a token of 4 or 8 logical pixels is available

#### Scenario: Arbitrary spacing is not representable

- **WHEN** a layout calls for a 13-pixel gap
- **THEN** no token provides it, and the nearest scale step is used instead

### Requirement: A type scale with display, heading, body, and label roles

The type scale SHALL define display, heading, body, and label roles with explicit size,
weight, line height, and letter spacing, and SHALL be defined independently of the font
family so a brand face can replace the system font without touching screens.

#### Scenario: Numeric readouts have their own role

- **WHEN** a reputation percentage is rendered
- **THEN** a display role exists for large numerals with tight letter spacing

#### Scenario: Swapping the font family does not change screens

- **WHEN** the font family is changed in one place
- **THEN** every text role picks it up and no screen code changes

### Requirement: Motion is standardised per interaction class

Motion SHALL be defined as duration and curve tokens grouped by interaction class — tap
feedback, surface transition, and emphasis — so two comparable interactions never animate at
different speeds.

#### Scenario: Comparable interactions share timing

- **WHEN** two different screens animate a surface appearing
- **THEN** both use the same surface-transition duration and curve

#### Scenario: Emphasis motion is distinct from ordinary motion

- **WHEN** the Reveal opens its card
- **THEN** it uses the emphasis tokens, which are slower and more pronounced than tap feedback

#### Scenario: Reduced-motion preference is respected

- **GIVEN** the operating system reports that the user prefers reduced motion
- **WHEN** a screen plays a decorative animation
- **THEN** the animation is skipped or reduced to a fade, and no information is conveyed only by movement

### Requirement: The shared reputation card uses the same palette

The server-rendered reputation card SHALL use the same colour values as the app's dark
palette, so the artifact a user shares and the app they came from read as one product.

#### Scenario: Card background matches the app's dark surface

- **WHEN** the card image is generated
- **THEN** its background and accent values are the dark-palette tokens, not an independent set

#### Scenario: A palette change reaches the card

- **WHEN** the dark palette's accent changes
- **THEN** the card's documented source of those values changes with it
