import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/core/theme/app_tokens.dart';
import 'package:repa/core/widgets/empty_state_widget.dart';
import 'package:repa/core/widgets/error_state_widget.dart';
import 'package:repa/core/widgets/kit/kit.dart';
import 'package:repa/core/widgets/reveal_countdown_widget.dart';
import 'package:repa/core/widgets/skeleton_loader.dart';
import 'package:repa/features/groups/domain/group.dart';
import 'package:repa/features/groups/presentation/widgets/pending_reveal_notice.dart';
import 'package:repa/features/groups/presentation/widgets/small_group_notice.dart';
import 'package:repa/features/reveal/domain/reveal.dart';
import 'package:repa/features/reveal/presentation/widgets/detector_sheet.dart';

/// Every widget the design system is responsible for, by name.
Map<String, Widget> surfaces() => {
      'EmptyStateWidget': const EmptyStateWidget(
        emoji: '🍑',
        title: 'Пока нет групп',
        subtitle: 'Создай первую',
        buttonText: 'Создать',
      ),
      'ErrorStateWidget': const ErrorStateWidget(message: 'boom'),
      'SkeletonLoader': const SkeletonLoader(height: 20),
      'GroupCardSkeleton': const GroupCardSkeleton(),
      'MemberCardSkeleton': const MemberCardSkeleton(),
      'VotingSessionSkeleton': const VotingSessionSkeleton(),
      'RevealCountdownWidget': const RevealCountdownWidget(
        revealAt: '2030-01-04T17:00:00Z',
      ),
      'SmallGroupNotice': const SmallGroupNotice(memberCount: 3),
      'PendingRevealNotice': PendingRevealNotice(
        season: ActiveSeason.fromJson(const {
          'id': 's1',
          'status': 'VOTING',
          'reveal_at': '2030-01-04T17:00:00Z',
          'voted_count': 1,
          'total_count': 6,
          'user_voted': true,
          'reveal_state': 'WAITING_FOR_VOTERS',
          'voters_needed': 2,
        }),
      ),
      'DetectorSheet': DetectorSheet(
        detector: const DetectorResult(
          purchased: false,
          voters: [],
          crystalBalance: 50,
          available: false,
        ),
        buying: false,
        crystalBalance: 50,
        onBuy: () {},
        onGoToShop: () {},
      ),
      'AppButton': const AppButton(label: 'Готово'),
      'AppCard': const AppCard(child: Text('card')),
      'AppChip': const AppChip(label: 'HOT'),
      'AppListTile': const AppListTile(title: 'Маша', subtitle: 'Админ'),
      'AppProgressBar': const AppProgressBar(value: 0.5),
      'AppStat': const AppStat(value: '67%', label: 'точность'),
      'AppSectionHeader': const AppSectionHeader(title: 'Участники'),
    };

void main() {
  group('every design-system surface renders under both themes', () {
    for (final brightness in Brightness.values) {
      for (final entry in surfaces().entries) {
        testWidgets('${entry.key} (${brightness.name})', (tester) async {
          // A generous surface keeps the check about theming rather than overflow.
          tester.view.physicalSize = const Size(1200, 2400);
          tester.view.devicePixelRatio = 1.0;
          addTearDown(tester.view.reset);

          await tester.pumpWidget(MaterialApp(
            theme: brightness == Brightness.dark ? AppTheme.dark : AppTheme.light,
            home: Scaffold(body: SingleChildScrollView(child: entry.value)),
          ));
          await tester.pump(const Duration(milliseconds: 500));

          expect(tester.takeException(), isNull,
              reason: '${entry.key} threw under the ${brightness.name} theme');
        });
      }
    }
  });

  group('text never lands on a surface it cannot be read on', () {
    testWidgets('resolved text roles use the palette text colours', (tester) async {
      for (final brightness in Brightness.values) {
        late BuildContext ctx;
        await tester.pumpWidget(MaterialApp(
          theme: brightness == Brightness.dark ? AppTheme.dark : AppTheme.light,
          home: Builder(builder: (context) {
            ctx = context;
            return const SizedBox();
          }),
        ));

        final palette = ctx.t.color;
        // Each resolved role must come from the palette, never from a literal.
        expect(ctx.ts.body.color, palette.textPrimary);
        expect(ctx.ts.heading1.color, palette.textPrimary);
        expect(ctx.ts.bodySecondary.color, palette.textSecondary);
        expect(ctx.ts.caption.color, palette.textSecondary);
        expect(ctx.ts.label.color, palette.textSecondary);
        expect(ctx.ts.button.color, palette.onAccentFill);
        expect(ctx.ts.accent.color, palette.accent);
      }
    });
  });
}
