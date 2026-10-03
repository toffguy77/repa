import 'package:flutter/material.dart';

/// The design tokens every Repa surface is built from.
///
/// Tokens are resolved from the ambient theme (`context.t`), never from static
/// constants — that is what makes a second palette possible at all. Names describe the
/// *role* a value plays, not what it looks like, so changing a hue never requires renaming.
///
/// Values and the rules they obey are documented in `docs/features/design-system.md`.
/// Contrast is guaranteed by `test/core/theme/contrast_test.dart`, not by eye.
@immutable
class AppTokens extends ThemeExtension<AppTokens> {
  const AppTokens({
    required this.brightness,
    required this.color,
    required this.elevation,
  });

  final Brightness brightness;
  final AppColorTokens color;
  final AppElevationTokens elevation;

  // Spacing, radius, border, type and motion do not vary by palette, so they are
  // compile-time constants rather than instance fields.
  static const space = AppSpacing();
  static const radius = AppRadius();
  static const border = AppBorder();
  static const text = AppTypeScale();
  static const motion = AppMotion();

  static const AppTokens dark = AppTokens(
    brightness: Brightness.dark,
    color: AppColorTokens.dark,
    elevation: AppElevationTokens.dark,
  );

  static const AppTokens light = AppTokens(
    brightness: Brightness.light,
    color: AppColorTokens.light,
    elevation: AppElevationTokens.light,
  );

  @override
  AppTokens copyWith({
    Brightness? brightness,
    AppColorTokens? color,
    AppElevationTokens? elevation,
  }) {
    return AppTokens(
      brightness: brightness ?? this.brightness,
      color: color ?? this.color,
      elevation: elevation ?? this.elevation,
    );
  }

  @override
  AppTokens lerp(AppTokens? other, double t) {
    if (other == null) return this;
    return AppTokens(
      // Brightness is categorical: it flips at the midpoint rather than blending.
      brightness: t < 0.5 ? brightness : other.brightness,
      color: color.lerp(other.color, t),
      elevation: elevation.lerp(other.elevation, t),
    );
  }
}

/// Role-named colours. Every field is defined in both palettes; a palette missing one is a
/// compile error, not an invisible control at runtime.
@immutable
class AppColorTokens {
  const AppColorTokens({
    required this.canvas,
    required this.surface,
    required this.surfaceRaised,
    required this.border,
    required this.accent,
    required this.accentFill,
    required this.accentFillPressed,
    required this.onAccentFill,
    required this.energy,
    required this.energyFill,
    required this.onEnergyFill,
    required this.textPrimary,
    required this.textSecondary,
    required this.success,
    required this.warning,
    required this.danger,
    required this.scrim,
  });

  /// The page background.
  final Color canvas;

  /// A panel sitting on the canvas.
  final Color surface;

  /// A panel sitting on another panel. In dark mode this is how depth is expressed —
  /// a drop shadow on a near-black canvas is invisible.
  final Color surfaceRaised;

  /// Hairline separator and the resting outline of a panel.
  final Color border;

  /// The brand voice as *foreground*: icons, links, emphasis text on a dark canvas.
  final Color accent;

  /// The brand voice as a filled control background. Separate from [accent] because a
  /// fill must be dark enough for [onAccentFill] text while a foreground accent must be
  /// light enough to read on the canvas — one value cannot satisfy both in dark mode.
  final Color accentFill;
  final Color accentFillPressed;
  final Color onAccentFill;

  /// The second voice — streaks, crystals, "new". Used in small areas only; violet plus
  /// acid lime is what keeps the UI from reading as corporate.
  final Color energy;
  final Color energyFill;
  final Color onEnergyFill;

  final Color textPrimary;
  final Color textSecondary;

  final Color success;
  final Color warning;
  final Color danger;

  /// Overlay behind modals and sheets.
  final Color scrim;

  static const AppColorTokens dark = AppColorTokens(
    canvas: Color(0xFF0B0712),
    surface: Color(0xFF161022),
    surfaceRaised: Color(0xFF1F1733),
    border: Color(0xFF2E2446),
    accent: Color(0xFF9B6DFF),
    accentFill: Color(0xFF7C3AED),
    accentFillPressed: Color(0xFF6228CC),
    onAccentFill: Color(0xFFFFFFFF),
    energy: Color(0xFFC8F751),
    energyFill: Color(0xFFC8F751),
    onEnergyFill: Color(0xFF17210A),
    textPrimary: Color(0xFFF6F3FF),
    textSecondary: Color(0xFFAFA3CC),
    success: Color(0xFF3DDC97),
    warning: Color(0xFFFFC45A),
    danger: Color(0xFFFF7A8A),
    scrim: Color(0xB3000000),
  );

