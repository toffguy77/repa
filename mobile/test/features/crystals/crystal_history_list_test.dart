import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/core/theme/app_tokens.dart';
import 'package:repa/features/crystals/domain/crystals.dart';
import 'package:repa/features/crystals/presentation/widgets/crystal_history_list.dart';

CrystalHistoryEntry grant(String reason, int delta) => CrystalHistoryEntry(
      delta: delta,
      type: 'BONUS',
      reason: reason,
      createdAt: '2026-10-02T10:00:00Z',
      isGrant: true,
    );

CrystalHistoryEntry purchase(int delta) => CrystalHistoryEntry(
      delta: delta,
      type: 'PURCHASE',
      reason: 'Purchase: $delta crystals',
      createdAt: '2026-10-02T11:00:00Z',
      isGrant: false,
    );

Future<void> pumpHistory(WidgetTester tester, List<CrystalHistoryEntry> entries) async {
  tester.view.physicalSize = const Size(1200, 2400);
  tester.view.devicePixelRatio = 1.0;
  addTearDown(tester.view.reset);

  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.dark,
    home: Scaffold(
      body: SingleChildScrollView(child: CrystalHistoryList(entries: entries)),
    ),
  ));
  await tester.pump(const Duration(milliseconds: 500));
}

void main() {
  group('CrystalHistoryList', () {
    testWidgets('shows nothing when there is no history', (tester) async {
      await pumpHistory(tester, const []);

      expect(find.text('Откуда кристаллы'), findsNothing);
    });

    testWidgets('a grant and a purchase render differently', (tester) async {
      await pumpHistory(tester, [
        grant('Подарок за регистрацию', 10),
        purchase(30),
      ]);

      expect(find.text('Откуда кристаллы'), findsOneWidget);
      expect(find.text('Подарок за регистрацию'), findsOneWidget);
      expect(find.text('Бесплатно'), findsOneWidget);

      // A grant is marked with the gift icon, a purchase with the bag.
      expect(find.byIcon(Icons.card_giftcard), findsOneWidget);
      expect(find.byIcon(Icons.shopping_bag_outlined), findsOneWidget);
    });

    testWidgets('a grant is tinted with the energy role', (tester) async {
      await pumpHistory(tester, [grant('Друг присоединился и проголосовал', 5)]);

      final icon = tester.widget<Icon>(find.byIcon(Icons.card_giftcard));
      expect(icon.color, AppColorTokens.dark.energy);
    });

    testWidgets('amounts are signed and use tabular figures', (tester) async {
      await pumpHistory(tester, [grant('Серия голосований', 5)]);

      final amount = tester.widget<Text>(find.text('+5'));
      expect(amount.style!.fontFeatures!.map((f) => f.feature), contains('tnum'));
      expect(amount.style!.color, AppColorTokens.dark.success);
    });

    testWidgets('a reasonless entry still names itself', (tester) async {
      await pumpHistory(tester, [
        const CrystalHistoryEntry(
          delta: 10,
          type: 'BONUS',
          reason: '',
          createdAt: '2026-10-02T10:00:00Z',
          isGrant: true,
        ),
      ]);

      expect(find.text('Подарок'), findsOneWidget);
    });
  });

  group('CrystalHistoryEntry', () {
    test('isGrant defaults to false for an older payload', () {
      final parsed = CrystalHistoryEntry.fromJson(const {
        'delta': 30,
        'type': 'PURCHASE',
        'reason': 'Purchase',
        'created_at': '2026-10-02T10:00:00Z',
      });

      expect(parsed.isGrant, isFalse);
    });

    test('parses a grant', () {
      final parsed = CrystalHistoryEntry.fromJson(const {
        'delta': 10,
        'type': 'BONUS',
        'reason': 'Подарок за регистрацию',
        'created_at': '2026-10-02T10:00:00Z',
        'is_grant': true,
      });

      expect(parsed.isGrant, isTrue);
      expect(parsed.delta, 10);
    });
  });
}
