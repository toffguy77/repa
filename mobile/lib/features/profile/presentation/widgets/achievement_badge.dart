import 'package:flutter/material.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';

class AchievementBadge extends StatelessWidget {
  final String type;
  final String? earnedAt;
  final bool unlocked;

  const AchievementBadge({
    super.key,
    required this.type,
    this.earnedAt,
    this.unlocked = true,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 100,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: unlocked ? Colors.white : AppColors.surface,
        borderRadius: BorderRadius.circular(12),
        boxShadow: unlocked
            ? [
                BoxShadow(
                  color: AppColors.primary.withValues(alpha: 0.15),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ]
            : null,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            unlocked ? _achievementEmoji(type) : '\u{1F512}',
            style: const TextStyle(fontSize: 32),
          ),
          const SizedBox(height: 6),
          Text(
            _achievementName(type),
            style: AppTextStyles.caption.copyWith(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: unlocked ? AppColors.textPrimary : AppColors.textSecondary,
            ),
            textAlign: TextAlign.center,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          if (earnedAt != null && unlocked) ...[
            const SizedBox(height: 4),
            Text(
              earnedAt!,
              style: AppTextStyles.caption.copyWith(fontSize: 10),
            ),
          ],
        ],
      ),
    );
  }

  static String _achievementEmoji(String type) {
    const map = {
      'SNIPER': '\u{1F3AF}',
      'ORACLE': '\u{1F52E}',
      'TELEPATH': '\u{1F9E0}',
      'BLIND': '\u{1F648}',
      'RANDOM': '\u{1F3B2}',
      'EXPERT_OF': '\u{1F393}',
      'BEST_FRIEND': '\u{1F91D}',
      'DETECTIVE': '\u{1F575}',
      'STRANGER': '\u{1F47B}',
      'LEGEND': '\u{1F451}',
      'CHANGEABLE': '\u{1F300}',
      'MONOPOLIST': '\u{1F3C6}',
      'ENIGMA': '\u{2753}',
      'RISING': '\u{1F4C8}',
      'PIONEER': '\u{1F680}',
      'STREAK_VOTER': '\u{1F525}',
      'FIRST_VOTER': '\u{26A1}',
      'LAST_VOTER': '\u{1F422}',
      'NIGHT_OWL': '\u{1F989}',
      'ANALYST': '\u{1F4CA}',
      'MEDIA': '\u{1F4F1}',
      'CONSPIRATOR': '\u{1F92B}',
      'RECRUITER': '\u{1F4E3}',
    };
    return map[type] ?? '\u{2B50}';
  }

  static String _achievementName(String type) {
    const map = {
      'SNIPER': 'Снайпер',
      'ORACLE': 'Оракул',
      'TELEPATH': 'Телепат',
      'BLIND': 'Слепой',
      'RANDOM': 'Рандомщик',
      'EXPERT_OF': 'Эксперт',
      'BEST_FRIEND': 'Лучший друг',
      'DETECTIVE': 'Детектив',
      'STRANGER': 'Незнакомец',
      'LEGEND': 'Легенда',
      'CHANGEABLE': 'Переменчивый',
      'MONOPOLIST': 'Монополист',
      'ENIGMA': 'Загадка',
      'RISING': 'Восходящая звезда',
      'PIONEER': 'Пионер',
      'STREAK_VOTER': 'Стрик',
      'FIRST_VOTER': 'Первый голос',
      'LAST_VOTER': 'Последний голос',
      'NIGHT_OWL': 'Ночная сова',
      'ANALYST': 'Аналитик',
      'MEDIA': 'Медийная личность',
      'CONSPIRATOR': 'Заговорщик',
      'RECRUITER': 'Рекрутер',
    };
    return map[type] ?? type;
  }
}