  static const AppColorTokens light = AppColorTokens(
    canvas: Color(0xFFFFFFFF),
    surface: Color(0xFFF6F4FB),
    surfaceRaised: Color(0xFFFFFFFF),
    border: Color(0xFFE6E1F0),
    accent: Color(0xFF7C3AED),
    accentFill: Color(0xFF7C3AED),
    accentFillPressed: Color(0xFF6228CC),
    onAccentFill: Color(0xFFFFFFFF),
    energy: Color(0xFF4D7C0F),
    energyFill: Color(0xFF4D7C0F),
    onEnergyFill: Color(0xFFFFFFFF),
    textPrimary: Color(0xFF150E24),
    textSecondary: Color(0xFF5C5375),
    success: Color(0xFF047857),
    warning: Color(0xFFA16207),
    danger: Color(0xFFDC2626),
    scrim: Color(0x66000000),
  );

  AppColorTokens lerp(AppColorTokens other, double t) {
    Color c(Color a, Color b) => Color.lerp(a, b, t)!;
    return AppColorTokens(
      canvas: c(canvas, other.canvas),
      surface: c(surface, other.surface),
      surfaceRaised: c(surfaceRaised, other.surfaceRaised),
      border: c(border, other.border),
      accent: c(accent, other.accent),
      accentFill: c(accentFill, other.accentFill),
      accentFillPressed: c(accentFillPressed, other.accentFillPressed),
      onAccentFill: c(onAccentFill, other.onAccentFill),
      energy: c(energy, other.energy),
      energyFill: c(energyFill, other.energyFill),
      onEnergyFill: c(onEnergyFill, other.onEnergyFill),
      textPrimary: c(textPrimary, other.textPrimary),
      textSecondary: c(textSecondary, other.textSecondary),
      success: c(success, other.success),
      warning: c(warning, other.warning),
      danger: c(danger, other.danger),
      scrim: c(scrim, other.scrim),
    );
  }

  /// Every field, for the tests that assert completeness and contrast.
  Map<String, Color> get all => {
        'canvas': canvas,
        'surface': surface,
        'surfaceRaised': surfaceRaised,
        'border': border,
        'accent': accent,
        'accentFill': accentFill,
        'accentFillPressed': accentFillPressed,
        'onAccentFill': onAccentFill,
        'energy': energy,
        'energyFill': energyFill,
        'onEnergyFill': onEnergyFill,
        'textPrimary': textPrimary,
        'textSecondary': textSecondary,
        'success': success,
        'warning': warning,
        'danger': danger,
        'scrim': scrim,
      };
}

/// Depth. A level is a *triple* — surface colour, outline colour, shadows — because dark
/// mode expresses depth by stacking surfaces and light mode by casting shadows; a single
/// numeric elevation scale stops working in one of them.
@immutable
class AppElevationTokens {
  const AppElevationTokens({required this.level0, required this.level1, required this.level2});

  /// Flush with the canvas.
  final AppElevation level0;

  /// A card.
  final AppElevation level1;

  /// A sheet, dialog, or menu.
  final AppElevation level2;

  static const AppElevationTokens dark = AppElevationTokens(
    level0: AppElevation(
      surface: Color(0xFF0B0712),
      outline: Color(0xFF2E2446),
      shadows: <BoxShadow>[],
    ),
    level1: AppElevation(
      surface: Color(0xFF161022),
      outline: Color(0xFF2E2446),
      shadows: <BoxShadow>[],
    ),
    level2: AppElevation(
      surface: Color(0xFF1F1733),
      outline: Color(0xFF2E2446),
      shadows: <BoxShadow>[],
    ),
  );

  static const AppElevationTokens light = AppElevationTokens(
    level0: AppElevation(
      surface: Color(0xFFFFFFFF),
      outline: Color(0xFFE6E1F0),
      shadows: <BoxShadow>[],
    ),
    level1: AppElevation(
      surface: Color(0xFFFFFFFF),
      outline: Color(0xFFE6E1F0),
      shadows: <BoxShadow>[
        BoxShadow(
          color: Color(0x0D140E24),
          blurRadius: 12,
          offset: Offset(0, 2),
        ),
      ],
    ),
    level2: AppElevation(
      surface: Color(0xFFFFFFFF),
      outline: Color(0xFFE6E1F0),
      shadows: <BoxShadow>[
        BoxShadow(
          color: Color(0x1A140E24),
          blurRadius: 28,
          offset: Offset(0, 8),
        ),
      ],
    ),
  );

