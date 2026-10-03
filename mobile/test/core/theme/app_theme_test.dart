import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/core/theme/app_tokens.dart';

/// Pumps a trivial app under [brightness] and returns the tokens the tree resolves.
Future<AppTokens> tokensUnder(WidgetTester tester, Brightness brightness) async {
  late AppTokens resolved;
  tester.platformDispatcher.platformBrightnessTestValue = brightness;
  addTearDown(tester.platformDispatcher.clearPlatformBrightnessTestValue);

  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.light,
    darkTheme: AppTheme.dark,
    themeMode: ThemeMode.system,
    home: Builder(builder: (context) {
      resolved = context.t;
      return const Scaffold(body: Text('hi'));
    }),
  ));
  await tester.pumpAndSettle();
  return resolved;
}

void main() {
  group('AppTheme', () {
    test('registers the tokens on both themes', () {
      expect(AppTheme.dark.extension<AppTokens>(), AppTokens.dark);
      expect(AppTheme.light.extension<AppTokens>(), AppTokens.light);
    });

    test('scaffold background is the canvas token', () {
      expect(AppTheme.dark.scaffoldBackgroundColor, AppColorTokens.dark.canvas);
      expect(AppTheme.light.scaffoldBackgroundColor, AppColorTokens.light.canvas);
    });

    test('brightness matches the palette', () {
      expect(AppTheme.dark.brightness, Brightness.dark);
      expect(AppTheme.light.brightness, Brightness.light);
    });

    test('the colour scheme is derived from tokens, not from a seed', () {
      expect(AppTheme.dark.colorScheme.primary, AppColorTokens.dark.accentFill);
      expect(AppTheme.dark.colorScheme.onPrimary, AppColorTokens.dark.onAccentFill);
      expect(AppTheme.dark.colorScheme.error, AppColorTokens.dark.danger);
      expect(AppTheme.light.colorScheme.primary, AppColorTokens.light.accentFill);
    });

    test('text theme carries the type scale metrics', () {
      final body = AppTheme.dark.textTheme.bodyMedium!;
      expect(body.fontSize, AppTokens.text.body.fontSize);
      expect(body.color, AppColorTokens.dark.textPrimary);

      final display = AppTheme.dark.textTheme.displayLarge!;
      expect(display.fontSize, AppTokens.text.display1.fontSize);
    });

    test('buttons are at least the minimum touch target tall', () {
      final size = AppTheme.dark.elevatedButtonTheme.style!.minimumSize!
          .resolve(<WidgetState>{})!;
      expect(size.height, greaterThanOrEqualTo(AppTokens.border.minTouchTarget));
    });

    test('surfaces are flat: no Material tint or elevation leaks through', () {
      expect(AppTheme.dark.appBarTheme.elevation, 0);
      expect(AppTheme.dark.appBarTheme.surfaceTintColor, Colors.transparent);
      expect(AppTheme.dark.cardTheme.elevation, 0);
      expect(AppTheme.dark.bottomSheetTheme.elevation, 0);
    });

    test('card and sheet take their radius from the scale', () {
      final card = AppTheme.dark.cardTheme.shape! as RoundedRectangleBorder;
      expect(card.borderRadius, AppTokens.radius.card);

      final sheet = AppTheme.dark.bottomSheetTheme.shape! as RoundedRectangleBorder;
      expect(sheet.borderRadius, AppTokens.radius.sheet);
    });

    test('the dark card sits on the level-1 surface with a hairline outline', () {
      expect(AppTheme.dark.cardTheme.color, AppElevationTokens.dark.level1.surface);
      final shape = AppTheme.dark.cardTheme.shape! as RoundedRectangleBorder;
      expect(shape.side.color, AppElevationTokens.dark.level1.outline);
    });
  });

  group('theme selection follows the platform', () {
    testWidgets('dark platform brightness yields the dark palette', (tester) async {
      final tokens = await tokensUnder(tester, Brightness.dark);

      expect(tokens.brightness, Brightness.dark);
      expect(tokens.color.canvas, AppColorTokens.dark.canvas);
    });

    testWidgets('light platform brightness yields the light palette', (tester) async {
      final tokens = await tokensUnder(tester, Brightness.light);

      expect(tokens.brightness, Brightness.light);
      expect(tokens.color.canvas, AppColorTokens.light.canvas);
    });

    testWidgets('switching appearance mid-session re-themes without recreating the screen',
        (tester) async {
      tester.platformDispatcher.platformBrightnessTestValue = Brightness.dark;
      addTearDown(tester.platformDispatcher.clearPlatformBrightnessTestValue);

      var buildCount = 0;
      final key = GlobalKey();

      await tester.pumpWidget(MaterialApp(
        theme: AppTheme.light,
        darkTheme: AppTheme.dark,
        themeMode: ThemeMode.system,
        home: Builder(
          key: key,
          builder: (context) {
            buildCount++;
            return Scaffold(
              body: ColoredBox(
                key: const Key('canvas-probe'),
                color: context.t.color.canvas,
              ),
            );
          },
        ),
      ));
      await tester.pumpAndSettle();

      final firstState = key.currentContext;
      expect(buildCount, greaterThan(0));

      tester.platformDispatcher.platformBrightnessTestValue = Brightness.light;
      await tester.pumpAndSettle();

      final box = tester.widget<ColoredBox>(find.byKey(const Key('canvas-probe')));
      expect(box.color, AppColorTokens.light.canvas,
          reason: 'the new palette should apply without a restart');
      expect(key.currentContext, same(firstState),
          reason: 'the screen element should be updated, not recreated');
    });
  });
}
