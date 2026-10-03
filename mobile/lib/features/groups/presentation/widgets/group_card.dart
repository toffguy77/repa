import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../domain/group.dart';
import '../../../../core/theme/app_tokens.dart';

class GroupCard extends StatelessWidget {
  final GroupListItem group;
  final VoidCallback onTap;

  const GroupCard({super.key, required this.group, required this.onTap});

  String _statusText(ActiveSeason? season) {
    if (season == null) return 'Нет активного сезона';
    switch (season.status) {
      case 'VOTING':
        return season.userVoted ? 'Ждём пятницы' : 'Голосуй!';
      case 'REVEALED':
        return 'Результаты готовы';
      default:
        return 'Сезон завершён';
    }
  }

  Color _statusColor(BuildContext context, ActiveSeason? season) {
    if (season == null) return context.t.color.textSecondary;
    if (season.status == 'VOTING' && !season.userVoted) {
      return context.t.color.accent;
    }
    if (season.status == 'REVEALED') return context.t.color.success;
    return context.t.color.textSecondary;
  }

  @override
  Widget build(BuildContext context) {
    final season = group.activeSeason;
    final needsVote = season != null && season.status == 'VOTING' && !season.userVoted;

    Widget card = GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
        color: context.t.elevation.level1.surface,
        borderRadius: AppTokens.radius.card,
        border: Border.all(
          color: context.t.elevation.level1.outline,
          width: AppTokens.border.hairline,
        ),
        boxShadow: context.t.elevation.level1.shadows,
      ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    group.name,
                    style: context.ts.heading2.copyWith(fontSize: 18),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Text(
                  '${group.memberCount} чел.',
                  style: context.ts.caption,
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (season != null && season.status == 'VOTING') ...[
              ClipRRect(
                borderRadius: BorderRadius.circular(AppTokens.radius.xs),
                child: LinearProgressIndicator(
                  value: season.totalCount > 0
                      ? season.votedCount / season.totalCount
                      : 0,
                  backgroundColor: context.t.color.surface,
                  color: context.t.color.accent,
                  minHeight: 6,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                '${season.votedCount} из ${season.totalCount} проголосовали',
                style: context.ts.caption,
              ),
            ],
            const SizedBox(height: 4),
            Text(
              _statusText(season),
              style: context.ts.body.copyWith(
                color: _statusColor(context, season),
                fontWeight: FontWeight.w600,
                fontSize: 14,
              ),
            ),
          ],
        ),
      ),
    );

    if (needsVote) {
      card = card
          .animate(onPlay: (c) => c.repeat(reverse: true))
          .shimmer(
            duration: context.motion(MotionClass.emphasis),
            color: context.t.color.accent.withValues(alpha: 0.08),
          );
    }

    return card;
  }
}
