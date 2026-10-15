import 'package:flutter/material.dart';

import 'app_tokens.dart';

/// Builds Material's [ThemeData] from [AppTokens].
///
/// Every widget theme configured here exists so a screen that was never restyled by hand
/// still picks up the design: that is what makes "some screens bespoke, the rest inherited"
/// a coherent result rather than a half-finished one.
class AppTheme {
  const AppTheme._();

  static ThemeData get dark => _build(AppTokens.dark);
  static ThemeData get light => _build(AppTokens.light);

  static ThemeData _build(AppTokens t) {
    final c = t.color;
    final text = AppTokens.text;
    final radius = AppTokens.radius;
    final space = AppTokens.space;
    final border = AppTokens.border;

    final scheme = ColorScheme(
      brightness: t.brightness,
      primary: c.accentFill,
      onPrimary: c.onAccentFill,
      secondary: c.energyFill,
      onSecondary: c.onEnergyFill,
      error: c.danger,
      onError: c.onAccentFill,
      surface: c.surface,
      onSurface: c.textPrimary,
      outline: c.border,
      scrim: c.scrim,
    );

    TextStyle on(TextStyle style, Color color) => style.copyWith(color: color);

    return ThemeData(
      useMaterial3: true,
      brightness: t.brightness,
      colorScheme: scheme,
      extensions: [t],
      scaffoldBackgroundColor: c.canvas,
      canvasColor: c.canvas,
      dividerColor: c.border,
      splashFactory: InkSparkle.splashFactory,

      textTheme: TextTheme(
        displayLarge: on(text.display1, c.textPrimary),
        displayMedium: on(text.display2, c.textPrimary),
        headlineMedium: on(text.heading1, c.textPrimary),
        headlineSmall: on(text.heading2, c.textPrimary),
        bodyLarge: on(text.body, c.textPrimary),
        bodyMedium: on(text.body, c.textPrimary),
        bodySmall: on(text.caption, c.textSecondary),
        labelLarge: on(text.bodyStrong, c.textPrimary),
        labelSmall: on(text.label, c.textSecondary),
      ),

      appBarTheme: AppBarTheme(
        backgroundColor: c.canvas,
        foregroundColor: c.textPrimary,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
        titleTextStyle: on(text.heading2, c.textPrimary),
        iconTheme: IconThemeData(color: c.textPrimary, size: 22),
      ),

      iconTheme: IconThemeData(color: c.textSecondary, size: 20),

      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ButtonStyle(
          backgroundColor: WidgetStateProperty.resolveWith((states) {
            if (states.contains(WidgetState.disabled)) {
              return c.accentFill.withValues(alpha: 0.35);
            }
            if (states.contains(WidgetState.pressed)) return c.accentFillPressed;
            return c.accentFill;
          }),
          foregroundColor: WidgetStateProperty.resolveWith((states) =>
              states.contains(WidgetState.disabled)
                  ? c.onAccentFill.withValues(alpha: 0.6)
                  : c.onAccentFill),
          minimumSize: WidgetStatePropertyAll(
            Size(double.infinity, border.minTouchTarget + 4),
          ),
          shape: WidgetStatePropertyAll(
            RoundedRectangleBorder(borderRadius: radius.control),
          ),
          textStyle: WidgetStatePropertyAll(text.bodyStrong),
          elevation: const WidgetStatePropertyAll(0),
        ),
      ),

      outlinedButtonTheme: OutlinedButtonThemeData(
        style: ButtonStyle(
          foregroundColor: WidgetStatePropertyAll(c.textPrimary),
          minimumSize: WidgetStatePropertyAll(
            Size(double.infinity, border.minTouchTarget + 4),
          ),
          side: WidgetStatePropertyAll(
            BorderSide(color: c.border, width: border.hairline),
          ),
          shape: WidgetStatePropertyAll(
            RoundedRectangleBorder(borderRadius: radius.control),
          ),
          textStyle: WidgetStatePropertyAll(text.bodyStrong),
        ),
      ),

      textButtonTheme: TextButtonThemeData(
        style: ButtonStyle(
          foregroundColor: WidgetStatePropertyAll(c.accent),
          minimumSize: WidgetStatePropertyAll(
            Size(0, border.minTouchTarget),
          ),
          textStyle: WidgetStatePropertyAll(text.bodyStrong),
        ),
      ),

      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: c.surface,
        hintStyle: on(text.body, c.textSecondary),
        labelStyle: on(text.caption, c.textSecondary),
        contentPadding: EdgeInsets.symmetric(
          horizontal: space.lg,
          vertical: space.md,
        ),
        border: OutlineInputBorder(
          borderRadius: radius.control,
          borderSide: BorderSide(color: c.border, width: border.hairline),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: radius.control,
          borderSide: BorderSide(color: c.border, width: border.hairline),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: radius.control,
          borderSide: BorderSide(color: c.accent, width: border.strong),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: radius.control,
          borderSide: BorderSide(color: c.danger, width: border.hairline),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: radius.control,
          borderSide: BorderSide(color: c.danger, width: border.strong),
        ),
      ),

