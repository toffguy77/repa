import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../domain/question_candidate.dart';

const _categoryEmojis = {
  'HOT': '\u{1F525}',
  'FUNNY': '\u{1F602}',
  'SECRETS': '\u{1F92B}',
  'SKILLS': '\u{1F3AF}',
  'ROMANCE': '\u{1F48C}',
  'STUDY': '\u{1F4DA}',
};

class QuestionCandidateCard extends StatelessWidget {
  final QuestionCandidate candidate;
  final bool selected;
  final bool disabled;
  final VoidCallback? onTap;

  const QuestionCandidateCard({
    super.key,
    required this.candidate,
    this.selected = false,
    this.disabled = false,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final emoji = _categoryEmojis[candidate.category] ?? '\u{2753}';

    return GestureDetector(
      onTap: disabled
          ? null
          : () {
              HapticFeedback.mediumImpact();
              onTap?.call();
            },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        curve: Curves.easeOutCubic,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: selected ? AppColors.primary : Colors.transparent,
            width: 2,
          ),
          boxShadow: [
            BoxShadow(
              color: selected
                  ? AppColors.primary.withValues(alpha: 0.15)
                  : Colors.black.withValues(alpha: 0.05),
              blurRadius: selected ? 16 : 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            Text(emoji, style: const TextStyle(fontSize: 32)),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                candidate.text,
                style: AppTextStyles.body.copyWith(
                  color: disabled && !selected
                      ? AppColors.textSecondary
                      : AppColors.textPrimary,
                ),
              ),
            ),
            if (selected)
              Container(
                width: 28,
                height: 28,
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.check, color: Colors.white, size: 18),
              ),
          ],
        ),
      ),
    );
  }
}
