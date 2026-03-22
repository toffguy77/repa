import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../domain/group.dart';

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

  Color _statusColor(ActiveSeason? season) {
    if (season == null) return AppColors.textSecondary;
    if (season.status == 'VOTING' && !season.userVoted) {
      return AppColors.primary;
    }
    if (season.status == 'REVEALED') return AppColors.success;
    return AppColors.textSecondary;
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
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    group.name,
                    style: AppTextStyles.headline2.copyWith(fontSize: 18),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Text(
                  '${group.memberCount} чел.',
                  style: AppTextStyles.caption,
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (season != null && season.status == 'VOTING') ...[
              ClipRRect(
                borderRadius: BorderRadius.circular(4),
                child: LinearProgressIndicator(
                  value: season.totalCount > 0
                      ? season.votedCount / season.totalCount
                      : 0,
                  backgroundColor: AppColors.surface,
                  color: AppColors.primary,
                  minHeight: 6,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                '${season.votedCount} из ${season.totalCount} проголосовали',
                style: AppTextStyles.caption,
              ),
            ],
            const SizedBox(height: 4),
            Text(
              _statusText(season),
              style: AppTextStyles.body.copyWith(
                color: _statusColor(season),
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
            duration: 2000.ms,
            color: AppColors.primary.withValues(alpha: 0.08),
          );
    }

    return card;
  }
}
