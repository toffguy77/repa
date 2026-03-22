import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';

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
                style: AppTextStyles.body.copyWith(
                  fontWeight: FontWeight.w500,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '${percentage.toStringAsFixed(0)}%',
              style: AppTextStyles.body.copyWith(
                fontWeight: FontWeight.w700,
                color: AppColors.primary,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(6),
          child: TweenAnimationBuilder<double>(
            tween: Tween(begin: 0, end: percentage / 100),
            duration: Duration(milliseconds: 600 + index * 200),
            curve: Curves.easeOutCubic,
            builder: (context, value, _) {
              return LinearProgressIndicator(
                value: value,
                backgroundColor: AppColors.surface,
                color: AppColors.primary,
                minHeight: 10,
              );
            },
          ),
        ),
      ],
    )
        .animate()
        .fadeIn(duration: 400.ms, delay: Duration(milliseconds: index * 150))
        .slideX(begin: 0.1, duration: 300.ms);
  }
}
