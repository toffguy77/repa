import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_tokens.dart';

class AttributeBar extends StatelessWidget {
  final String questionText;
  final double percentage;
  final int index;

  const AttributeBar({
    super.key,
    required this.questionText,
    required this.percentage,
    required this.index,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: Text(
                questionText,
                style: context.ts.body.copyWith(
                  fontWeight: FontWeight.w500,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '${percentage.toStringAsFixed(0)}%',
              style: context.ts.body.copyWith(
                fontWeight: FontWeight.w700,
                color: context.t.color.accent,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(AppTokens.radius.xs),
          child: TweenAnimationBuilder<double>(
            tween: Tween(begin: 0, end: percentage / 100),
            duration: context.motion(MotionClass.emphasis) + AppTokens.motion.stagger(index * 2),
            curve: Curves.easeOutCubic,
            builder: (context, value, _) {
              return LinearProgressIndicator(
                value: value,
                backgroundColor: context.t.color.surface,
                color: context.t.color.accent,
                minHeight: 10,
              );
            },
          ),
        ),
      ],
    )
        .animate()
        .fadeIn(duration: context.motion(MotionClass.emphasis), delay: Duration(milliseconds: index * 150))
        .slideX(begin: 0.1, duration: context.motion(MotionClass.surface));
  }
}