      cardTheme: CardThemeData(
        color: t.elevation.level1.surface,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: radius.card,
          side: BorderSide(color: t.elevation.level1.outline, width: border.hairline),
        ),
      ),

      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: t.elevation.level2.surface,
        surfaceTintColor: Colors.transparent,
        modalBarrierColor: c.scrim,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: radius.sheet),
      ),

      dialogTheme: DialogThemeData(
        backgroundColor: t.elevation.level2.surface,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: radius.card),
        titleTextStyle: on(text.heading2, c.textPrimary),
        contentTextStyle: on(text.body, c.textSecondary),
      ),

      chipTheme: ChipThemeData(
        backgroundColor: c.surface,
        selectedColor: c.accentFill,
        disabledColor: c.surface,
        labelStyle: on(text.caption, c.textPrimary),
        secondaryLabelStyle: on(text.caption, c.onAccentFill),
        side: BorderSide(color: c.border, width: border.hairline),
        shape: RoundedRectangleBorder(borderRadius: radius.chip),
        padding: EdgeInsets.symmetric(horizontal: space.md, vertical: space.sm),
      ),

      progressIndicatorTheme: ProgressIndicatorThemeData(
        color: c.accent,
        linearTrackColor: c.surface,
        circularTrackColor: c.surface,
        linearMinHeight: 6,
      ),

      bottomNavigationBarTheme: BottomNavigationBarThemeData(
        backgroundColor: c.canvas,
        selectedItemColor: c.accent,
        unselectedItemColor: c.textSecondary,
        selectedLabelStyle: text.label,
        unselectedLabelStyle: text.label,
        type: BottomNavigationBarType.fixed,
        elevation: 0,
      ),

      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: c.canvas,
        indicatorColor: c.accentFill.withValues(alpha: 0.18),
        surfaceTintColor: Colors.transparent,
        elevation: 0,
      ),

      snackBarTheme: SnackBarThemeData(
        backgroundColor: t.elevation.level2.surface,
        contentTextStyle: on(text.body, c.textPrimary),
        actionTextColor: c.accent,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: radius.tile),
        elevation: 0,
      ),

      listTileTheme: ListTileThemeData(
        iconColor: c.textSecondary,
        textColor: c.textPrimary,
        titleTextStyle: on(text.bodyStrong, c.textPrimary),
        subtitleTextStyle: on(text.caption, c.textSecondary),
        minVerticalPadding: space.md,
        shape: RoundedRectangleBorder(borderRadius: radius.tile),
      ),

      dividerTheme: DividerThemeData(
        color: c.border,
        thickness: border.hairline,
        space: 1,
      ),

      tabBarTheme: TabBarThemeData(
        labelColor: c.textPrimary,
        unselectedLabelColor: c.textSecondary,
        labelStyle: text.bodyStrong,
        unselectedLabelStyle: text.body,
        indicatorColor: c.accent,
        dividerColor: c.border,
      ),

      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: c.accentFill,
        foregroundColor: c.onAccentFill,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: radius.tile),
      ),

      pageTransitionsTheme: const PageTransitionsTheme(
        builders: {
          TargetPlatform.android: FadeForwardsPageTransitionsBuilder(),
          TargetPlatform.iOS: FadeForwardsPageTransitionsBuilder(),
        },
      ),
    );
  }
}
