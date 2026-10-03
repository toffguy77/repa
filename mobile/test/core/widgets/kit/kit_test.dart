import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/core/theme/app_tokens.dart';
import 'package:repa/core/widgets/kit/kit.dart';

/// Pumps [child] under one of the real app themes.
Future<void> pumpKit(
  WidgetTester tester,
  Widget child, {
  Brightness brightness = Brightness.dark,
}) async {
  await tester.pumpWidget(MaterialApp(
    theme: brightness == Brightness.dark ? AppTheme.dark : AppTheme.light,
    home: Scaffold(body: Center(child: child)),
  ));
  // Material animates theme changes, and AppTokens.lerp holds the source values for the
  // first half of that transition. Advance past it with a fixed pump rather than
  // pumpAndSettle, which never returns while a spinner is on screen.
  await tester.pump(const Duration(seconds: 1));
}

AppColorTokens palette(Brightness b) =>
    b == Brightness.dark ? AppColorTokens.dark : AppColorTokens.light;

void main() {
  group('AppButton', () {
    testWidgets('primary renders the label and fires on tap', (tester) async {
      var taps = 0;
      await pumpKit(tester, AppButton(label: 'Готово', onPressed: () => taps++));

      expect(find.text('Готово'), findsOneWidget);
      await tester.tap(find.byType(AppButton));
      expect(taps, 1);
    });

    testWidgets('a null onPressed disables the underlying button', (tester) async {
      await pumpKit(tester, const AppButton(label: 'Нельзя'));

      final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
      expect(button.onPressed, isNull);
    });

    testWidgets('loading shows a spinner, hides the label and blocks taps',
        (tester) async {
      var taps = 0;
      await pumpKit(
        tester,
        AppButton(label: 'Сохранить', loading: true, onPressed: () => taps++),
      );

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Сохранить'), findsNothing);

      await tester.tap(find.byType(AppButton));
      expect(taps, 0, reason: 'a loading button must not fire');
    });

    testWidgets('every variant meets the minimum touch target', (tester) async {
      for (final variant in AppButtonVariant.values) {
        await pumpKit(
          tester,
          AppButton(label: 'Tap', variant: variant, onPressed: () {}),
        );
        final size = tester.getSize(find.byType(AppButton));
        expect(
          size.height,
          greaterThanOrEqualTo(AppTokens.border.minTouchTarget),
          reason: '$variant is only ${size.height} tall',
        );
      }
    });

    testWidgets('secondary is outlined and ghost has no container', (tester) async {
      await pumpKit(
        tester,
        const AppButton(
          label: 'Alt',
          variant: AppButtonVariant.secondary,
          onPressed: null,
        ),
      );
      expect(find.byType(OutlinedButton), findsOneWidget);

      await pumpKit(
        tester,
        const AppButton(
          label: 'Alt',
          variant: AppButtonVariant.ghost,
          onPressed: null,
        ),
      );
      expect(find.byType(TextButton), findsOneWidget);
    });

    testWidgets('danger recolours the primary fill', (tester) async {
      await pumpKit(tester, AppButton(label: 'Удалить', danger: true, onPressed: () {}));

      final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
      final bg = button.style!.backgroundColor!.resolve(<WidgetState>{});
      expect(bg, AppColorTokens.dark.danger);
    });

    testWidgets('an icon renders alongside the label', (tester) async {
      await pumpKit(
        tester,
        AppButton(label: 'Поделиться', icon: Icons.share, onPressed: () {}),
      );

      expect(find.byIcon(Icons.share), findsOneWidget);
      expect(find.text('Поделиться'), findsOneWidget);
    });
  });

  group('AppCard', () {
    for (final brightness in Brightness.values) {
      testWidgets('takes the active theme tokens (${brightness.name})', (tester) async {
        await pumpKit(
          tester,
          const AppCard(child: Text('body')),
          brightness: brightness,
        );

        final decorated = tester.widget<DecoratedBox>(
          find.descendant(
            of: find.byType(AppCard),
            matching: find.byType(DecoratedBox),
          ).first,
        );
        final decoration = decorated.decoration as BoxDecoration;
        final expected = brightness == Brightness.dark
            ? AppElevationTokens.dark.level1
            : AppElevationTokens.light.level1;

        expect(decoration.color, expected.surface);
        expect(decoration.boxShadow, expected.shadows);
        expect(decoration.borderRadius, AppTokens.radius.card);
      });
    }

    testWidgets('dark cards carry no shadow, light cards do', (tester) async {
      await pumpKit(tester, const AppCard(child: Text('x')));
      var decoration = (tester
              .widget<DecoratedBox>(find.descendant(
                of: find.byType(AppCard),
                matching: find.byType(DecoratedBox),
              ).first)
              .decoration as BoxDecoration);
      expect(decoration.boxShadow, isEmpty);

      await pumpKit(tester, const AppCard(child: Text('x')),
          brightness: Brightness.light);
      decoration = (tester
              .widget<DecoratedBox>(find.descendant(
                of: find.byType(AppCard),
                matching: find.byType(DecoratedBox),
              ).first)
              .decoration as BoxDecoration);
      expect(decoration.boxShadow, isNotEmpty);
    });

    testWidgets('raised uses the level-2 surface', (tester) async {
      await pumpKit(tester, const AppCard(raised: true, child: Text('x')));

      final decoration = (tester
              .widget<DecoratedBox>(find.descendant(
                of: find.byType(AppCard),
                matching: find.byType(DecoratedBox),
              ).first)
              .decoration as BoxDecoration);
      expect(decoration.color, AppElevationTokens.dark.level2.surface);
    });

    testWidgets('a tappable card fires', (tester) async {
      var taps = 0;
      await pumpKit(tester, AppCard(onTap: () => taps++, child: const Text('x')));

      await tester.tap(find.text('x'));
      expect(taps, 1);
    });
  });

  group('AppSectionHeader', () {
    testWidgets('renders the title, uppercases the overline, and shows trailing',
        (tester) async {
      await pumpKit(
        tester,
        const AppSectionHeader(
          title: 'Участники',
          overline: 'группа',
          trailing: Icon(Icons.add),
        ),
      );

      expect(find.text('Участники'), findsOneWidget);
      expect(find.text('ГРУППА'), findsOneWidget);
      expect(find.byIcon(Icons.add), findsOneWidget);
    });
  });

  group('AppChip', () {
    testWidgets('selected uses the accent fill and its on-colour', (tester) async {
      await pumpKit(tester, const AppChip(label: 'HOT', selected: true));

      final text = tester.widget<Text>(find.text('HOT'));
      expect(text.style!.color, AppColorTokens.dark.onAccentFill);
    });

    testWidgets('unselected uses primary text on the surface', (tester) async {
      await pumpKit(tester, const AppChip(label: 'HOT'));

      final text = tester.widget<Text>(find.text('HOT'));
      expect(text.style!.color, AppColorTokens.dark.textPrimary);
    });

    testWidgets('a tappable chip still meets the touch target', (tester) async {
      await pumpKit(tester, AppChip(label: 'HOT', onTap: () {}));

      expect(
        tester.getSize(find.byType(AppChip)).height,
        greaterThanOrEqualTo(AppTokens.border.minTouchTarget),
      );
    });

    testWidgets('the energy tone uses the energy roles', (tester) async {
      await pumpKit(
        tester,
        const AppChip(label: '7', tone: AppChipTone.energy, selected: true),
      );

      final text = tester.widget<Text>(find.text('7'));
      expect(text.style!.color, AppColorTokens.dark.onEnergyFill);
    });
  });

  group('AppListTile', () {
    testWidgets('a row is at least the minimum target tall at compact density',
        (tester) async {
      await pumpKit(tester, const AppListTile(title: 'Маша'));

      expect(
        tester.getSize(find.byType(AppListTile)).height,
        greaterThanOrEqualTo(AppTokens.border.minTouchTarget),
      );
    });

    testWidgets('renders subtitle, leading and trailing', (tester) async {
      await pumpKit(
        tester,
        const AppListTile(
          title: 'Маша',
          subtitle: 'Админ',
          leading: Icon(Icons.person),
          trailing: Icon(Icons.chevron_right),
        ),
      );

      expect(find.text('Маша'), findsOneWidget);
      expect(find.text('Админ'), findsOneWidget);
      expect(find.byIcon(Icons.person), findsOneWidget);
      expect(find.byIcon(Icons.chevron_right), findsOneWidget);
    });
  });

  group('AppProgressBar', () {
    testWidgets('determinate carries its value', (tester) async {
      await pumpKit(tester, const AppProgressBar(value: 0.4));

      final bar = tester.widget<LinearProgressIndicator>(
          find.byType(LinearProgressIndicator));
      expect(bar.value, 0.4);
      expect(bar.color, AppColorTokens.dark.accent);
    });

    testWidgets('indeterminate has a null value', (tester) async {
      await pumpKit(tester, const AppProgressBar());

      final bar = tester.widget<LinearProgressIndicator>(
          find.byType(LinearProgressIndicator));
      expect(bar.value, isNull);
    });

    testWidgets('tones pick different colours', (tester) async {
      await pumpKit(tester, const AppProgressBar(value: 1, tone: AppProgressTone.energy));

      final bar = tester.widget<LinearProgressIndicator>(
          find.byType(LinearProgressIndicator));
      expect(bar.color, AppColorTokens.dark.energy);
    });
  });

  group('AppStat', () {
    testWidgets('uses the display role with tabular figures and an uppercase label',
        (tester) async {
      await pumpKit(tester, const AppStat(value: '67%', label: 'точность'));

      final value = tester.widget<Text>(find.text('67%'));
      expect(value.style!.fontSize, AppTokens.text.display2.fontSize);
      expect(value.style!.fontFeatures!.map((f) => f.feature), contains('tnum'));
      expect(find.text('ТОЧНОСТЬ'), findsOneWidget);
    });

    testWidgets('compact uses the numeric body role', (tester) async {
      await pumpKit(
        tester,
        const AppStat(value: '12', label: 'серия', compact: true),
      );

      final value = tester.widget<Text>(find.text('12'));
      expect(value.style!.fontSize, AppTokens.text.numeric.fontSize);
    });

    testWidgets('the energy tone recolours the numeral', (tester) async {
      await pumpKit(
        tester,
        const AppStat(value: '3', label: 'кристаллы', tone: AppStatTone.energy),
      );

      final value = tester.widget<Text>(find.text('3'));
      expect(value.style!.color, AppColorTokens.dark.energy);
    });
  });

  group('showAppSheet', () {
    testWidgets('uses the sheet radius and level-2 surface', (tester) async {
      await tester.pumpWidget(MaterialApp(
        theme: AppTheme.dark,
        home: Scaffold(
          body: Builder(
            builder: (context) => AppButton(
              label: 'open',
              onPressed: () => showAppSheet<void>(
                context: context,
                builder: (_) => const Text('inside'),
              ),
            ),
          ),
        ),
      ));

      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      expect(find.text('inside'), findsOneWidget);

      final sheet = tester.widget<BottomSheet>(find.byType(BottomSheet));
      final shape = sheet.shape! as RoundedRectangleBorder;
      expect(shape.borderRadius, AppTokens.radius.sheet);
      expect(sheet.backgroundColor, AppElevationTokens.dark.level2.surface);
    });
  });
}
