import 'package:flutter/material.dart';
import '../../../../core/theme/app_tokens.dart';


/// Minimum group size at which aggregated results stop identifying individual voters.
/// Mirrors `eligibility.MinDetectorMembers` on the backend.
const int kAnonymityMinMembers = 5;

/// Warns that in a very small group the percentages give away who voted how. Shown
/// *before* voting, so members can decide what to answer rather than find out after the
/// Reveal.
class SmallGroupNotice extends StatelessWidget {
  const SmallGroupNotice({super.key, required this.memberCount});

  final int memberCount;

  @override
  Widget build(BuildContext context) {
    if (memberCount >= kAnonymityMinMembers) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: context.t.color.warning.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(AppTokens.radius.md),
        border: Border.all(color: context.t.color.warning.withValues(alpha: 0.3)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(Icons.visibility_outlined,
              size: 20, color: context.t.color.warning),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Вас пока мало',
                  style: context.ts.body.copyWith(
                    fontWeight: FontWeight.w600,
                    color: context.t.color.warning,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  'В группе меньше $kAnonymityMinMembers человек — по проценту '
                  'можно догадаться, кто как проголосовал. Позови ещё людей.',
                  style: context.ts.caption,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
