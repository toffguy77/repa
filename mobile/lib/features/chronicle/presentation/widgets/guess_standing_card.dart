import 'package:flutter/material.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../domain/chronicle.dart';

/// Who in the group predicts it best.
///
/// Three states, all of which happen in a real group: shown, withheld because the group has too little
/// history for the ranking to mean anything, and shown with some members not yet ranked.
class GuessStandingCard extends StatelessWidget {
  const GuessStandingCard({super.key, required this.standing});

  final GuessStanding standing;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Знатоки', style: context.ts.body),
          const SizedBox(height: 4),
          if (!standing.available)
            Text(
              standing.reason.isEmpty
                  ? 'Пока рано сравнивать'
                  : standing.reason,
              style: context.ts.caption,
            )
          else ...[
            Text('Кто точнее угадывает, что скажет группа',
                style: context.ts.caption),
            const SizedBox(height: 8),
            ...standing.members.map((m) => _StandingRow(entry: m)),
          ],
        ],
      ),
    );
  }
}

class _StandingRow extends StatelessWidget {
  const _StandingRow({required this.entry});

  final StandingEntry entry;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          SizedBox(
            width: 24,
            child: Text(
              entry.isRanked ? '${entry.rank}' : '—',
              style: context.ts.caption,
            ),
          ),
          Text(entry.avatarEmoji ?? '\u{1F346}'),
          const SizedBox(width: 6),
          Expanded(child: Text(entry.username, style: context.ts.body)),
          if (entry.isRanked)
            // The place, and how much history it rests on — not the accuracy figure, which the server
            // withholds on purpose: next to the published winners above it would recover individual
            // votes. The seasons count is what makes a place interpretable; a single bad week should
            // be visible as such.
            Text('за ${_plural(entry.seasonsPlayed)}', style: context.ts.caption)
          else
            Text('ещё не играл', style: context.ts.caption),
        ],
      ),
    );
  }

  static String _plural(int n) {
    final mod100 = n % 100;
    if (mod100 >= 11 && mod100 <= 14) return '$n реп';
    switch (n % 10) {
      case 1:
        return '$n репу';
      case 2:
      case 3:
      case 4:
        return '$n репы';
      default:
        return '$n реп';
    }
  }
}
