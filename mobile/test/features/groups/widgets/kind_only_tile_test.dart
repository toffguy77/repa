import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/groups/presentation/widgets/kind_only_tile.dart';

Future<void> pumpTile(
  WidgetTester tester, {
  required bool kindOnly,
  required bool isAdmin,
  Future<String?> Function(bool)? onChanged,
}) async {
  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.dark,
    home: Scaffold(
      body: KindOnlyTile(
        kindOnly: kindOnly,
        isAdmin: isAdmin,
        onChanged: onChanged ?? (_) async => null,
      ),
    ),
  ));
  await tester.pump(const Duration(milliseconds: 400));
}

void main() {
  group('KindOnlyTile', () {
    testWidgets('the admin can change the setting', (tester) async {
      final changes = <bool>[];
      await pumpTile(
        tester,
        kindOnly: false,
        isAdmin: true,
        onChanged: (value) async {
          changes.add(value);
          return null;
        },
      );

      final switchWidget = tester.widget<Switch>(find.byType(Switch));
      expect(switchWidget.onChanged, isNotNull, reason: 'admin switch must be enabled');

      await tester.tap(find.byType(Switch));
      await tester.pumpAndSettle();
      expect(changes, [true]);
    });

    testWidgets('a non-admin sees the setting but cannot change it',
        (tester) async {
      await pumpTile(tester, kindOnly: true, isAdmin: false);

      // Visible, so the group's questions do not look arbitrary...
      expect(find.byType(Switch), findsOneWidget);
      // ...but not changeable: a null onChanged is how Flutter disables a Switch.
      expect(tester.widget<Switch>(find.byType(Switch)).onChanged, isNull);
      expect(find.textContaining('Менять может только админ'), findsOneWidget);
    });

    testWidgets('tells the admin when the change applies', (tester) async {
      await pumpTile(tester, kindOnly: false, isAdmin: true);
      // The surprising part: the open season deliberately does not change.
      expect(find.textContaining('со следующей репы'), findsOneWidget);
    });

    testWidgets('explains what the setting does to the questions',
        (tester) async {
      await pumpTile(tester, kindOnly: false, isAdmin: true);
      expect(find.textContaining('Колкие вопросы'), findsOneWidget);
    });

    testWidgets('shows the refusal so the admin can act on it', (tester) async {
      await pumpTile(
        tester,
        kindOnly: false,
        isAdmin: true,
        onChanged: (_) async =>
            'В режиме «только добрые вопросы» эти категории недоступны',
      );

      await tester.tap(find.byType(Switch));
      await tester.pumpAndSettle();

      expect(find.textContaining('эти категории недоступны'), findsOneWidget);
    });

    testWidgets('reflects the stored value rather than local state',
        (tester) async {
      // The switch's position comes from the group, not from the tap: the server owns the value, and
      // a refused change must not leave the switch showing something that was never saved.
      await pumpTile(
        tester,
        kindOnly: false,
        isAdmin: true,
        onChanged: (_) async => 'отказано',
      );

      await tester.tap(find.byType(Switch));
      await tester.pumpAndSettle();

      expect(tester.widget<Switch>(find.byType(Switch)).value, isFalse);
    });
  });
}
