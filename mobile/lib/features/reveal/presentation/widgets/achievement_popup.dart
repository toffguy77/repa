import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../domain/reveal.dart';

class AchievementPopup extends StatefulWidget {
  final List<AchievementDto> achievements;
  final VoidCallback onDismiss;

  const AchievementPopup({
    super.key,
    required this.achievements,
    required this.onDismiss,
  });

  @override
  State<AchievementPopup> createState() => _AchievementPopupState();
}

class _AchievementPopupState extends State<AchievementPopup> {
  int _currentIndex = 0;

  void _next() {
    HapticFeedback.lightImpact();
    if (_currentIndex < widget.achievements.length - 1) {
      setState(() => _currentIndex++);
    } else {
      widget.onDismiss();
    }
  }

  @override
  Widget build(BuildContext context) {
    final achievement = widget.achievements[_currentIndex];
    final isLast = _currentIndex == widget.achievements.length - 1;

    return GestureDetector(
      onTap: _next,
      child: Container(
        color: Colors.black.withValues(alpha: 0.7),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                _achievementEmoji(achievement.type),
                style: const TextStyle(fontSize: 80),
              )
                  .animate(
                    key: ValueKey(_currentIndex),
                    onPlay: (c) => HapticFeedback.mediumImpact(),
                  )
                  .scale(
                    begin: const Offset(0.3, 0.3),
                    end: const Offset(1.0, 1.0),
                    duration: 500.ms,
                    curve: Curves.elasticOut,
                  )
                  .fadeIn(duration: 300.ms),
              const SizedBox(height: 24),
              Text(
                'Новая ачивка!',
                style: AppTextStyles.headline2.copyWith(
                  color: Colors.white,
                ),
              ).animate(key: ValueKey('title_$_currentIndex'))
                  .fadeIn(duration: 400.ms, delay: 200.ms),
              const SizedBox(height: 12),
              Text(
                _achievementName(achievement.type),
                style: AppTextStyles.headline1.copyWith(
                  color: AppColors.primary,
                  fontSize: 28,
                ),
                textAlign: TextAlign.center,
              )
                  .animate(key: ValueKey('name_$_currentIndex'))
                  .fadeIn(duration: 400.ms, delay: 300.ms)
                  .slideY(begin: 0.2, duration: 400.ms),
              const SizedBox(height: 8),
              Text(
                _achievementDescription(achievement.type),
                style: AppTextStyles.body.copyWith(
                  color: Colors.white70,
                ),
                textAlign: TextAlign.center,
              ).animate(key: ValueKey('desc_$_currentIndex'))
                  .fadeIn(duration: 400.ms, delay: 400.ms),
              const SizedBox(height: 40),
              Text(
                isLast ? 'Нажми чтобы закрыть' : 'Нажми чтобы продолжить',
                style: AppTextStyles.caption.copyWith(
                  color: Colors.white54,
                ),
              ).animate().fadeIn(duration: 400.ms, delay: 600.ms),
              if (widget.achievements.length > 1) ...[
                const SizedBox(height: 12),
                Text(
                  '${_currentIndex + 1} / ${widget.achievements.length}',
                  style: AppTextStyles.caption.copyWith(
                    color: Colors.white54,
                  ),
                ),
              ],
            ],
          ),
        ),
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

  static String _achievementDescription(String type) {
    const map = {
      'SNIPER': 'Угадал топ-1 атрибут лидера',
      'ORACLE': 'Все голоса совпали с результатами',
      'TELEPATH': 'Совпал с большинством',
      'BLIND': 'Ни одного совпадения',
      'RANDOM': 'Голосовал за всех по-разному',
      'BEST_FRIEND': 'Голосовал за одного человека чаще всех',
      'DETECTIVE': 'Купил детектор',
      'STRANGER': 'Никто не голосовал за тебя',
      'LEGEND': 'Топ-1 по атрибуту 3 сезона подряд',
      'CHANGEABLE': 'Каждый сезон новый топ-атрибут',
      'MONOPOLIST': 'Получил 80%+ по одному атрибуту',
      'ENIGMA': 'Никто не набрал больше 30% по тебе',
      'RISING': 'Самый большой рост процента',
      'PIONEER': 'Первый участник группы',
      'STREAK_VOTER': 'Голосовал 3+ сезонов подряд',
      'FIRST_VOTER': 'Первым проголосовал в сезоне',
      'LAST_VOTER': 'Последним проголосовал',
      'NIGHT_OWL': 'Голосовал после полуночи',
      'ANALYST': 'Проголосовал за каждого участника',
      'MEDIA': 'Поделился карточкой',
      'CONSPIRATOR': 'Все голоса за одного человека',
      'RECRUITER': 'Привёл 3+ человек в группу',
    };
    return map[type] ?? '';
  }
}
