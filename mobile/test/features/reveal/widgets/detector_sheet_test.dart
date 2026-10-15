import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/reveal/domain/reveal.dart';
import 'package:repa/features/reveal/presentation/widgets/detector_sheet.dart';
import 'package:repa/core/theme/app_theme.dart';

DetectorResult detector({
  bool purchased = false,
  bool available = true,
  int balance = 50,
  List<VoterProfile> voters = const [],
}) =>
    DetectorResult(
      purchased: purchased,
      voters: voters,
      crystalBalance: balance,
      available: available,
    );

Future<void> pumpSheet(WidgetTester tester, DetectorResult d,
        {int balance = 50}) =>
    tester.pumpWidget(MaterialApp(
      theme: AppTheme.dark,
      home: Scaffold(
        body: DetectorSheet(
          detector: d,
          buying: false,
          crystalBalance: balance,
          onBuy: () {},
          onGoToShop: () {},
        ),
      ),
    ));

void main() {
  group('DetectorSheet small-group gating', () {
    testWidgets('disables the action in a group under five', (tester) async {
      await pumpSheet(tester, detector(available: false));

      final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
      expect(button.onPressed, isNull,
          reason: 'the button must not be tappable when the detector is unavailable');
      expect(find.text('Детектор недоступен'), findsOneWidget);
      expect(find.textContaining('от 5 человек'), findsOneWidget);
    });

    testWidgets('offers the purchase once the group is big enough',
        (tester) async {
      await pumpSheet(tester, detector(available: true));

      expect(find.text('Детектор недоступен'), findsNothing);
      expect(find.textContaining('Показать всех'), findsOneWidget);
    });

    testWidgets('unavailability wins over an insufficient balance',
        (tester) async {
      await pumpSheet(tester, detector(available: false, balance: 0), balance: 0);

      expect(find.text('Детектор недоступен'), findsOneWidget);
      expect(find.text('Купить кристаллы'), findsNothing);
    });
  });

  ladderTests();

  group('DetectorResult', () {
    test('available defaults to true for an older backend payload', () {
      final parsed = DetectorResult.fromJson({
        'purchased': false,
        'voters': <dynamic>[],
        'crystal_balance': 10,
      });
      expect(parsed.available, isTrue);
    });

    test('parses available from the payload', () {
      final parsed = DetectorResult.fromJson({
        'purchased': false,
        'voters': <dynamic>[],
        'crystal_balance': 10,
        'available': false,
      });
      expect(parsed.available, isFalse);
    });
  });
}

void ladderTests() {
  DetectorResult ladder({
    int voterCount = 3,
    List<DetectorHint> hints = const [],
    bool hintAvailable = true,
    bool purchased = false,
    bool available = true,
    int balance = 50,
  }) =>
      DetectorResult(
        purchased: purchased,
        voters: const [],
        crystalBalance: balance,
        available: available,
        voterCount: voterCount,
        hints: hints,
        hintAvailable: hintAvailable,
      );

  Future<void> pumpLadder(
    WidgetTester tester,
    DetectorResult d, {
    int balance = 50,
    VoidCallback? onBuyHint,
  }) async {
    tester.view.physicalSize = const Size(1200, 2400);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(MaterialApp(
      theme: AppTheme.dark,
      home: Scaffold(
        body: SingleChildScrollView(
          child: DetectorSheet(
            detector: d,
            buying: false,
            crystalBalance: balance,
            onBuy: () {},
            onBuyHint: onBuyHint ?? () {},
            onGoToShop: () {},
          ),
        ),
      ),
    ));
    await tester.pump(const Duration(milliseconds: 500));
  }

  group('detector ladder', () {
    testWidgets('the free count is visible without any purchase', (tester) async {
      await pumpLadder(tester, ladder(voterCount: 4));

      expect(find.text('4'), findsOneWidget);
      expect(find.text('ПРОГОЛОСОВАЛИ ПРО ТЕБЯ'), findsOneWidget);
    });

    testWidgets('both paid rungs are offered with their prices', (tester) async {
      await pumpLadder(tester, ladder());

      expect(find.textContaining('Подсказка — 3'), findsOneWidget);
      expect(find.textContaining('Показать всех — 10'), findsOneWidget);
    });

    testWidgets('a revealed hint shows one character, never a name', (tester) async {
      await pumpLadder(
        tester,
        ladder(hints: const [DetectorHint(firstLetter: 'М', avatarEmoji: '🍆')]),
      );

      expect(find.text('М•••'), findsOneWidget);
      expect(find.text('Подсказка'), findsOneWidget);
      expect(find.textContaining('Маша'), findsNothing);
    });

    testWidgets('the hint rung disappears when nothing is left to reveal',
        (tester) async {
      await pumpLadder(tester, ladder(hintAvailable: false));

      expect(find.textContaining('Подсказка — '), findsNothing);
      expect(find.textContaining('Показать всех'), findsOneWidget);
    });

    testWidgets('tapping the hint fires the callback', (tester) async {
      var taps = 0;
      await pumpLadder(tester, ladder(), onBuyHint: () => taps++);

      await tester.tap(find.textContaining('Подсказка — 3'));
      expect(taps, 1);
    });

    testWidgets('an insufficient balance relabels the hint rung', (tester) async {
      await pumpLadder(tester, ladder(balance: 0), balance: 0);

      expect(find.textContaining('Не хватает кристаллов на подсказку'), findsOneWidget);
    });

    testWidgets('an owned list shows the voters and no rungs', (tester) async {
      await tester.pumpWidget(MaterialApp(
        theme: AppTheme.dark,
        home: Scaffold(
          body: SingleChildScrollView(
            child: DetectorSheet(
              detector: const DetectorResult(
                purchased: true,
                voters: [VoterProfile(id: 'u2', username: 'Маша')],
                crystalBalance: 50,
                voterCount: 3,
                hintAvailable: false,
              ),
              buying: false,
              crystalBalance: 50,
              onBuy: () {},
              onGoToShop: () {},
            ),
          ),
        ),
      ));
      await tester.pump(const Duration(milliseconds: 500));

      expect(find.text('Маша'), findsOneWidget);
      expect(find.textContaining('Показать всех'), findsNothing);
    });

    testWidgets('a small group shows the count but refuses every paid rung',
        (tester) async {
      await pumpLadder(tester, ladder(available: false, voterCount: 3));

      expect(find.text('3'), findsOneWidget, reason: 'a count names nobody');
      expect(find.text('Детектор недоступен'), findsOneWidget);
      expect(find.textContaining('Подсказка — '), findsNothing);
      expect(find.textContaining('Показать всех'), findsNothing);
    });
  });

  group('DetectorResult ladder defaults', () {
    test('an older backend payload yields a safe ladder', () {
      final parsed = DetectorResult.fromJson(const {
        'purchased': false,
        'voters': <dynamic>[],
        'crystal_balance': 10,
      });

      expect(parsed.voterCount, 0);
      expect(parsed.hints, isEmpty);
      expect(parsed.hintAvailable, isFalse);
      expect(parsed.hintCost, 3);
      expect(parsed.fullCost, 10);
    });

    test('a hint parses without a username field', () {
      final hint = DetectorHint.fromJson(const {
        'first_letter': 'М',
        'avatar_emoji': '🍆',
      });

      expect(hint.firstLetter, 'М');
      expect(hint.toJson().containsKey('username'), isFalse);
    });
  });
}
