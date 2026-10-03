import 'package:flutter/material.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../../../core/widgets/reveal_countdown_widget.dart';
import '../../domain/reveal.dart';

/// What is known before the Reveal: a number, maybe one emoji, and a countdown.
///
/// This is what the Tuesday and Thursday pushes have been promising. Before it existed, the tap
/// landed on a progress bar — which does not merely fail to retain, it teaches people to ignore the
/// next notification.
///
/// It deliberately shows no name, no attribute and no percentage. The *what* stays sealed until
/// Friday; see docs/features/reveal.md → Anticipation.
class AnticipationPanel extends StatelessWidget {
  const AnticipationPanel({super.key, required this.state});

  final AnticipationState? state;

  @override
  Widget build(BuildContext context) {
    final s = state;
    if (s == null) return const SizedBox.shrink();

    final t = context.t;
    final hasVotes = s.votersAboutMe > 0;

    return AppCard(
      child: Column(
        children: [
          AppStat(
            value: '${s.votersAboutMe}',
            label: 'ответили про тебя',
            tone: hasVotes ? AppStatTone.accent : AppStatTone.primary,
          ),
          if (!hasVotes) ...[
            SizedBox(height: AppTokens.space.sm),
            Text(
              'Пока тихо. Как только кто-то ответит — узнаешь.',
              style: context.ts.caption,
              textAlign: TextAlign.center,
            ),
          ],

          // The teaser: one emoji, from Thursday. Enough to think about all evening, not enough to
          // know anything.
          if (s.teaserEmoji.isNotEmpty) ...[
            SizedBox(height: AppTokens.space.lg),
            Divider(color: t.color.border, height: 1),
            SizedBox(height: AppTokens.space.lg),
            Text(s.teaserEmoji, style: const TextStyle(fontSize: 48)),
            SizedBox(height: AppTokens.space.sm),
            Text(
              'Один твой атрибут уже почти определился',
              style: context.ts.caption,
              textAlign: TextAlign.center,
            ),
          ],

          if (s.revealAt.isNotEmpty) ...[
            SizedBox(height: AppTokens.space.lg),
            RevealCountdownWidget(revealAt: s.revealAt),
          ],
        ],
      ),
    );
  }
}
