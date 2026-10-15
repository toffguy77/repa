import 'package:flutter/material.dart';

import '../../domain/group.dart';
import '../../../../core/theme/app_tokens.dart';

/// Explains what a pending Reveal is waiting for. Shown instead of leaving the member
/// to guess why Friday came and went: the backend tells us exactly how many more people
/// need to join or vote (see `reveal_state` in docs/features/groups.md).
class PendingRevealNotice extends StatelessWidget {
  const PendingRevealNotice({super.key, required this.season});

  final ActiveSeason season;

  @override
  Widget build(BuildContext context) {
    // Nothing to explain while the Reveal is simply on schedule.
    if (season.revealState == SeasonRevealState.scheduled ||
        season.revealState == SeasonRevealState.revealed) {
      return const SizedBox.shrink();
    }

    final waitingForPeople = season.isWaitingForPeople;
    final color = waitingForPeople ? context.t.color.accent : context.t.color.textSecondary;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(AppTokens.radius.md),
        border: Border.all(color: color.withValues(alpha: 0.25)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            waitingForPeople ? Icons.group_add_outlined : Icons.schedule,
            size: 20,
            color: color,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  season.revealHeadline,
                  style: context.ts.body.copyWith(
                    fontWeight: FontWeight.w600,
                    color: color,
                  ),
                ),
                const SizedBox(height: 2),
                Text(season.revealExplanation, style: context.ts.caption),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
