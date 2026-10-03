import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../domain/question_candidate.dart';
import '../../../../core/theme/app_tokens.dart';

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
        duration: context.motion(MotionClass.surface),
        curve: Curves.easeOutCubic,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: context.t.elevation.level1.surface,
          borderRadius: BorderRadius.circular(AppTokens.radius.lg),
          border: Border.all(
            color: selected ? context.t.color.accent : Colors.transparent,
            width: 2,
          ),
          boxShadow: [
            BoxShadow(
              color: selected
                  ? context.t.color.accent.withValues(alpha: 0.15)
                  : context.t.color.scrim.withValues(alpha: 0.05),
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
                style: context.ts.body.copyWith(
                  color: disabled && !selected
                      ? context.t.color.textSecondary
                      : context.t.color.textPrimary,
                ),
              ),
            ),
            if (selected)
              Container(
                width: 28,
                height: 28,
                decoration: BoxDecoration(
                  color: context.t.color.accentFill,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.check,
                  color: context.t.color.onAccentFill,
                  size: 18,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
