import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';

const _categoryEmojis = {
  'HOT': '\u{1F525}',
  'FUNNY': '\u{1F602}',
  'SECRETS': '\u{1F92B}',
  'SKILLS': '\u{1F3AF}',
  'ROMANCE': '\u{1F496}',
  'STUDY': '\u{1F4DA}',
};

class QuestionCard extends StatelessWidget {
  final String text;
  final String category;

  const QuestionCard({
    super.key,
    required this.text,
    required this.category,
  });

  @override
  Widget build(BuildContext context) {
    final emoji = _categoryEmojis[category] ?? '\u{2753}';

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.1),
            blurRadius: 20,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(emoji, style: const TextStyle(fontSize: 40)),
          const SizedBox(height: 16),
          Text(
            text,
            style: AppTextStyles.headline2.copyWith(
              fontSize: 20,
              height: 1.3,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    )
        .animate()
        .slideX(begin: 0.3, duration: 300.ms, curve: Curves.easeOut)
        .fadeIn(duration: 300.ms);
  }
}
