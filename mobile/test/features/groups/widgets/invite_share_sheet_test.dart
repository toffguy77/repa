import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/groups/presentation/join_group_screen.dart';
import 'package:repa/features/groups/presentation/widgets/invite_share_sheet.dart';

Future<void> pumpSheet(WidgetTester tester, String code) async {
  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.dark,
    home: Scaffold(body: InviteShareSheet(inviteCode: code)),
  ));
  await tester.pump(const Duration(milliseconds: 500));
}

void main() {
  group('formatInviteCode', () {
    test('groups a six-character code into two triples', () {
      expect(formatInviteCode('AB2CD3'), 'AB2 CD3');
    });

    test('leaves a legacy code alone rather than mis-grouping it', () {
      const legacy = '0f1d4e6a-6b3c-4a1e-9f2e-123456789abc';
      expect(formatInviteCode(legacy), legacy);
    });

    test('matches the backend URL form', () {
      expect(inviteUrlFor('AB2CD3'), 'https://repa.app/join/AB2CD3');
    });
  });

  group('InviteShareSheet', () {
    testWidgets('shows the grouped code and the full link', (tester) async {
      await pumpSheet(tester, 'AB2CD3');

      expect(find.text('AB2 CD3'), findsOneWidget);
      expect(find.text('КОД ГРУППЫ'), findsOneWidget);
      expect(find.text('https://repa.app/join/AB2CD3'), findsOneWidget);
      expect(find.textContaining('продиктуй код'), findsOneWidget);
    });

    testWidgets('copying the code copies the stored form, not the grouped one',
        (tester) async {
      final copied = <String>[];
      tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        (call) async {
          if (call.method == 'Clipboard.setData') {
            copied.add((call.arguments as Map)['text'] as String);
          }
          return null;
        },
      );
      addTearDown(() => tester.binding.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, null));

      await pumpSheet(tester, 'AB2CD3');
      await tester.tap(find.text('AB2 CD3'));
      await tester.pump();

      expect(copied, ['AB2CD3'],
          reason: 'the gap is cosmetic and must not be copied');
    });

    testWidgets('copying the link copies the URL', (tester) async {
      final copied = <String>[];
      tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        (call) async {
          if (call.method == 'Clipboard.setData') {
            copied.add((call.arguments as Map)['text'] as String);
          }
          return null;
        },
      );
      addTearDown(() => tester.binding.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, null));

      await pumpSheet(tester, 'AB2CD3');
      await tester.tap(find.byIcon(Icons.copy));
      await tester.pump();

      expect(copied, ['https://repa.app/join/AB2CD3']);
    });

    testWidgets('a legacy code still renders', (tester) async {
      await pumpSheet(tester, '0f1d4e6a-6b3c-4a1e-9f2e-123456789abc');

      expect(find.textContaining('0f1d4e6a'), findsWidgets);
    });
  });

  extractorTests();

  group('UpperCaseFormatter', () {
    final formatter = UpperCaseFormatter();

    TextEditingValue apply(String text) => formatter.formatEditUpdate(
          TextEditingValue.empty,
          TextEditingValue(text: text),
        );

    test('upper-cases a typed code', () {
      expect(apply('ab2cd3').text, 'AB2CD3');
    });

    test('leaves a pasted URL untouched', () {
      const url = 'https://repa.app/join/ab2cd3';
      expect(apply(url).text, url,
          reason: 'upper-casing a host or path can break the link');
    });

    test('is a no-op on empty input', () {
      expect(apply('').text, '');
    });
  });
}

void extractorTests() {
  group('extractInviteCode', () {
    test('passes a bare code through', () {
      expect(extractInviteCode('AB2CD3'), 'AB2CD3');
    });

    test('pulls the code out of a full link', () {
      expect(extractInviteCode('https://repa.app/join/AB2CD3'), 'AB2CD3');
    });

    test('drops tracking parameters a share sheet appended', () {
      expect(
        extractInviteCode('https://repa.app/join/AB2CD3?utm_source=tg'),
        'AB2CD3',
      );
    });

    test('drops a fragment', () {
      expect(extractInviteCode('https://repa.app/join/AB2CD3#x'), 'AB2CD3');
    });

    test('trims surrounding whitespace', () {
      expect(extractInviteCode('  https://repa.app/join/AB2CD3  '), 'AB2CD3');
    });

    test('leaves case and separators to the server', () {
      // The client only has to produce a valid path segment.
      expect(extractInviteCode('ab2 cd3'), 'ab2 cd3');
    });

    test('keeps a legacy UUID intact', () {
      const legacy = '0f1d4e6a-6b3c-4a1e-9f2e-123456789abc';
      expect(extractInviteCode('https://repa.app/join/$legacy'), legacy);
    });
  });
}