  AppElevationTokens lerp(AppElevationTokens other, double t) => AppElevationTokens(
        level0: level0.lerp(other.level0, t),
        level1: level1.lerp(other.level1, t),
        level2: level2.lerp(other.level2, t),
      );
}

@immutable
class AppElevation {
  const AppElevation({
    required this.surface,
    required this.outline,
    required this.shadows,
  });

  final Color surface;
  final Color outline;
  final List<BoxShadow> shadows;

  AppElevation lerp(AppElevation other, double t) => AppElevation(
        surface: Color.lerp(surface, other.surface, t)!,
        outline: Color.lerp(outline, other.outline, t)!,
        shadows: t < 0.5 ? shadows : other.shadows,
      );
}

/// One geometric scale, 4-pixel base. Deliberately dense at the small end: the product's
/// density comes from spacing, not from shrinking what a user has to hit.
@immutable
class AppSpacing {
  const AppSpacing();

  final double xs = 4;
  final double sm = 8;
  final double md = 12;
  final double lg = 16;
  final double xl = 24;
  final double xxl = 32;
  final double huge = 48;

  List<double> get all => [xs, sm, md, lg, xl, xxl, huge];
}

/// Radii diverge by role rather than being uniform: tight controls against soft cards is
/// what makes the density read as intentional instead of cramped.
@immutable
class AppRadius {
  const AppRadius();

  /// Chip, badge, tag.
  final double xs = 6;

  /// Button, input.
  final double sm = 10;

  /// List tile, inline notice.
  final double md = 14;

  /// Card.
  final double lg = 20;

  /// Bottom sheet, modal.
  final double xl = 28;

  /// Pill.
  final double full = 999;

  BorderRadius get chip => BorderRadius.circular(xs);
  BorderRadius get control => BorderRadius.circular(sm);
  BorderRadius get tile => BorderRadius.circular(md);
  BorderRadius get card => BorderRadius.circular(lg);
  BorderRadius get sheet => BorderRadius.vertical(top: Radius.circular(xl));
}

@immutable
class AppBorder {
  const AppBorder();

  final double hairline = 1;
  final double strong = 1.5;

  /// The smallest a touch target may be, however small the control is drawn.
  final double minTouchTarget = 44;
}

/// Type roles with explicit metrics. [fontFamily] is resolved in exactly one place so a
/// brand face can replace the system font without touching a single screen.
@immutable
class AppTypeScale {
  const AppTypeScale();

  /// null means the platform's system font (per docs/specs/master_context.md).
  String? get fontFamily => null;

  TextStyle get display1 => TextStyle(
        fontFamily: fontFamily,
        fontSize: 40,
        height: 44 / 40,
        fontWeight: FontWeight.w800,
        letterSpacing: -1.2,
      );

  TextStyle get display2 => TextStyle(
        fontFamily: fontFamily,
        fontSize: 32,
        height: 36 / 32,
        fontWeight: FontWeight.w800,
        letterSpacing: -0.8,
      );

  TextStyle get heading1 => TextStyle(
        fontFamily: fontFamily,
        fontSize: 24,
        height: 28 / 24,
        fontWeight: FontWeight.w700,
        letterSpacing: -0.4,
      );

  TextStyle get heading2 => TextStyle(
        fontFamily: fontFamily,
        fontSize: 19,
        height: 24 / 19,
        fontWeight: FontWeight.w700,
        letterSpacing: -0.2,
      );

  TextStyle get body => TextStyle(
        fontFamily: fontFamily,
        fontSize: 15,
        height: 22 / 15,
        fontWeight: FontWeight.w400,
      );

  TextStyle get bodyStrong => TextStyle(
        fontFamily: fontFamily,
        fontSize: 15,
        height: 22 / 15,
        fontWeight: FontWeight.w600,
      );

  TextStyle get caption => TextStyle(
        fontFamily: fontFamily,
        fontSize: 13,
        height: 18 / 13,
        fontWeight: FontWeight.w400,
      );

  /// Uppercase micro-label. Supplies the dense texture the product's density depends on.
  TextStyle get label => TextStyle(
        fontFamily: fontFamily,
        fontSize: 11,
        height: 14 / 11,
        fontWeight: FontWeight.w700,
        letterSpacing: 0.8,
      );

