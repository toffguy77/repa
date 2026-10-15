import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_tokens.dart';

void main() {
  group('AppTokens.lerp', () {
    test('returns the target at t = 1', () {
      final result = AppTokens.dark.lerp(AppTokens.light, 1);

      expect(result.brightness, Brightness.light);
      expect(result.color.canvas, AppColorTokens.light.canvas);
      expect(result.color.textPrimary, AppColorTokens.light.textPrimary);
    });

    test('returns the source at t = 0', () {
      final result = AppTokens.light.lerp(AppTokens.dark, 0);

      expect(result.brightness, Brightness.light);
      expect(result.color.canvas, AppColorTokens.light.canvas);
    });

    test('blends colours in between', () {
      final result = AppTokens.dark.lerp(AppTokens.light, 0.5);

      expect(result.color.canvas, isNot(AppColorTokens.dark.canvas));
      expect(result.color.canvas, isNot(AppColorTokens.light.canvas));
    });

    test('brightness flips at the midpoint rather than blending', () {
      expect(AppTokens.dark.lerp(AppTokens.light, 0.49).brightness, Brightness.dark);
      expect(AppTokens.dark.lerp(AppTokens.light, 0.5).brightness, Brightness.light);
    });

    test('a null target leaves the tokens unchanged', () {
      expect(AppTokens.dark.lerp(null, 0.5), AppTokens.dark);
    });
  });

  group('palette completeness', () {
    test('both palettes define exactly the same roles', () {
      expect(
        AppColorTokens.dark.all.keys.toSet(),
        AppColorTokens.light.all.keys.toSet(),
      );
    });

    test('the two palettes are distinct instances with different values', () {
      expect(AppColorTokens.dark.canvas, isNot(AppColorTokens.light.canvas));
      expect(AppColorTokens.dark.textPrimary, isNot(AppColorTokens.light.textPrimary));
    });

    test('dark elevations carry no shadows, because a shadow on near-black is invisible',
        () {
      for (final level in [
        AppElevationTokens.dark.level0,
        AppElevationTokens.dark.level1,
        AppElevationTokens.dark.level2,
      ]) {
        expect(level.shadows, isEmpty);
      }
    });

    test('light elevations above flush do cast shadows', () {
      expect(AppElevationTokens.light.level0.shadows, isEmpty);
      expect(AppElevationTokens.light.level1.shadows, isNotEmpty);
      expect(AppElevationTokens.light.level2.shadows, isNotEmpty);
    });

    test('dark depth comes from stacked surfaces', () {
      expect(
        AppElevationTokens.dark.level1.surface,
        isNot(AppElevationTokens.dark.level0.surface),
      );
      expect(
        AppElevationTokens.dark.level2.surface,
        isNot(AppElevationTokens.dark.level1.surface),
      );
    });
  });

  group('spacing scale', () {
    test('every step is a multiple of the 4px base', () {
      for (final step in AppTokens.space.all) {
        expect(step % 4, 0, reason: '$step is not on the 4px grid');
      }
    });

    test('steps increase monotonically', () {
      final steps = AppTokens.space.all;
      for (var i = 1; i < steps.length; i++) {
        expect(steps[i], greaterThan(steps[i - 1]));
      }
    });

    test('the small end is dense', () {
      expect(AppTokens.space.xs, 4);
      expect(AppTokens.space.sm, 8);
      expect(AppTokens.space.md, 12);
    });
  });

  group('radius scale', () {
    test('controls are tighter than cards, cards tighter than sheets', () {
      expect(AppTokens.radius.sm, lessThan(AppTokens.radius.lg));
      expect(AppTokens.radius.lg, lessThan(AppTokens.radius.xl));
    });

    test('sheet radius is top-only', () {
      final sheet = AppTokens.radius.sheet;
      expect(sheet.topLeft.x, AppTokens.radius.xl);
      expect(sheet.bottomLeft.x, 0);
    });
  });

  group('type scale', () {
    test('every role declares size, weight and height', () {
      AppTokens.text.all.forEach((name, style) {
        expect(style.fontSize, isNotNull, reason: '$name has no size');
        expect(style.fontWeight, isNotNull, reason: '$name has no weight');
        expect(style.height, isNotNull, reason: '$name has no line height');
      });
    });

    test('display roles are larger than headings, headings larger than body', () {
      expect(AppTokens.text.display1.fontSize!,
          greaterThan(AppTokens.text.heading1.fontSize!));
      expect(AppTokens.text.heading1.fontSize!,
          greaterThan(AppTokens.text.body.fontSize!));
      expect(AppTokens.text.body.fontSize!,
          greaterThan(AppTokens.text.caption.fontSize!));
    });

    test('the scale is compact: body is 15 and caption 13', () {
      expect(AppTokens.text.body.fontSize, 15);
      expect(AppTokens.text.caption.fontSize, 13);
    });

    test('display roles use tight negative tracking', () {
      expect(AppTokens.text.display1.letterSpacing!, lessThan(0));
      expect(AppTokens.text.display2.letterSpacing!, lessThan(0));
    });

    test('the micro-label is loose and heavy', () {
      expect(AppTokens.text.label.letterSpacing!, greaterThan(0));
      expect(AppTokens.text.label.fontWeight, FontWeight.w700);
    });

    test('numeric uses tabular figures so counters do not shift layout', () {
      expect(
        AppTokens.text.numeric.fontFeatures!.map((f) => f.feature),
        contains('tnum'),
      );
    });

    test('the font family is resolved in one place for every role', () {
      final family = AppTokens.text.fontFamily;
      AppTokens.text.all.forEach((name, style) {
        expect(style.fontFamily, family,
            reason: '$name does not take the scale\'s family');
      });
    });
  });

  group('motion', () {
    test('tap is faster than surface, surface faster than emphasis', () {
      final tap = AppTokens.motion.durationOf(MotionClass.tap);
      final surface = AppTokens.motion.durationOf(MotionClass.surface);
      final emphasis = AppTokens.motion.durationOf(MotionClass.emphasis);

      expect(tap, lessThan(surface));
      expect(surface, lessThan(emphasis));
    });

    test('every class has a curve', () {
      for (final c in MotionClass.values) {
        expect(AppTokens.motion.curveOf(c), isNotNull);
      }
    });
  });

  group('context helpers', () {
    testWidgets('context.t resolves the registered tokens', (tester) async {
      late AppTokens resolved;
      await tester.pumpWidget(MaterialApp(
        theme: ThemeData(extensions: const [AppTokens.dark]),
        home: Builder(builder: (context) {
          resolved = context.t;
          return const SizedBox();
        }),
      ));

      expect(resolved.brightness, Brightness.dark);
      expect(resolved.color.canvas, AppColorTokens.dark.canvas);
    });

    testWidgets('motion returns the class duration normally', (tester) async {
      late Duration d;
      await tester.pumpWidget(MaterialApp(
        theme: ThemeData(extensions: const [AppTokens.dark]),
        home: Builder(builder: (context) {
          d = context.motion(MotionClass.surface);
          return const SizedBox();
        }),
      ));

      expect(d, const Duration(milliseconds: 220));
    });

    testWidgets('motion collapses to zero under reduced motion', (tester) async {
      late Duration tap;
      late Duration emphasis;
      await tester.pumpWidget(MaterialApp(
        theme: ThemeData(extensions: const [AppTokens.dark]),
        home: MediaQuery(
          data: const MediaQueryData(disableAnimations: true),
          child: Builder(builder: (context) {
            tap = context.motion(MotionClass.tap);
            emphasis = context.motion(MotionClass.emphasis);
            return const SizedBox();
          }),
        ),
      ));

      expect(tap, Duration.zero);
      expect(emphasis, Duration.zero);
    });
  });
}
