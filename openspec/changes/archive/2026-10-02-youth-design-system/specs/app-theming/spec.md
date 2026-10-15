# Spec Delta

## Purpose

How a Repa client picks and applies a theme — dark by default, light when the system asks
for it — and the component kit that keeps every screen rendering the same way without each
one restating the design.

## ADDED Requirements

### Requirement: The app ships both themes and follows the system

The app SHALL provide both a dark and a light theme and SHALL follow the operating system's
appearance setting. Dark SHALL be the reference palette: it is the palette the token set is
designed and contrast-checked against first, and the one the shared reputation card matches.

#### Scenario: Dark system setting yields dark

- **GIVEN** the operating system is set to dark appearance
- **WHEN** the app launches
- **THEN** the dark theme is applied

#### Scenario: Light system setting yields light

- **GIVEN** the operating system is set to light appearance
- **WHEN** the app launches
- **THEN** the light theme is applied

#### Scenario: Neither theme is hardcoded

- **WHEN** the app is built
- **THEN** both a light and a dark theme are supplied and the mode is resolved from the platform, not fixed in code

#### Scenario: Switching appearance mid-session

- **GIVEN** the app is running in dark
- **WHEN** the operating system switches to light appearance
- **THEN** the app re-renders in the light theme without a restart and without losing the current screen

### Requirement: Both themes are complete

Every colour role the app uses SHALL be defined in both themes. A screen SHALL NOT be
readable in one theme and unreadable in the other.

#### Scenario: Every screen renders in both themes

- **WHEN** each screen is rendered under the dark theme and again under the light theme
- **THEN** no text, icon, or control is invisible against its background in either

#### Scenario: A missing role is a build-time problem, not a runtime one

- **WHEN** a new colour role is added to one theme only
- **THEN** the omission is caught before the app runs rather than appearing as an invisible control

### Requirement: Shared components carry the design

The app SHALL provide components for its repeated patterns — primary, secondary, and ghost
buttons; card surface; chip; progress bar; bottom sheet; list tile; numeric readout; and
section header — and screens SHALL use them instead of restyling Material widgets inline.

#### Scenario: A button looks the same everywhere

- **WHEN** a primary action is rendered on any two screens
- **THEN** both have the same height, radius, type, and pressed feedback

#### Scenario: A component adapts to the active theme

- **WHEN** a card surface is rendered under each theme
- **THEN** it takes that theme's surface, border, and elevation tokens with no per-screen override

#### Scenario: Components meet the minimum touch target

- **WHEN** any interactive component is rendered
- **THEN** its touch target is at least 44 logical pixels in both dimensions, even when drawn smaller

### Requirement: Density is compact without sacrificing reachability

Layouts SHALL use the compact end of the spacing scale, while interactive targets SHALL
remain at least 44 logical pixels. Visual tightness SHALL NOT be achieved by shrinking what
a user has to hit.

#### Scenario: A dense list is still tappable

- **WHEN** a members list is rendered at compact density
- **THEN** each row's tappable area is at least 44 logical pixels tall

#### Scenario: Small controls keep a large target

- **WHEN** an icon button is drawn at 20 logical pixels
- **THEN** its surrounding touch target is still at least 44

### Requirement: Restyled surfaces keep their behaviour

Restyling a screen SHALL NOT change what it does. Every state a screen could previously
render — loading, empty, error, offline, and each content state — SHALL still be reachable
and still convey the same information.

#### Scenario: Empty and error states survive the restyle

- **WHEN** a restyled screen has no data, or fails to load
- **THEN** it still shows an empty or error state with the same meaning as before

#### Scenario: Existing flows still complete

- **WHEN** a user completes voting on the restyled voting screen
- **THEN** the same votes are cast and the same completion screen follows