  /// Counters and percentages. Tabular figures so a changing number does not shift layout.
  TextStyle get numeric => TextStyle(
        fontFamily: fontFamily,
        fontSize: 15,
        height: 20 / 15,
        fontWeight: FontWeight.w700,
        fontFeatures: const [FontFeature.tabularFigures()],
      );

  Map<String, TextStyle> get all => {
        'display1': display1,
        'display2': display2,
        'heading1': heading1,
        'heading2': heading2,
        'body': body,
        'bodyStrong': bodyStrong,
        'caption': caption,
        'label': label,
        'numeric': numeric,
      };
}

/// What a piece of motion is *for*. Screens pick a class, never a number, so two
/// comparable interactions can never animate at different speeds.
enum MotionClass {
  /// Press feedback, selection, small state flips.
  tap,

  /// A surface appearing, a route transition, a list reflow.
  surface,

  /// The Reveal. The only place motion is allowed to be the point.
  emphasis,
}

@immutable
class AppMotion {
  const AppMotion();

  /// One step of an entrance stagger. Lists and slide-ups offset each child by a multiple
  /// of this, so a stagger never becomes a per-screen guess.
  Duration stagger(int index) => Duration(milliseconds: 60 * index);

  /// Period of the skeleton shimmer loop. Not an interaction class — it is a continuous
  /// ambient loop, so it has its own named value rather than borrowing one.
  Duration get shimmerPeriod => const Duration(milliseconds: 1200);

  Duration durationOf(MotionClass c) {
    switch (c) {
      case MotionClass.tap:
        return const Duration(milliseconds: 120);
      case MotionClass.surface:
        return const Duration(milliseconds: 220);
      case MotionClass.emphasis:
        return const Duration(milliseconds: 420);
    }
  }

  Curve curveOf(MotionClass c) {
    switch (c) {
      case MotionClass.tap:
        return Curves.easeOut;
      case MotionClass.surface:
        return Curves.easeOutCubic;
      case MotionClass.emphasis:
        return Curves.easeOutBack;
    }
  }
}

/// Type roles already bound to the active palette's text colours.
///
/// Exists so a screen cannot render a correctly-sized heading in the wrong colour: the
/// common case is "this role, in the colour that role implies", and spelling that out at
/// every call site is how the old `AppTextStyles` ended up with colours baked in.
@immutable
class ResolvedText {
  const ResolvedText(this._color);

  final AppColorTokens _color;

  TextStyle get display1 => AppTokens.text.display1.copyWith(color: _color.textPrimary);
  TextStyle get display2 => AppTokens.text.display2.copyWith(color: _color.textPrimary);
  TextStyle get heading1 => AppTokens.text.heading1.copyWith(color: _color.textPrimary);
  TextStyle get heading2 => AppTokens.text.heading2.copyWith(color: _color.textPrimary);
  TextStyle get body => AppTokens.text.body.copyWith(color: _color.textPrimary);
  TextStyle get bodyStrong => AppTokens.text.bodyStrong.copyWith(color: _color.textPrimary);

  /// Body text in the de-emphasised colour.
  TextStyle get bodySecondary => AppTokens.text.body.copyWith(color: _color.textSecondary);

  TextStyle get caption => AppTokens.text.caption.copyWith(color: _color.textSecondary);
  TextStyle get label => AppTokens.text.label.copyWith(color: _color.textSecondary);
  TextStyle get numeric => AppTokens.text.numeric.copyWith(color: _color.textPrimary);

  /// Label on a filled accent control.
  TextStyle get button => AppTokens.text.bodyStrong.copyWith(color: _color.onAccentFill);

  /// The accent voice as text.
  TextStyle get accent => AppTokens.text.bodyStrong.copyWith(color: _color.accent);
}

extension AppTokensContext on BuildContext {
  /// The active tokens. A missing extension is a wiring bug that should fail at the first
  /// build rather than silently fall back to values that look almost right.
  AppTokens get t => Theme.of(this).extension<AppTokens>()!;

  /// Duration for [c], collapsed to zero when the platform asks for reduced motion.
  /// Durations collapse rather than widgets disappearing, so layout cannot shift between
  /// the two modes.
  Duration motion(MotionClass c) => MediaQuery.disableAnimationsOf(this)
      ? Duration.zero
      : AppTokens.motion.durationOf(c);

  Curve motionCurve(MotionClass c) => AppTokens.motion.curveOf(c);

  /// Type roles pre-bound to the active palette. See [ResolvedText].
  ResolvedText get ts => ResolvedText(t.color);
}
